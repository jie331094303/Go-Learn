package main

//chan goroutine learn
import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

var matchCount int = 0
var currentWorkCout int = 0
var maxWorker int = 32
var searchRequest = make(chan string)
var workDone chan bool = make(chan bool)
var foundMatch chan bool = make(chan bool)

func main() {
	start := time.Now()
	var searchLocation string = strings.ReplaceAll(`C:\`, "\\", "/")
	currentWorkCout = 1
	go Search(searchLocation, true)
	WaitForWorker()
	fmt.Println("匹配数目:", matchCount)
	fmt.Print("花费了", time.Since(start))
}

func Search(path string, master bool) {
	fmt.Println("工人数：", currentWorkCout)
	files, erro := ioutil.ReadDir(path)
	if erro == nil {
		for _, file := range files {
			//var fileName string = file.Name()
			fileName := file.Name()
			if strings.Contains(fileName, "0") {
				foundMatch <- true
			}
			if file.IsDir() {
				if currentWorkCout < maxWorker {
					searchRequest <- path + fileName + "/"
				} else {
					Search(path+fileName+"/", false)
				}
			}
		}
		if master {
			workDone <- true
		}
	} else {
		workDone <- true
		//fmt.Println("ERRO!!", erro.Error())
	}
}

func WaitForWorker() {
	for {
		select {
		case <-foundMatch:
			matchCount++
		case path := <-searchRequest:
			currentWorkCout++
			go Search(path, true)
		case <-workDone:
			currentWorkCout--
			if currentWorkCout == 0 {
				return
			}
		}
	}
}
