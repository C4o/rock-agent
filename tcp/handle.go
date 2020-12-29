package tcp

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"github.com/C4o/rock-agent/common"
	"github.com/C4o/rock-agent/logger"
	"time"

	tp "github.com/henrylee2cn/erpc/v6"
	//tp "github.com/henrylee2cn/teleport"
)

const (
	RESTART = 0
	RELOAD  = 1
	START   = 2
	STOP    = 3
)

type Frame struct {
	Type   int    // 1: rock.d 2: pool.d 3: resty.d 4:server.d 5:ssl.d 6:rock.conf
	Action int    // 1: update 2:delete
	Name   string // item name
	Data   []byte // data
}

type Data struct {
	Host   string `json:"host"`
	IP     string `json:"ip"`
	UUID   string `json:"uuid"`
	Status string `json:"status"`
}

func (data *Data) String() string {

	clientInfo, err := json.Marshal(data)
	if err != nil {
		return "NULL"
	}
	return string(clientInfo)

}

func (frame *Frame) ERR(code int32, r interface{}) *tp.Status {
	logger.ERR(logger.DEBUG, "%d:%d:%s %v", frame.Type, frame.Action, frame.Name, r)
	return nil
}

func (frame *Frame) Remove(path string) *tp.Status {
	err := os.Remove(path)
	if err != nil {
		return frame.ERR(500, err)
	}

	return frame.ERR(200, "successful")
}

func (frame *Frame) Write(path string) *tp.Status {
	err := ioutil.WriteFile(path, frame.Data, 0644)
	if err != nil {
		return frame.ERR(500, err)
	}

	return frame.ERR(200, "successful")
}

func (frame *Frame) String() string {
	return fmt.Sprintf("%d:%d:%s:%s", frame.Type, frame.Action, frame.Name, frame.Data)
}

func (frame *Frame) Path() string {
	switch frame.Type {
	case 1: //rock.d
		return fmt.Sprintf("%s/rock.d/%s", common.Conf.Rock, frame.Name)
	case 2: //pool.d
		return fmt.Sprintf("%s/pool.d/%s", common.Conf.Rock, frame.Name)
	case 3: //resty.d
		return fmt.Sprintf("%s/resty.d/%s", common.Conf.Rock, frame.Name)
	case 4: //server.d
		return fmt.Sprintf("%s/conf/server.d/%s", common.Conf.Rock, frame.Name)
	case 5: //ssl.d
		return fmt.Sprintf("%s/conf/ssl.d/%s", common.Conf.Rock, frame.Name)
	case 6: // rock.conf
		return fmt.Sprintf("%s/conf/rock.conf", common.Conf.Rock)
	default:
		return "/dev/null"
	}
}

func (frame *Frame) Operation(op int) *tp.Status {

	var cmd *exec.Cmd
	var stderr bytes.Buffer

	switch op {
	case 1:
		cmd = exec.Command("systemctl", "reload", "rock")
	case 2:
		cmd = exec.Command("systemctl", "restart", "rock")
	}

	cmd.Stderr = &stderr
	if e := cmd.Start(); e != nil {
		return tp.NewStatus(500, stderr.String(), e)
	}

	if e := cmd.Wait(); e != nil {
		return tp.NewStatus(500, stderr.String(), e)
	}
	return frame.ERR(200, "successful")

}

func (frame *Frame) Update() *tp.Status {
	switch frame.Type {
	case 1, 2, 3: //rock.d
		return frame.Write(frame.Path())
	case 4, 5, 6:
		frame.Write(frame.Path())
		return frame.Operation(1)
	default:
		return frame.ERR(401, "not found") // not found type
	}
}

func (frame *Frame) Delete() *tp.Status {
	switch frame.Type {
	case 1, 6: //rock.d
		return frame.ERR(403, "Permission denied")
	case 2, 3, 4, 5: // pool.d , resty.d
		return frame.Remove(frame.Path())
	default:
		return frame.ERR(402, "not found") // not found type
	}
}

func (frame *Frame) Do() *tp.Status {
	switch frame.Action {
	case 1: //update
		return frame.Update()

	case 2: //delete
		return frame.Delete()
	default:
		return frame.ERR(401, "not found action")
	}
}

type Agent struct {
	tp.PushCtx
}

type ServerUpdate struct {
	Hash  string
	Agent []byte
}

func (agent *Agent) Recv(frame *Frame) *tp.Status {
	//logger.ERR(logger.DEBUG, "receiver frame from /agent/recv : %v", frame)
	return frame.Do()
}

// 刪除agent日志
//func (agent *Agent) Dela(server *int) *tp.Status {

//// 如果接收到0，就删除agent日志
//if *server == 0 {
//var f *os.File
//var err error

