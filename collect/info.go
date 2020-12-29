package collect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"github.com/C4o/rock-agent/common"
	"github.com/C4o/rock-agent/logger"
	"github.com/C4o/rock-agent/tcp"
	"time"

	//"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	pro "github.com/shirou/gopsutil/process"
)

var (
	FileList []string = []string{
		"conf/rock.conf",
		"rock.d/app",
		"rock.d/balancer",
		"rock.d/index",
		"rock.d/rewrite",
		"rock.d/security",
	}
	DirList []string = []string{
		"conf/server.d/",
		"conf/ssl.d/",
		"pool.d/",
		"resty.d/",
		"sbin/",
	}
)

type Data struct {
	Name string
	Time string
}

type FileTime struct {
	Data []Data
}

func (ft *FileTime) Get() {

	var err error
	var file, dir string
	var rd []os.FileInfo
	var f os.FileInfo
	var data = Data{}
	rock := common.Conf.Rock

	for _, file = range FileList {
		if f, err = os.Stat(rock + "/" + file); err == nil {
			data.Name = file
			data.Time = f.ModTime().Format("2006-01-02 15:04:05")
			ft.Data = append(ft.Data, data)
			//ft.Data[file] = f.ModTime().Format("2006-01-02 15:04:05")
		} else {
			logger.ERR(logger.ERROR, "[!] filetime GET function get file status error in file : %s", file)
		}
	}

	for _, dir = range DirList {
		if rd, err = ioutil.ReadDir(rock + "/" + dir); err == nil {
			for _, f = range rd {
				if !f.IsDir() {
					data.Name = dir + f.Name()
					data.Time = f.ModTime().Format("2006-01-02 15:04:05")
					ft.Data = append(ft.Data, data)
					//ft.Data[dir+f.Name()] = f.ModTime().Format("2006-01-02 15:04:05")
				}
			}
		} else {
			logger.ERR(logger.ERROR, "[!] filetime ReadDir error in dir : %s", dir)
		}
	}

}

type DiskInfo struct {
	Total  uint64
	Used   uint64
	Fstype string
}

type IMessage struct {
	RN   int                 //process number
	RS   bool                //rock process status
	RP   map[string]string   // process info
	MEM  map[string]uint64   // mem info
	CPU  map[string]float64  //cpu info
	Disk map[string]DiskInfo // disk info
	NET  map[string]int      // net info
	File FileTime            // file info

	Pri    string                 //private addr
	Hc     map[string]interface{} //healthcheck info
	Errlog []byte                 // rock errlog stat
}

type Info struct {
	Status bool
	URI    string      //sever path
	Client *tcp.Client // client
	Msg    *IMessage   //
}

func round(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}

func (info *Info) File() {

	info.Msg.File = FileTime{Data: make([]Data, 0)}
	info.Msg.File.Get()
	//fmt.Printf("filetime data : %s", info.Msg.File.Data)

}

func (info *Info) Disk() {
	info.Msg.Disk = map[string]DiskInfo{
		"/":               DiskInfo{0, 0, "unkown"},
		"/vdb":            DiskInfo{0, 0, "unkown"},
		"/var/log":        DiskInfo{0, 0, "unkown"},
		"/usr/local/rock": DiskInfo{0, 0, "unkown"},
	}

	for path, _ := range info.Msg.Disk {
		stat, err := disk.Usage(path)
		if err != nil {
			continue
		} else {
			info.Msg.Disk[path] = DiskInfo{stat.Total, stat.Used, stat.Fstype}
		}
	}
}

func (info *Info) Cpu() {
	info.Msg.CPU = make(map[string]float64)
	//if infos, err := cpu.Times(false); err != nil {
	//logger.ERR(logger.ERROR, "get cpu info fail %v", err)
	//return
	//} else {
	//v := infos[0]
	//total := v.Total()
	//info.Msg.CPU = map[string]float64{
	//"user":   round(v.User/total, 4),
	//"system": round(v.System/total, 4),
	//"idle":   round(v.Idle/total, 4),
	//"total":  round(total, 4),
	//}
	//}
	info.Msg.CPU = getCpu()
	//logger.ERR(logger.ERROR, "cpu is : %s", info.Msg.CPU)
}

