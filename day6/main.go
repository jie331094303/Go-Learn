package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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

func main() {
	indexUrl := "https://open.kingdee.com/K3Cloud/PDM/BD基础_files/BD基础_toc.html"
	Spider(indexUrl)
}

func Spider(indexUrl string) {
	//获取拼接标题前缀
	requestUrls := GetSpiderUrls(indexUrl)
	urlTitle := strings.ReplaceAll(indexUrl, strings.Split(indexUrl, "/")[len(strings.Split(indexUrl, "/"))-1], "")
	for _, url := range requestUrls {
		fmt.Println(urlTitle + url)
		testurl := "https://open.kingdee.com/K3Cloud/PDM/BD%E5%9F%BA%E7%A1%80_files/BD%E5%9F%BA%E7%A1%8014.htm"
		StartSpider(testurl)
	}

}

func GetSpiderUrls(indexUrl string) []string {
	doc := GetUrlToDocumentRoot(indexUrl)
	if doc.Error != nil {
		log.Printf("Get %s doc is erro", indexUrl)
	}
	requestUrls := GetUrlLinkData(doc)
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

func GetUrlLinkData(doc soup.Root) []string {
	urls := []string{}
	root := doc.FindAll("div", "class", "sngl")
	for i, son := range root {
		if i == 7 {
			source := son.Find("a").Attrs()["href"]
			//table := son.Find("img").Attrs()["title"]
			//fmt.Println(source, table)
			urls = append(urls, source)
		}
	}
	return urls
}

func StartSpider(url string) {
	doc := GetUrlToDocumentRoot(url)
	h2content := doc.Find("h2").Find("a").Text()
	if strings.Contains(h2content, "数据表") {
		var model TabelModel
		GetModelHead(doc, &model)
		GetModelEntity(doc, &model)
		InsertToDB(url, &model)
	} else {
		log.Printf("当前网页不需要spider:%s", url)
	}

}

func GetModelHead(doc soup.Root, model *TabelModel) {
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
		content = strings.TrimSpace(content)
		switch index {
		case 0:
			model.name = content
		case 1:
			model.table = content
		case 3:
			model.subClass = content
		}
	}
}

func GetModelEntity(doc soup.Root, model *TabelModel) {
	//entitys := model []Entity{}
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
			content = strings.TrimSpace(content)
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
}

func InsertToDB(url string, model *TabelModel) {
	var server = "192.168.83.147"
	var port = 1433
	var user = "sa"
	var password = "abv@2018"
	var database = "Test"
	sqlStr := GetSql(url, model)
	fmt.Println(sqlStr)
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
	log.Printf("插入%d行数据", effectRow)
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
