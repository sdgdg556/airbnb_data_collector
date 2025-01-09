package model

// Task 表示一个抓取任务
type Task struct {
	Name    string            `json:"name"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

// BookingInfo 表示预订信息，根据实际情况修改
type BookingInfo struct {
	ID          int    `json:"id"`
	TaskName    string `json:"task_name"`
	Property    string `json:"property"`
	Price       string `json:"price"`
	Description string `json:"description"`
}
