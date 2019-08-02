package test

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zhaobingss/sqlmap/engine"
	"sync"
	"testing"
)

type Resource struct {
	ID          int            `db:"id"`
	Pid         int            `db:"pid"`
	Type        string         `db:"type"`
	Name        string         `db:"name"`
	Code        string         `db:"code"`
	Description string         `db:"description"`
	URL         sql.NullString `db:"url"`
	Icon        string         `db:"icon"`
	Seq         int            `db:"seq"`
	CreateTime  string         `db:"create_time"`
	UpdateTime  string         `db:"update_time"`
}
var eg *engine.SqlEngine
var wg sync.WaitGroup
func init() {
	var err error
	eg, err = engine.NewEngine("mysql", "root:root@(127.0.0.1:3306)/test", "E:/project/mine/go/sqlmap/sql")
	if err != nil {
		panic(err)
	}
	eg.GetDB().SetMaxOpenConns(5)
	eg.GetDB().SetMaxIdleConns(5)
}

func TestSliceStruct_test(t *testing.T)  {
	srcs := make([]*Resource, 0)
	err := eg.Select(&srcs, `my.selectALL`, nil)
	if err != nil {
		panic(err)
	}
	for _, v := range srcs {
		fmt.Println(v)
	}
}

func TestStruct_test(t *testing.T) {
	src := Resource{}
	err := eg.SelectOne(&src, "my.selectOne", nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(src)
}

func TestMap_test(t *testing.T) {
	ret, err := eg.Query("my.selectALL", nil)
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range ret {
		fmt.Println(v)
	}
}

func TestConcurrency_test(t *testing.T)  {
	wg.Add(1)
	go TestSliceStruct_test(nil)
	go TestMap_test(nil)
	go TestStruct_test(nil)
	wg.Wait()
}
