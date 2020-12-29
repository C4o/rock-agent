package collect

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"github.com/C4o/rock-agent/common"
	"github.com/C4o/rock-agent/logger"
	"github.com/C4o/rock-agent/tcp"
	"time"
)

type Message struct {
	Rock   map[string]string // name: string  stat: hash
	Pool   map[string]string // name: string  stat: hash
	Resty  map[string]string // name: string  stat: hash
	Server map[string]string //name: string  stat: hash
	SSL    map[string]string //name: string  stat: hash
	RC     string
}

type Frame struct {
	URI    string      // server path
	Client *tcp.Client // send
	Msg    *Message
}

func Hash(path string) string {
	txt, err := ioutil.ReadFile(path)
	if err != nil {
		return "null"
	}

	txt = append(txt, "em.sec.140986"...)

	return fmt.Sprintf("%x", md5.Sum(txt))
}

func (frame *Frame) pool() error {
	path := fmt.Sprintf("%s/pool.d", common.Conf.Rock)
	if path == "" {
		return errors.New("not found")
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	var name string
	for _, file := range files {
		name = file.Name()
		frame.Msg.Pool[name] = Hash(fmt.Sprintf("%s/%s", path, name))
	}

	return nil
}

func (frame *Frame) resty() error {
	path := fmt.Sprintf("%s/resty.d", common.Conf.Rock)
	if path == "" {
		return errors.New("not found")
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	var name string
	for _, file := range files {
		name = file.Name()
		frame.Msg.Resty[name] = Hash(fmt.Sprintf("%s/%s", path, name))
	}

	return nil
}

func (frame *Frame) conf() error {
	frame.Msg.RC = Hash(fmt.Sprintf("%s/conf/rock.conf", common.Conf.Rock))

	ssl := fmt.Sprintf("%s/conf/ssl.d", common.Conf.Rock)
	if ssl == "" {
		return errors.New("ssl.d not found")
	}

	files, err := ioutil.ReadDir(ssl)
	if err != nil {
		return err
	}

	var name string
	for _, file := range files {
		name = file.Name()
		frame.Msg.SSL[name] = Hash(fmt.Sprintf("%s/%s", ssl, name))
	}

	server := fmt.Sprintf("%s/conf/server.d", common.Conf.Rock)
	if server == "" {
		return errors.New("server.d not found")
	}

	files, err = ioutil.ReadDir(server)
	if err != nil {
		return err
	}

	for _, file := range files {
		name = file.Name()
		frame.Msg.Server[name] = Hash(fmt.Sprintf("%s/%s", server, name))
	}

	return nil
}

func (frame *Frame) rock() error {
	frame.Msg.Rock["app"] = Hash(fmt.Sprintf("%s/rock.d/app", common.Conf.Rock))
	frame.Msg.Rock["index"] = Hash(fmt.Sprintf("%s/rock.d/index", common.Conf.Rock))
	frame.Msg.Rock["balancer"] = Hash(fmt.Sprintf("%s/rock.d/balancer", common.Conf.Rock))
	frame.Msg.Rock["rewrite"] = Hash(fmt.Sprintf("%s/rock.d/rewrite", common.Conf.Rock))
	frame.Msg.Rock["security"] = Hash(fmt.Sprintf("%s/rock.d/security", common.Conf.Rock))
	return nil
}

func (frame *Frame) Send() {
	frame.Msg = &Message{
		Rock:   make(map[string]string),
		Resty:  make(map[string]string),
		Pool:   make(map[string]string),
		Server: make(map[string]string),
		SSL:    make(map[string]string),
		RC:     "",
	}

	if err := frame.rock(); err != nil {
		logger.ERR(logger.ERROR, "%v", err)
		return
	}

	if err := frame.resty(); err != nil {
		logger.ERR(logger.ERROR, "%v", err)
		return
	}

	if err := frame.pool(); err != nil {
		logger.ERR(logger.ERROR, "%v", err)
		return
	}

	if err := frame.conf(); err != nil {
		logger.ERR(logger.ERROR, "%v", err)
		return
	}

	client := frame.Client

	if !client.Status {
		logger.ERR(logger.ERROR, "send rock stat fail tcp inactive")
		return
	}

	var result int
	stat := client.Session.Call(frame.URI, frame.Msg, &result).Status()
	if !stat.OK() {
		logger.ERR(logger.ERROR, "%v", stat)
		return
	}

	//logger.ERR(logger.ERROR, "send successful %v" , frame)
}

func Do(client *tcp.Client) {
	s1 := time.NewTicker(1 * time.Second)
	s30 := time.NewTicker(30 * time.Second)
	defer func() {
		s1.Stop()
		s30.Stop()
	}()

	stat := Frame{Client: client, URI: "/rock/stat"}
	info := Info{Client: client, URI: "/rock/info", Status: false, Msg: &IMessage{}}
	status := Status{Client: client, URI: "/rock/status"}
	access := KafkaAccess{Chan: make(chan string, 4096)}

	access.Thread()
	//系统启动时就执行一次发送日志
	logger.ERR(logger.DEBUG, "KafkaStatus is %+v.", common.Conf.KafkaStatus)

	go access.Start()

	for {
		select {
		case <-s30.C:
			go info.Send()
		case <-s1.C:
			status.Send()
			stat.Send()
		}
	}
}
