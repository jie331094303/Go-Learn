package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/google/uuid"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type TabelModel struct {
	name     string
	table    string
	subClass string
	entity   []Entity
}

type Entity struct {
	name        string
	dbColumn    string
	fieldType   string
	description string
}

var totalCout int = 0
var filterCout int = 0
var dealErroUrlCount int = 0
var spiderDoneCount = 0
var spiderDBRecordCount int64 = 0
var findNeedSpiderUrlCount chan int = make(chan int)
var findDealErro chan bool = make(chan bool)
var findFilter chan bool = make(chan bool)
var spiderDone chan bool = make(chan bool)
var workDone chan bool = make(chan bool)
var curSpiderDBCount chan int64 = make(chan int64)

func main() {
	InitLogParamter()
	indexUrls := [...]string{
		"https://open.kingdee.com/K3Cloud/PDM/BD基础_files/BD基础_toc.html",
		"https://open.kingdee.com/K3Cloud/PDM/FIN财务_files/FIN财务_toc.html",
		"https://open.kingdee.com/K3Cloud/PDM/SCM供应链_files/SCM供应链_toc.html",
		"https://open.kingdee.com/K3Cloud/PDM/SCO供应链协同_files/SCO供应链协同_toc.html",
		"https://open.kingdee.com/K3Cloud/PDM/MFG制造_files/MFG制造_toc.html",
		"https://open.kingdee.com/K3Cloud/PDM/成本管理_files/成本管理_toc.html",
		"https://open.kingdee.com/K3Cloud/PDM/集团财务_files/集团财务_toc.html",
	}
	now := time.Now()
	for _, indexUrl := range indexUrls {
		go Spider(indexUrl)
	}
	chanManager()
	log.Println("耗时:", time.Since(now))
	log.Printf("全部网页为:%d", totalCout)
	log.Printf("过滤网页为:%d", filterCout)
	log.Printf("处理异常网页为:%d", dealErroUrlCount)
	log.Printf("爬取到的网页为:%d", spiderDoneCount)
	log.Printf("爬取到数据库的行数:%d", spiderDBRecordCount) //包含表头spiderDoneCount，单据的真实数据要减去spiderDoneCount

}

func InitLogParamter() {
	logFile, err := os.OpenFile("C:/Users/Administrator/Desktop/log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return
	}
	log.SetOutput(logFile)
	log.SetOutput(logFile)
	log.SetFlags(log.Lmicroseconds | log.Ldate)
}

func chanManager() {
	var maxWorker int
	for {
		select {
		case urlCount := <-findNeedSpiderUrlCount:
			totalCout += urlCount
		case <-findFilter:
			filterCout++
		case <-findDealErro:
			dealErroUrlCount++
		case <-spiderDone:
			spiderDoneCount++
		case spiderDBCount := <-curSpiderDBCount:
			spiderDBRecordCount += spiderDBCount
		case <-workDone:
			maxWorker++
			if maxWorker == 7 {
				return
			}

		}
	}
}

func Spider(indexUrl string) {
	//获取拼接标题前缀
	lastStr := strings.Split(indexUrl, "/")[len(strings.Split(indexUrl, "/"))-1]
	typeName := strings.Split(lastStr, "_")[0]
	requestUrls := GetSpiderUrls(typeName, indexUrl)
	urlTitle := strings.ReplaceAll(indexUrl, lastStr, "")
	findNeedSpiderUrlCount <- len(requestUrls)
	for _, url := range requestUrls {
		//fmt.Println(urlTitle + url)
		StartSpider(urlTitle + url)
	}
	workDone <- true
	// testUrl := "https://open.kingdee.com/K3Cloud/PDM/BD基础_files/Home_LightBlue.html"
	// StartSpider(testUrl)

}

func GetSpiderUrls(typeName, indexUrl string) []string {
	doc := GetUrlToDocumentRoot(indexUrl)
	if doc.Error != nil {
		log.Printf("Get %s doc is erro", indexUrl)
	}
	requestUrls := GetUrlLinkData(typeName, doc)
	return requestUrls
}

func GetUrlToDocumentRoot(url string) soup.Root {
	var doc soup.Root
	res, erro := http.Get(url)
	if erro != nil {
		log.Printf("Request %s erro \n", url)
	}
	if res.StatusCode == 200 {
		defer res.Body.Close()
		bodyStr := DealGBKToUTF8(res.Body)
		doc = soup.HTMLParse(bodyStr)
	} else {
		log.Printf("Get %s StatusCode is falid", url)
	}
	return doc
}

func DealGBKToUTF8(body io.ReadCloser) string {
	gbkBody, _ := ioutil.ReadAll(body)
	reader := transform.NewReader(bytes.NewReader(gbkBody), simplifiedchinese.GBK.NewDecoder())
	utf8Body, _ := ioutil.ReadAll(reader)
	return string(utf8Body)
}

