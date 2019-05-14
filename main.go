package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zhaobingss/sqlmap/engine"
	"time"
)

func main() {

	eg := engine.New()
	//err := eg.Init("mysql", "root:root@(127.0.0.1:3306)/test", "sql")
	err := eg.Init("mysql", "jipiao668:!@#jipiao@(139.196.210.136:3306)/jipiao668", "sql")
	if err != nil {
		panic(err)
	}
	eg.GetDB().SetMaxOpenConns(200)
	eg.GetDB().SetMaxIdleConns(200)


	//for i := 0; i < 900; i++ {
	//	go func(i int) {
	//		mp := map[string]interface{}{}
	//		mp["id"] = 2
	//		m, err := eg.Query("my.selectALL", mp)
	//		if err != nil {
	//			panic(err)
	//		}
	//		fmt.Println(i, m)
	//	}(i)
	//}

	//ss := eg.NewSession()
	//err = ss.BeginTx()
	//if err != nil {
	//	panic(err)
	//}
	//mp := map[string]interface{}{}
	//mp["name"] = "zhangsan"
	//mp["pass"] = "123"
	//r, err := ss.Exec("my.insert", mp)
	//if err != nil {
	//	panic(err)
	//}
	//
	//id, _ := r.LastInsertId()
	//fmt.Println(id)
	//
	//time.Sleep(time.Duration(10) * time.Second)
	//
	//err = ss.Commit()
	//if err != nil {
	//	panic(err)
	//}

	//err = ss.Rollback()
	//if err != nil {
	//	panic(err)
	//}

	//type Int int
	//var i Int
	//i = 5
	//
	//e := reflect.ValueOf(&i).Elem()
	//
	////name := (&e).Type().Name()
	//name := (&e).Elem().String()
	//fmt.Println(name)


	InitPolicy(eg)


	time.Sleep(time.Duration(300) * time.Second)
}

func InitPolicy(eg *engine.SqlEngine)  {

	pKeys, err := eg.Query("my.SelectPolicyKeyList", nil)
	if err != nil {
		panic(err)
	}

	if len(pKeys) < 1 {
		panic(err)
	}

	t := time.Now().UnixNano()/1000000
	for _, v := range pKeys {
		err := initPolicyByKey(eg, v)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println(time.Now().UnixNano()/1000000 - t)
}

func initPolicyByKey(eg *engine.SqlEngine, mapKey map[string]string) error {
	param := map[string]interface{}{}
	param["cid"] = mapKey["cid"]
	param["vendor"] = mapKey["vendor"]
	param["trip_type"] = mapKey["trip_type"]
	_, err := eg.Query("my.SelectPolicyList", param)

	return err
}
