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
	path := strings.ReplaceAll(`C:\`, "\\", "/")
	Search(path)
	fmt.Println(fileCount, time.Since(start))
}

func Search(path string) {
	files, err := ioutil.ReadDir(path)
	if err == nil {
		for _, file := range files {
			fileName := file.Name()
			if strings.Contains(fileName, "0") {
				fileCount++
			}
			if file.IsDir() {
				//fmt.Println(path + fileName + "/")
				Search(path + fileName + "/")
			}
		}
	} else {
		//fmt.Println(err.Error())
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