func GetUrlLinkData(typeName string, doc soup.Root) []string {
	urls := []string{}
	root := doc.FindAll("div", "class", "sngl")
	for _, son := range root {
		source := son.Find("a").Attrs()["href"]
		if strings.Contains(source, typeName) {
			urls = append(urls, source)
		}

	}
	return urls
}

func StartSpider(url string) {
	doc := GetUrlToDocumentRoot(url)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("过滤h2是否有数据表描述失败:%v", url)
			findFilter <- true
		}
	}()
	h2content := doc.Find("h2").Find("a").Text()

	if strings.Contains(h2content, "数据表") {
		var model TabelModel
		isHeadErro := GetModelHead(url, doc, &model)
		if !isHeadErro {
			isEntityErro := GetModelEntity(url, doc, &model)
			if !isEntityErro {
				InsertToDB(url, &model)
			}
		}
	} else {
		log.Printf("当前网页不需要spider:%s", url)
		findFilter <- true
	}

}

func GetModelHead(url string, doc soup.Root, model *TabelModel) bool {
	var isErro bool = false
	tableDec := doc.Find("table", "class", "Form").Find("tbody").FindAll("tr")
	for index, root := range tableDec {
		if !(index == 0 || index == 1 || index == 3) {
			continue
		}

		var content string = ""
		subRoot := root.Find("td").FindNextElementSibling().Find("p")
		//subRoot := soup.HTMLParse(html)
		if strings.Contains(subRoot.HTML(), "<a") {
			content = subRoot.Find("a").Text()
		} else {
			content = subRoot.Text()
		}
		content = strings.ReplaceAll(strings.TrimSpace(content), "'", "\"")
		switch index {
		case 0:
			model.name = content
		case 1:
			model.table = content
		case 3:
			model.subClass = content
		}
		defer func() {
			if err := recover(); err != nil {
				log.Printf("获取头部异常:%s", url)
				isErro = true
				findDealErro <- true
			}
		}()
	}

	return isErro
}

func GetModelEntity(url string, doc soup.Root, model *TabelModel) bool {
	var isErro bool = false
	nodes := doc.Find("table", "class", "Grid").Find("tbody").FindAll("tr")
	for index, node := range nodes {
		if index == 0 {
			continue
		}
		var entity Entity
		tdRoots := node.FindAll("td")
		for j, tdRoot := range tdRoots {
			content := tdRoot.Find("p").Text()
			if j == 3 || j == 4 {
				continue
			}

			content = strings.ReplaceAll(strings.TrimSpace(content), "'", "\"")
			switch j {
			case 0:
				entity.name = content
			case 1:
				entity.dbColumn = content
			case 2:
				entity.fieldType = content
			case 5:
				entity.description = content
			}
		}
		model.entity = append(model.entity, entity)
	}
	defer func() {
		if err := recover(); err != nil {
			log.Printf("获取头部单据体异常:%s", url)
			isErro = true
			findDealErro <- true
		}
	}()
	return isErro
}

func InsertToDB(url string, model *TabelModel) {
	var server = "192.168.83.144"
	var port = 1433
	var user = "sa"
	var password = "kingdee@2018"
	var database = "Test"
	sqlStr := GetSql(url, model)
	//fmt.Println(sqlStr)
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

	//产生查询语句的Statement
	result, err := conn.Exec(sqlStr)
	if err != nil {
		log.Fatal("insert failed:", err.Error())
	}
	effectRow, err := result.RowsAffected()
	if err != nil {
		log.Fatal("insert failed:", err.Error())
	}
	if effectRow > 0 {
		log.Printf("成功抓取：%s", url)
		curSpiderDBCount <- effectRow
	}
	spiderDone <- true

}

func GetSql(url string, model *TabelModel) string {
	var sqlStr string
	uuidWithHyphen := uuid.New()
	//fmt.Println(uuidWithHyphen)
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	//fmt.Println(uuid)

	sqlStr = fmt.Sprintf(`
		INSERT INTO [dbo].[T_KD]
				([FID]
				,[FUrl]
				,[FName]
				,[FTable]
				,[FSubClass])
		VALUES	('%s','%s','%s','%s','%s')
		`, uuid, url, model.name, model.table, model.subClass)

	for _, v := range model.entity {
		sqlStr += fmt.Sprintf(`
		INSERT INTO [dbo].[T_KDEntry]
				([FID]
				,[FName]
				,[FDBColumn]
				,[FFieldType]
				,[FDescription])
		VALUES ('%s','%s','%s','%s','%s')
		`, uuid, v.name, v.dbColumn, v.fieldType, v.description)
	}
	return sqlStr
}
