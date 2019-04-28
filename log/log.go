package log

import "fmt"

var Info func(f interface{}, v ...interface{}) = printInfo
var Error func(f interface{}, v ...interface{}) = printError

func printInfo(f interface{}, v ...interface{}) {
	fmt.Println("INF: ", f, v)
}

func printError(f interface{}, v ...interface{}) {
	fmt.Println("ERR: ", f, v)
}

/// 注册日志函数
func RegisterLogFunc(err, inf func(f interface{}, v ...interface{})) {
	Info = inf
	Error = err
}
