package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"sync"
	"time"
)

type TestRequest struct {
	Serial string `json:"serial"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Imei   string `json:"imei"`
}

var keys []string

func getAllGoAsync() ([]TestRequest, error) {
	var wg sync.WaitGroup
	ch := make(chan TestRequest)

	client := getRedisClient()
	data, _ := client.MGet(keys...).Result()

	for _, i := range data {
		wg.Add(1)

		go (func(v interface{}, wg *sync.WaitGroup, ch chan TestRequest) {
			var res = &TestRequest{}
			str := fmt.Sprintf("%v", v)
			err := json.Unmarshal([]byte(str), res)
			if err == nil {
				ch <- *res
			}

			defer wg.Done()
		})(i, &wg, ch)
	}

	go func() {
		wg.Wait()
		defer close(ch)
	}()

	var result []TestRequest
	for i := range ch {
		result = append(result, i)
	}

	return result, nil
}

func addCache() {
	data := &TestRequest{
		Serial: "serial data",
		Imei:   "imei data",
		Width:  0,
		Height: 0,
	}

	client := getRedisClient()

	for i := 0; i < len(keys); i++ {
		key := keys[i]
		data.Serial = fmt.Sprintf("serial %d", i)
		data.Imei = fmt.Sprintf("imei %d", i)
		data.Width = i * 10
		data.Height = i * 3

		value, _ := json.Marshal(data)
		_, err := client.Set(key, value, time.Minute*60).Result()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func getRedisClient() redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	return *client
}

func main() {
	for i := 0; i < 500; i++ {
		keys = append(keys, fmt.Sprintf("key_%d", i))
	}

	//addCache()

	start := time.Now()
	result, err := getAllGoAsync()
	if err != nil {
		log.Panic(err.Error())
	}

	elapsed := time.Since(start).Milliseconds()

	fmt.Printf("process is completed in %d miliseconds. result length is %d", elapsed, len(result))
}
