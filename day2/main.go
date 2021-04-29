package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

var fileCount int = 0

func main() {
	start := time.Now()
	//path := strings.ReplaceAll(`C:\Users\Administrator\Desktop\Client_frp\`, "\\", "/")
	path := strings.ReplaceAll(`D:\`, "\\", "/")
	Search(path)
	fmt.Println(fileCount, time.Since(start))
}

func Search(path string) {
	files, err := ioutil.ReadDir(path)
	if err == nil {
		for _, file := range files {
			fileName := file.Name()
			if strings.Contains(fileName, "0") || strings.Contains(fileName, "1") || strings.Contains(fileName, "a") {
				fileCount++
			}
			if file.IsDir() {
				Search(path + fileName + "/")
				if fileCount < 10 {
					fmt.Println(fileName)
				}
				fileCount++
			}
		}
	} else {
		fmt.Println(err.Error())
	}
}

// func test1() {
// 	var wg sync.WaitGroup
// 	wg.Add(2)
// 	go func() {
// 		count("羊", 3)
// 		wg.Done()
// 	}()
// 	go func() {
// 		count("牛", 3)
// 		wg.Done()
// 	}()
// 	wg.Wait()
// }

// func count(who string, num int) {
// 	for i := 0; i < num; i++ {
// 		fmt.Println(who)
// 	}
// }
