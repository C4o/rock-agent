package common

import (
	"fmt"
	"log"
	"os"
)

var (
	ERROR = 0
	INFO  = 2
	DEBUG = 4
)

func NewLogger() error {
	uri := fmt.Sprintf("%s", Conf.Error.Dir)
	output, err := os.OpenFile(uri, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	log.SetOutput(output)

	return nil
}

func ERR(level int, format string, v ...interface{}) {

	debug := Conf.Debug

	if level == debug {
		slogger(level, format, v)
	} else {
		if debug > level {
			slogger(level, format, v)
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
