package biz

import (
	"data_collector/config"
	"data_collector/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/goquery"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Consumer struct {
	redis      *redis.Client
	DB         *sql.DB
	queueName  string
	workersNum int
}

func NewConsumer(config *config.Config, queueName string, workersNum int) *Consumer {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s",
		config.MySQL.Username, config.MySQL.Password, config.MySQL.Host, config.MySQL.Database))
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	return &Consumer{
		redis: redis.NewClient(&redis.Options{
			Addr:     config.Redis.Host,
			Password: config.Redis.Password,
			DB:       0,
		}),
		DB:         db,
		queueName:  queueName,
		workersNum: workersNum,
	}
}

func (c *Consumer) Consume() {
	log.Printf("start consumer...queueName: %s, workers: %s", c.queueName, c.workersNum)
	var wg sync.WaitGroup
	wg.Add(c.workersNum)
	for i := 0; i < c.workersNum; i++ {
		go func() {
			defer wg.Done()
			for {
				taskJSON, err := c.redis.RPop(c.queueName).Result()
				if err != nil {
					log.Printf("Error consuming task: %v", err)
					continue
				}
				var task model.Task
				if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
					log.Printf("Error unmarshalling task: %v", err)
					continue
				}
				c.ProcessTask(task)
			}
		}()
	}
	wg.Wait()
}

func (c *Consumer) saveToMySQL(infos []model.BookingInfo) {
	stmt, err := c.DB.Prepare("INSERT INTO airbnb_bookings(hotel_name, star, price, price_before_taxes, check_in_date, check_out_date, guests) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		log.Fatalf("Failed to prepare SQL statement: %v", err)
	}
	defer stmt.Close()
	for _, info := range infos {
		_, err := stmt.Exec(info.HotelName, info.Star, info.Price, info.PriceBeforeTaxes, info.CheckInDate, info.CheckOutDate, info.Guests)
		if err != nil {
			log.Printf("Error inserting data: %v", err)
		}
	}
}

func (c *Consumer) ProcessTask(task model.Task) {
	//headers := make(map[string]string, 0)
	//headers["user-agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0"
	//task := model.Task{
	//	Name:    "Europe travel",
	//	Url:     "https://www.airbnb.com/s/Europe/homes?tab_id=home_tab&refinement_path s%5B%5D=%2Fhomes&flexible_trip_lengths%5B%5D=one_week&monthly_start_da te=2024-12-01&monthly_length=3&monthly_end_date=2025-03-01&price_filte r_input_type=0&channel=EXPLORE&place_id=ChIJhdqtz4aI7UYRefD8s-aZ73I&da te_picker_type=calendar&source=structured_search_input_header&search_t ype=filter_change",
	//	Headers: headers,
	//}
	client := &http.Client{}
	url := task.Url
	roomHrefs := make([]string, 0)
	// 获取所有房源链接
	for {
		// 最后一页
		if url == "" {
			break
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
			return
		}
		for key, value := range task.Headers {
			req.Header.Set(key, value)
		}
		// 发送HTTP GET请求
		response, err := client.Do(req)
		if err != nil {
			log.Fatalf("请求失败:", err)
			return
		}
		defer response.Body.Close()

		// 检查响应状态码
		if response.StatusCode != http.StatusOK {
			log.Fatalf("请求失败，状态码: %d\n", response.StatusCode)
			return
		}

		// 使用goquery解析HTML
		doc, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			log.Fatalf("解析HTML失败: %v", err)
		}

		// 查找所有房源链接元素
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists && strings.Contains(href, "/rooms/") {
				roomHrefs = append(roomHrefs, href)
			}
			// 获取下一页
			ariaLabel, exists := s.Attr("aria-label")
			if exists && ariaLabel == "下一个" {
				url = href
			}
		})
	}
	// 获取每个链接下房源信息
	bookingInfos := make([]model.BookingInfo, 0)
	for _, roomHref := range roomHrefs {
		req, err := http.NewRequest("GET", roomHref, nil)
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
			return
		}
		for key, value := range task.Headers {
			req.Header.Set(key, value)
		}
		// 发送HTTP GET请求
		response, err := client.Do(req)
		if err != nil {
			log.Fatalf("请求房源信息失败:", err)
			return
		}
		defer response.Body.Close()

		// 检查响应状态码
		if response.StatusCode != http.StatusOK {
			log.Fatalf("请求房源信息失败，状态码: %d\n", response.StatusCode)
			return
		}

		// 使用goquery解析HTML
		doc, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			log.Fatalf("解析房源信息HTML失败: %v", err)
		}
		// 查找所有房源链接元素
		bookingInfo := model.BookingInfo{}
		// hotelName
		doc.Find("div.toieuka.atm_c8_2x1prs.atm_g3_1jbyh58.atm_fr_11a07z3.atm_cs_10d11i2.atm_c8_sz6sci__oggzyc.atm_g3_17zsb9a__oggzyc.atm_fr_kzfbxz__oggzyc.dir.dir-ltr h2").Each(func(i int, s *goquery.Selection) {
			bookingInfo.HotelName = s.Text()
		})
		// star
		doc.Find("div.r1lutz1s.atm_c8_o7aogt.atm_c8_l52nlx__oggzyc").Each(func(i int, s *goquery.Selection) {
			if s.Text() == "新上线" {
				bookingInfo.Star = -1
			} else {
				star, err := strconv.ParseFloat(s.Text(), 64)
				if err != nil {
					log.Fatalf("ParseFloat err: %+v", err)
				}
				bookingInfo.Star = star
			}
		})
		doc.Find("div._m495dq").Each(func(i int, s *goquery.Selection) {
			// price
			totalPriceText := s.Find("span._j1kt73").Text()
			totalPrice, err := convertToNumber(totalPriceText)
			if err != nil {
				log.Fatalf("转换price失败 err: %+v", err)
			} else {
				bookingInfo.Price = float64(totalPrice)
			}
			// price_before_taxes
			priceText := s.Find("span._1k4xcdh").First().Text()
			priceBeforeTaxes, err := convertToNumber(priceText)
			if err != nil {
				log.Printf("转换price_before_taxes失败: %v", err)
			} else {
				bookingInfo.PriceBeforeTaxes = float64(priceBeforeTaxes)
			}
		})
		// check_in_date
		checkIn := doc.Find("[data-testid='change-dates-checkIn']").Text()
		checkIn = strings.TrimSpace(checkIn)
		bookingInfo.CheckInDate = strings.Replace(checkIn, "/", "-", -1)
		// check_out_date
		checkOut := doc.Find("[data-testid='change-dates-checkOut']").Text()
		checkOut = strings.TrimSpace(checkOut)
		bookingInfo.CheckOutDate = strings.Replace(checkOut, "/", "-", -1)
		// guests
		guests := doc.Find("span._j1kt73").Text()
		guestsNum, err := strconv.ParseInt(strings.Fields(guests)[0], 10, 64)
		if err != nil {
			log.Printf("转换guests失败: %v", err)
		}
		bookingInfo.Guests = guestsNum
		bookingInfos = append(bookingInfos, bookingInfo)
	}
	c.saveToMySQL(bookingInfos)
}

func convertToNumber(text string) (int, error) {
	// 去除非数字字符
	re := regexp.MustCompile(`[^\d]`)
	cleanText := re.ReplaceAllString(text, "")

	return strconv.Atoi(cleanText)
}
