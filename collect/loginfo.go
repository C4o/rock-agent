package collect

import (
	"fmt"
	"os"
	"github.com/C4o/rock-agent/common"
	"github.com/C4o/rock-agent/logger"
	"github.com/C4o/rock-agent/tcp"
	"strings"
	"time"

	tp "github.com/henrylee2cn/erpc/v6"
	//tp "github.com/henrylee2cn/teleport"
)

type LogSize struct {
	Name string
	Size int64
}

type LogInfo struct {
	Path     string
	FileData []LogSize
}

func (li LogInfo) Get() {

	logger.ERR(logger.DEBUG, "/access/getlog get.")
	// 获取前一天日志的路径
	d := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	var fn string
	var ls = LogSize{}

	for i := 0; i < 24; i++ {
		if i < 10 {
			fn = fmt.Sprintf("%s%s/access.%s.0%d.log.tar.gz", li.Path, d, d, i)
		} else {
			fn = fmt.Sprintf("%s%s/access.%s.%d.log.tar.gz", li.Path, d, d, i)
		}
		if fi, err := os.Stat(fn); err == nil {
			ls.Name = fn
			ls.Size = fi.Size()
			li.FileData[i] = ls
		} else {
			logger.ERR(logger.ERROR, "file : %s, error : %v\n", fn, err)
		}
	}

}

func (li LogInfo) Send() {

	logger.ERR(logger.DEBUG, "/access/getlog send data: %v.", li.FileData)
	session := tcp.Cli.Session
	session.Push("/rockp/accesslog", li.FileData)

}

type AccessFrame struct {
	Path string
}

type DeleteFrame struct {
	Dir []string
}

type Access struct {
	tp.PushCtx
}

func GenLog(path string) string {

	tmpArr := strings.Split(path, "/")
	path = ""
	for i := 0; i < len(tmpArr)-1; i++ {
		path += tmpArr[i] + "/"
	}
	if tmpArr[0] != "" {
		path = common.Conf.Rock + "/" + path
	}

	return path
}

func (a *Access) Getlog(s *AccessFrame) *tp.Status {

	logger.ERR(logger.DEBUG, "/access/getlog : %v", s)

	s.Path = GenLog(s.Path)

	li := LogInfo{Path: s.Path, FileData: make([]LogSize, 24)}
	li.Get()
	li.Send()
	return nil
}

func (a *Access) Deletelog(s *DeleteFrame) *tp.Status {

	var err error
	var filename string
	logger.ERR(logger.DEBUG, "/access/deletelog : %v", s)

	for _, v := range s.Dir {
		filename = GenLog(common.Conf.Access) + v
		if _, err = os.Stat(filename); err == nil {
			logger.ERR(logger.DEBUG, "filename is : %s", filename)
			if err = os.Remove(filename); err != nil {
				logger.ERR(logger.ERROR, "remove log file error : %v", err)
			}
		}
	}
	return nil
}
