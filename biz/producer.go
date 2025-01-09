package biz

import (
	"data_collector/config"
	"data_collector/model"
	"encoding/json"
	"github.com/go-redis/redis/v7"
	"io/ioutil"
	"log"
)

type Producer struct {
	redis     *redis.Client
	queueName string
}

func NewProducer(config *config.Config, queueName string) *Producer {
	return &Producer{
		redis: redis.NewClient(&redis.Options{
			Addr:     config.Redis.Host,
			Password: config.Redis.Password,
			DB:       0, // use default DB
		}),
		queueName: queueName,
	}
}

func (p *Producer) Produce(dataFile string) {
	log.Printf("start produce...queueName: %s, dataFile: %s", p.queueName, dataFile)
	file, err := ioutil.ReadFile(dataFile)
	if err != nil {
		log.Fatalf("Failed to read tasks file: %v", err)
	}
	var tasks []model.Task
	if err := json.Unmarshal(file, &tasks); err != nil {
		log.Fatalf("Failed to unmarshal tasks: %v", err)
	}
	for _, task := range tasks {
		taskJSON, err := json.Marshal(task)
		if err != nil {
			log.Printf("Error marshalling task: %v", err)
			continue
		}
		// 保证task唯一性
		if _, err := p.redis.SAdd("airbnb_task_ids", taskJSON).Result(); err == nil {
			p.redis.LPush(p.queueName, taskJSON)
		}
	}
}
