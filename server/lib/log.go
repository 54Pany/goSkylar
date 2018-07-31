package lib

import (
	"log"
	"os"
)

func LogSetting() { //解析参数付给logF
	outfile, err := os.OpenFile(CurrentDate()+".log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}
	if outfile != nil {
		log.SetOutput(outfile) //设置log的输出文件，不设置log输出默认为stdout
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile) //设置答应日志每一行前的标志信息，这里设置了日期，打印时间，当前go文件的文件名
}
