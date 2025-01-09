package model

// Task 表示一个抓取任务
type Task struct {
	Name    string            `json:"name"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

// BookingInfo

type BookingInfo struct {
	ID               int64   `json:"id"`
	HotelName        string  `json:"hotel_name"`
	Star             float64 `json:"star"`
	Price            float64 `json:"price"`
	PriceBeforeTaxes float64 `json:"price_before_taxes"`
	CheckInDate      string  `json:"check_in_date"`
	CheckOutDate     string  `json:"check_out_date"`
	Guests           int64   `json:"guests"`
}
