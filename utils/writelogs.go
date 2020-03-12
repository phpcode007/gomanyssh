package utils

import (
	"log"
	"os"
)

func Write_log(writeFileName *os.File, logInfo string) {

	// 创建一个日志对象
	writeLog := log.New(writeFileName, "", log.LstdFlags)
	writeLog.Println(logInfo)
}
