package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/denisenkom/go-mssqldb"
)

type BModel struct {
	FPic       string
	FTitle     string
	FBvid      string
	FPalyCount int
	FComment   int
}

func main() {
	http.HandleFunc("/", QueryHTX) //bilibili
	http.ListenAndServe(":9000", nil)
}

func QueryHTX(w http.ResponseWriter, r *http.Request) {
	//构建模板
	temp, erro := template.ParseFiles("./b.tmpl")
	if erro != nil {
		fmt.Println("Get Template Erro")
	}
	bModels := QueryBData("老师好我叫何同学")
	// temp.Execute(w, map[string]interface{}{
	// 	"Data": bModels,
	// })

	temp.Execute(w, bModels)
}

func QueryBData(author string) []BModel {
	//workStatus <- false
	//fmt.Println(currentWorker, strRow, endRow)
	var server = "192.168.83.144"
	var port = 1433
	var user = "sa"
	var password = "kingdee@2018"
	var database = "Test"

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
	sqlContent := fmt.Sprintf("select FPic,FTitle,FBvid,FPalyCount,FComment from T_UPInfo where FAuthor = '%s'", author)

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

	models := []BModel{}
	//遍历每一行
	for rows.Next() {
		dbModel := BModel{}
		rows.Scan(&dbModel.FPic, &dbModel.FTitle, &dbModel.FBvid, &dbModel.FPalyCount, &dbModel.FComment) //将查到的数据写入到这行中
		dbModel.FBvid = "https://www.bilibili.com/video/" + dbModel.FBvid
		//if strings.Contains(dbModel.json, "trade_partner.create") { //判断是否查到我想要的
		//isFind <- true
		models = append(models, dbModel)
		//fmt.Printf("id:%v \n ", id)
		//}
		//fmt.Printf("id:%v /t josn:%v ", id, json)
		//PrintRow(colsdata) //打印此行
	}
	//workStatus <- true
	defer rows.Close()
	return models
}
