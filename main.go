package main

import (
	"data_collector/biz"
	"data_collector/config"
	"log"
	"os"
	"strconv"
	"strings"
	//
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	configConent, err := config.GetConfig("../config/config.yaml")
	if err != nil {
		log.Fatalf("load config err: %+v, config_path: ../config/config.yaml", err)
	}
	if len(os.Args) < 6 {
		log.Fatal("Usage: airbnb-cli start [consumer|producer] [--workers=N] --queue=<your-queue-server> --data tasks.json")
	}
	for i := 0; i < len(os.Args); i++ {
		log.Printf("os_args[%d]: %s", i, os.Args[i])
	}
	action := os.Args[2]
	queueName := ""
	workers := 0
	dataFile := ""

	for i := 3; i < len(os.Args); i++ {
		switch {
		case strings.HasPrefix(os.Args[i], "--queue="):
			queueName = strings.TrimPrefix(os.Args[i], "--queue=")
		case strings.HasPrefix(os.Args[i], "--workers="):
			var err error
			workers, err = strconv.Atoi(strings.TrimPrefix(os.Args[i], "--workers="))
			if err != nil {
				log.Fatalf("Invalid value for --workers: %v", err)
			}
		case strings.HasPrefix(os.Args[i], "--data"):
			// 单独处理--data task.json不带"="的情况
			if i+1 < len(os.Args) {
				dataFile = os.Args[i+1]
			} else {
				log.Fatal("Missing value for --data")
			}
		}
	}

	switch action {
	case "consumer":
		consumer := biz.NewConsumer(configConent, queueName, workers)
		consumer.Consume()
	case "producer":
		producer := biz.NewProducer(configConent, queueName)
		producer.Produce(dataFile)
	default:
		log.Fatal("Invalid action. Use 'consumer' or 'producer'.")
	}
}
