package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/anaskhan96/soup"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func main() {
	var url string = "https://open.kingdee.com/K3Cloud/PDM/BD基础_files/BD基础_toc.html"
	res, erro := http.Get(url)
	if erro != nil {
		log.Printf("Request %s erro", url)
	}
	defer res.Body.Close()
	gbkBody, _ := ioutil.ReadAll(res.Body)
	reader := transform.NewReader(bytes.NewReader(gbkBody), simplifiedchinese.GBK.NewDecoder())
	utf8Body, _ := ioutil.ReadAll(reader)
	bodyStr := string(utf8Body)
	doc := soup.HTMLParse(bodyStr)
	root := doc.FindAll("div", "class", "sngl")
	for _, son := range root {
		source := son.Find("a").Attrs()["href"]
		table := son.Find("img").Attrs()["title"]
		fmt.Println(source, table)
	}

}
