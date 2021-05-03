package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

var currentWorker int = 0
var maxWorker int = 3
var workDone chan bool = make(chan bool)
var queryRequest chan bool = make(chan bool)
var isFind chan bool = make(chan bool)
var foudCount = 0
var dbMaxRow int = 100 //2149768
var addRowIndex chan bool = make(chan bool)
var curStartRow int = 0
var curEndRow int = 0
var step int = 9

//var currentTreadQuery chan bool = make(chan bool)

func main() {
	startTime := time.Now()
	currentWorker = 1
	curStartRow = 1
	curEndRow = 10
	go Query(1, 10)
	WaitGroup()
	fmt.Println(foudCount, time.Since(startTime))
}

func WaitGroup() {
	for {
		select {
		case <-addRowIndex: //新增下一次查询的下标
			curStartRow = curEndRow + 1
			curEndRow = curStartRow + step
		case <-isFind:
			foudCount++
		case <-queryRequest:
			currentWorker++
			curStartRow = curEndRow + 1
			curEndRow = curStartRow + step
			go Query(curStartRow, curEndRow)
		case <-workDone:
			currentWorker--
			//fmt.Println(currentWorker)
			if currentWorker == 0 {
				return
			}
		}
	}
}

func Query(strRow, endRow int) {
	fmt.Println(currentWorker, strRow, endRow)
	addRowIndex <- true
	if currentWorker < maxWorker { //新增Go routine
		queryRequest <- true
	}

	var server = "127.0.0.1"
	var port = 7001
	var user = "sa"
	var password = "xxx"
	var database = "xxx"

	//连接字符串
	connString := fmt.Sprintf("server=%s;port=%d;database=%s;user id=%s;password=%s", server, port, database, user, password)
	// if isdebug {
	// 	fmt.Println(connString)
	// }
	//建立连接
	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Fatal("Open Connection failed:", err.Error())
	}
	defer conn.Close()

	sqlContent := fmt.Sprintf("select FID,FPARAMCONTENT from T_SWSAPILOG where FID between %d and %d", strRow, endRow)
	//产生查询语句的Statement
	stmt, err := conn.Prepare(sqlContent)
	if err != nil {
		log.Fatal("Prepare failed:", err.Error())
	}
	defer stmt.Close()

	//通过Statement执行查询
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Query failed:", err.Error())
	}

	//遍历每一行
	for rows.Next() {
		var json string
		var id int
		rows.Scan(&id, &json)                               //将查到的数据写入到这行中
		if strings.Contains(json, "trade_partner.create") { //判断是否查到我想要的
			isFind <- true
		}
		//fmt.Printf("id:%v /t josn:%v ", id, json)
		//PrintRow(colsdata) //打印此行
	}

	if curEndRow <= dbMaxRow {
		Query(curStartRow, curEndRow)
	}

	if curEndRow >= dbMaxRow {
		workDone <- true
	}
	defer rows.Close()
}