func (info *Info) Mem() {
	info.Msg.MEM = make(map[string]uint64)
	if infos, err := mem.VirtualMemory(); err != nil {
		logger.ERR(logger.ERROR, "get mem info fail %v", err)
		return
	} else {
		info.Msg.MEM = map[string]uint64{
			"total": infos.Total,
			"used":  infos.Used,
			"Free":  infos.Free,
		}
	}
}

func (info *Info) Net() {
	info.Msg.NET = map[string]int{
		"TOTAL":       0,
		"ESTABLISHED": 0,
		"SYN_SENT":    0,
		"SYN_RECV":    0,
		"FIN_WAIT1":   0,
		"FIN_WAIT2":   0,
		"TIME_WAIT":   0,
		"CLOSE":       0,
		"CLOSE_WAIT":  0,
		"LAST_ACK":    0,
		"LISTEN":      0,
		"CLOSING":     0,
	}

	if nets, err := net.Connections("tcp"); err != nil {
		logger.ERR(logger.ERROR, "get net info fail %v", err)
		return
	} else {
		info.Msg.NET["TOTAL"] = len(nets)
		for _, v := range nets {
			info.Msg.NET[v.Status]++
		}
	}
}

func (info *Info) RockPro() {
	info.Msg.RN = 0
	info.Msg.RS = false
	info.Msg.RP = make(map[string]string)

	p, err := pro.Processes()
	if err != nil {
		logger.ERR(logger.ERROR, "get process fail %v", err)
		return
	}

	for _, v := range p {
		if name, _ := v.Name(); name == "rock" {
			cmdline, err := v.Cmdline()
			ctime, err := v.CreateTime()
			if err != nil {
				info.Msg.RS = false
				return
			}
			info.Msg.RS = true
			info.Msg.RN++
			info.Msg.RP[fmt.Sprintf("%d", v.Pid)] = fmt.Sprintf("%s %s", cmdline,
				time.Unix(ctime/1000, 0).Format("2006-01-02 15:04:05"))
		}
	}

}

func (info *Info) Errlog() {
	uri := fmt.Sprintf("http://127.0.0.1%s", common.Conf.Errlog)
	resp, err := http.Get(uri)
	if err != nil {
		logger.ERR(logger.ERROR, "get hc fail %v", err)
		return
	}

	defer resp.Body.Close()

	info.Msg.Errlog, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ERR(logger.ERROR, "read errlog fail %v", err)
		return
	}

	if len(info.Msg.Errlog) < 8 {
		return
	}
}

func (info *Info) Hc() {

	uri := fmt.Sprintf("http://127.0.0.1%s", common.Conf.Health)

	resp, err := http.Get(uri)
	if err != nil {
		fmt.Printf("get hc fail %v", err)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ERR(logger.ERROR, "read hc fail 1 %v", err)
		return
	}

	var hc map[string]interface{}
	err = json.Unmarshal(body, &hc)
	if err != nil {
		logger.ERR(logger.ERROR, "read hc fail 2 %v, %s", err, body)
		return
	}

	info.Msg.Hc = hc
}

func (info *Info) Send() {

	info.File() // file modtime info

	info.Cpu() // cpu info

	info.Mem() // mem info

	info.Net() // net info

	info.Disk() // disk info

	info.RockPro() //

	info.Hc() //hc

	info.Errlog() //errlog

	session := info.Client.Session

	info.Msg.Pri = session.LocalAddr().String()

	var result int
	stat := session.Call(info.URI, info.Msg, &result).Status()
	if !stat.OK() {
		logger.ERR(logger.DEBUG, "%v", stat)
	}
}
