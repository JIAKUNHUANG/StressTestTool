package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	URL         string `yaml:"url"`
	Concurrency int    `yaml:"concurrency"`
	Duration    string `yaml:"duration"`
	Data        string `yaml:"data"`
	Cookie      string `yaml:"cookie"`
}

func main() {
	// 读取配置文件
	configFile := "config.yml"
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	var count int
	var mu sync.Mutex

	duration, err := time.ParseDuration(config.Duration)
	if err != nil {
		fmt.Printf("Error parsing duration: %v\n", err)
		os.Exit(1)
	}

	// 创建HTTP客户端
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        config.Concurrency * 2,
			MaxIdleConnsPerHost: config.Concurrency * 2,
			IdleConnTimeout:     10 * time.Second,
		},
	}

	start := time.Now()
	end := start.Add(duration)

	fmt.Printf("Starting stress test at %s\n", start.Format("2006-01-02 15:04:05"))
	// 并发执行
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go stressTestRoutinue(client, config.URL, []byte(config.Data), config.Cookie, &wg, &count, &mu, end)
	}

	wg.Wait()

	// 结果输出
	totalRequests := count
	qps := float64(totalRequests) / duration.Seconds()

	fmt.Printf("Total requests: %d\n", totalRequests)
	fmt.Printf("Total Time: %v\n", duration)
	fmt.Printf("QPS: %.2f\n", qps)
}

func stressTestRoutinue(client *http.Client, url string, data []byte, cookie string, wg *sync.WaitGroup, count *int, mu *sync.Mutex, end time.Time) {
	defer wg.Done()
	var routinue int

	for time.Now().Before(end) {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		/* 		// 随机数生成cookie
		   		rand.Seed(time.Now().UnixNano())
		   		randNum := rand.Intn(100000)
		   		randNumStr := fmt.Sprintf("%08d", randNum)
		   		cookie := "token=" + randNumStr + "ASDFGHJKLQWERTYUIOPZXCVBNM" */

		req.Header.Set("Cookie", cookie)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			continue
		}

		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			continue
		}

		routinue++
	}
	mu.Lock()
	*count += routinue
	mu.Unlock()
}
