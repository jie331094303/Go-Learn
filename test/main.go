package main

import (
	"fmt"
	"time"
)

func main() {
	//str := strconv.ParseInt("1405544146", 10, 64)
	date := time.Unix(1620722374, 0).Format("2006-01-02 15:04:05")
	date.Format("2006-01-02 15:04:05")
	fmt.Println(date.Format("2006-01-02 15:04:05"))
}
