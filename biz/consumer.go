package biz

import (
	"data_collector/config"
	"data_collector/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"net/http"
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
		DB: db,
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
				taskJSON, err := c.redis.RPop("airbnb_tasks").Result()
				if err != nil {
					log.Printf("Error consuming task: %v", err)
					continue
				}
				var task model.Task
				if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
					log.Printf("Error unmarshalling task: %v", err)
					continue
				}
				c.processTask(task)
			}
		}()
	}
	wg.Wait()
}

func (c *Consumer) saveToMySQL(infos []model.BookingInfo) {
	stmt, err := c.DB.Prepare("INSERT INTO booking_infos(task_name, property, price, description) VALUES(?,?,?,?)")
	if err != nil {
		log.Fatalf("Failed to prepare SQL statement: %v", err)
	}
	defer stmt.Close()
	for _, info := range infos {
		_, err := stmt.Exec(info.TaskName, info.Property, info.Price, info.Description)
		if err != nil {
			log.Printf("Error inserting data: %v", err)
		}
	}
}

func (c *Consumer) processTask(task model.Task) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", task.Url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}
	for key, value := range task.Headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Printf("Error parsing HTML: %v", err)
		return
	}
	// 添加解析 HTML 并提取预订信息的逻辑
	// 遍历节点，找到预订信息元素，提取信息并存储到 BookingInfo 结构体中
	var bookingInfos []model.BookingInfo
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "listing" {
					// 提取信息
					bookingInfo := BookingInfo{
						TaskName:    task.Name,
						Property:    ,
						Price:       ,
						Description: ,
					}
					bookingInfos = append(bookingInfos, bookingInfo)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	c.saveToMySQL(bookingInfos)
}