//f, err = os.OpenFile(common.Conf.Error.Dir, os.O_WRONLY|os.O_TRUNC, 066)
//if err != nil {
//logger.ERR(logger.ERROR, "open agent log error : %v", err)
//}
//defer f.Close()
//_, err = f.Write([]byte(""))
//if err != nil {
//logger.ERR(logger.ERROR, "delete agent log error : %v", err)
//}

//}
//return nil
//}

// 接收服务端传递的配置参数并修改配置文件
func (agent *Agent) Reloadc(server *common.Server) *tp.Status {

	var err error
	logger.ERR(logger.DEBUG, "reloadc %v ", server)
	err = server.ConfToFile()
	if err != nil {
		logger.ERR(logger.ERROR, "conftofile error : %v", err)
	} else {
		if err = common.LoadConf(); err != nil {
			logger.ERR(logger.INFO, "LoadConf ERROR.")
		}
	}

	return nil

}

// 修改完配置文件后重连服务端
func (agent *Agent) Reloads(server *common.Server) *tp.Status {

	var err error
	logger.ERR(logger.DEBUG, "reloads %v ", server)
	err = server.ConfToFile()
	if err != nil {
		logger.ERR(logger.ERROR, "conftofile error : %v", err)
	} else {
		if err = common.LoadConf(); err != nil {
			logger.ERR(logger.ERROR, "LoadConf ERROR.")
		}
		time.Sleep(time.Second)
		// 更新配置后直接硬重启
		cmd := exec.Command("systemctl", "restart", "agent")
		cmd.Start()
		logger.ERR(logger.ERROR, "restart agent over.")
		//Cli.Reload()
	}

	return nil

}

// 更新agent可执行文件并强制重启
func (agent *Agent) Update(u *ServerUpdate) *tp.Status {

	logger.ERR(logger.DEBUG, "Hash of new agent : %s", u.Hash)

	rock := common.Conf.Rock
	var err error

	// 校验agent的hash
	if fmt.Sprintf("%x", md5.Sum(append(u.Agent, "em.sec.140986"...))) != u.Hash {
		logger.ERR(logger.ERROR, "wrong size with agent, fail to update.")
		return nil
	}

	// 删除上次更新的老版本可执行文件
	if err = os.Remove(rock + "/sbin/agent-old"); err != nil {
		logger.ERR(logger.ERROR, "delete old agent error : %v", err)
	}
	logger.ERR(logger.DEBUG, "delete old version agent")

	// 重命名agent备份
	if err = os.Rename(rock+"/sbin/agent", rock+"/sbin/agent-old"); err != nil {
		logger.ERR(logger.ERROR, "rename old agent failed for : %v.", err)
		return nil
	}

	// 下载新agent并给执行权限
	if err = ioutil.WriteFile(rock+"/sbin/agent", u.Agent, 755); err != nil {
		logger.ERR(logger.ERROR, "write agent file error : %v", err)
		// 下载失败后恢复旧版本agent
		if err = os.Rename(rock+"/sbin/agent-old", rock+"/sbin/agent"); err != nil {
			logger.ERR(logger.ERROR, "recover agent from old version failed for : %v.", err)
			return nil
		}
	}

	// 重启agent服务
	logger.ERR(logger.ERROR, "restart agent start.")
	cmd := exec.Command("systemctl", "restart", "agent")
	cmd.Start()
	logger.ERR(logger.ERROR, "restart agent over.")

	return nil

}

func (agent *Agent) Systemctl(do *int) *tp.Status {
	logger.ERR(logger.DEBUG, "/agent/systemctl req : %d ", &do)
	var cmd *exec.Cmd
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	switch *do {
	case START:
		logger.ERR(logger.DEBUG, "/agent/systemctl req : start")
		cmd = exec.Command("systemctl", "start", "rock")
	case STOP:
		logger.ERR(logger.DEBUG, "/agent/systemctl req : stop ")
		cmd = exec.Command("systemctl", "stop", "rock")
	case RELOAD:
		logger.ERR(logger.DEBUG, "/agent/systemctl req : reload ")
		cmd = exec.Command("systemctl", "reload", "rock")
	case RESTART:
		logger.ERR(logger.DEBUG, "/agent/systemctl req : restart")
		cmd = exec.Command("systemctl", "restart", "rock")
	default:
		logger.ERR(logger.DEBUG, "/agent/systemctl req : default")
		return tp.NewStatus(404, "not found systemctl action :%d", *do)
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if e := cmd.Start(); e != nil {
		return tp.NewStatus(500, stderr.String(), e)
	}

	if e := cmd.Wait(); e != nil {
		return tp.NewStatus(500, stderr.String(), e)
	}

	return nil
}
