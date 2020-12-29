package logger

import (
	"fmt"
	"log"
	"os"
	"github.com/C4o/rock-agent/common"
)

var (
	ERROR = 0
	INFO  = 2
	DEBUG = 4
)

func NewLogger() error {
	uri := fmt.Sprintf("%s", common.Conf.Error.Dir)
	output, err := os.OpenFile(uri, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	log.SetOutput(output)

	return nil
}

func ERR(level int, format string, v ...interface{}) {

	debug := common.Conf.Debug

	// 如果等级一样，就只打印该等级而没有其他等级日志
	if level == debug {
		slogger(level, format, v...)
	} else {
		// 如果等级不一样，就把小于配置文件等级的所有日志等级全部打印
		if debug > level {
			slogger(level, format, v...)
		}
	}
}

func slogger(level int, format string, v ...interface{}) {

	switch level {
	case 0:
		log.Println(fmt.Sprintf("[error] "+format, v...))
	case 2:
		log.Println(fmt.Sprintf("[info] "+format, v...))
	case 4:
		log.Println(fmt.Sprintf("[debug] "+format, v...))
	}
}
