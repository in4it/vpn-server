package logging

import "fmt"

var Loglevel = 3

const LOG_ERROR = 1
const LOG_INFO = 2
const LOG_DEBUG = 16

func DebugLog(err error) {
	if Loglevel&LOG_DEBUG == LOG_DEBUG {
		fmt.Println("debug: " + err.Error())
	}
}

func ErrorLog(err error) {
	if Loglevel&LOG_ERROR == LOG_ERROR {
		fmt.Println("error: " + err.Error())
	}
}

func InfoLog(info string) {
	if Loglevel&LOG_INFO == LOG_INFO {
		fmt.Println("info: " + info)
	}
}
