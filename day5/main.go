package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

type model struct {
	id   int
	json string
}

var currentWorker int = 0
var maxWorker int = 100
var workStatus chan bool = make(chan bool) //false 刚开工 true 完工了

var isFind chan bool = make(chan bool)
var foudCount = 0

//var currentTreadQuery chan bool = make(chan bool)

func main() {
	startTime := time.Now()
	fmt.Println(startTime)
	var dbMaxRow int = 2151444 //2151444
	var step int = dbMaxRow / maxWorker
	var startRow int = 1
	dbData := QueryAllData()
	for i := 0; i < maxWorker; i++ {
		currentWorker++
		endRow := startRow + step
		go QueryCount(dbData, startRow, endRow)
		startRow = endRow + 1
	}
	WaitGroup()
	fmt.Println(foudCount, time.Since(startTime))
}

func WaitGroup() {
	for {
		select {
		case <-isFind:
			foudCount++
		case workStatus := <-workStatus:
			if workStatus {
				currentWorker--
			} else {
				currentWorker++
			}
			//fmt.Println(currentWorker)
			if currentWorker == 0 {
				return
			}
		}
	}
}

func QueryAllData() []model {
	//workStatus <- false
	//fmt.Println(currentWorker, strRow, endRow)
	now := time.Now()
	var server = "192.168.83.140"
	var port = 1433
	var user = "sa"
	var password = "kingdee@2018"
	var database = "SHIWANSI"

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
	sqlContent := "select FID,FPARAMCONTENT from T_SWSAPILOG"

	//sqlContent := fmt.Sprintf("select FID,FPARAMCONTENT from T_SWSAPILOG where FID between %d and %d", strRow, endRow)
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

	models := []model{}
	//遍历每一行
	for rows.Next() {
		dbModel := model{}
		rows.Scan(&dbModel.id, &dbModel.json) //将查到的数据写入到这行中
		//if strings.Contains(dbModel.json, "trade_partner.create") { //判断是否查到我想要的
		//isFind <- true
		models = append(models, dbModel)
		//fmt.Printf("id:%v \n ", id)
		//}
		//fmt.Printf("id:%v /t josn:%v ", id, json)
		//PrintRow(colsdata) //打印此行
	}
	fmt.Printf("查询数据库时间:%v \n", time.Since(now))
	//workStatus <- true
	defer rows.Close()
	return models
}

func QueryCount(models []model, startRow int, endRow int) {
	fmt.Printf("查询区间:%d-%d\n", startRow, endRow)
	for _, v := range models {
		if startRow <= v.id && v.id <= endRow && strings.Contains(v.json, "trade_partner.create") {
			isFind <- true
		}
	}
	workStatus <- true
}
