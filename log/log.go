package log

import "fmt"

var Info = printInfo
var Error = printError

/// the default info log func
func printInfo(f interface{}, v ...interface{}) {
	fmt.Println("INF: ", f, v)
}

/// the default error log func
func printError(f interface{}, v ...interface{}) {
	fmt.Println("ERR: ", f, v)
}

/// register the log func
func RegisterLogFunc(err, inf func(f interface{}, v ...interface{})) {
	Info = inf
	Error = err
}
