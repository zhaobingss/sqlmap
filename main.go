package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zhaobingss/sqlmap/engine"
	"github.com/zhaobingss/sqlmap/template"
	"time"
)

func main() {

	eg := engine.New()
	err := eg.Init("mysql", "root:root@(127.0.0.1:3306)/test", "sql", template.DEFAULT)
	if err != nil {
		panic(err)
	}
	eg.GetDB().SetMaxOpenConns(800)
	eg.GetDB().SetMaxIdleConns(200)

	for i := 0; i < 900; i++ {
		go func() {
			mp := map[string]interface{}{}
			mp["id"] = 2
			m, err := eg.Query("my_selectALL", mp)
			if err != nil {
				panic(err)
			}
			fmt.Println(m)
		}()
	}


	time.Sleep(time.Duration(300) * time.Second)
}
