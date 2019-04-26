package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zhaobingss/sqlmap/engine"
	"time"
)

func main() {

	eg := engine.New()
	err := eg.Init("mysql", "root:root@(127.0.0.1:3306)/test", "sql")
	if err != nil {
		panic(err)
	}
	eg.GetDB().SetMaxOpenConns(200)
	eg.GetDB().SetMaxIdleConns(200)
	for i := 0; i < 900; i++ {
		go func(i int) {
			mp := map[string]interface{}{}
			mp["id"] = 2
			m, err := eg.Query("my.selectALL", mp)
			if err != nil {
				panic(err)
			}
			fmt.Println(i, m)
		}(i)
	}

	//ss := eg.NewSession()
	//err = ss.BeginTx()
	//if err != nil {
	//	panic(err)
	//}
	//mp := map[string]interface{}{}
	//mp["name"] = "zhangsan"
	//mp["pass"] = "123"
	//r, err := ss.Exec("my_insert", mp)
	//if err != nil {
	//	panic(err)
	//}
	//
	//id, _ := r.LastInsertId()
	//fmt.Println(id)
	//
	//err = ss.Commit()
	//if err != nil {
	//	panic(err)
	//}
	//time.Sleep(time.Duration(10) * time.Second)
	//
	//err = ss.Rollback()
	//if err != nil {
	//	panic(err)
	//}


	time.Sleep(time.Duration(300) * time.Second)
}
