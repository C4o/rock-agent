package common

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var C string

type Config struct {
	Rock     string `yaml:"rock"`
	Pid      string `yaml:"pid"`
	UUID     string `yaml:"uuid"`
	Dir      string `yaml:"dir"`
	Resolver string `yaml:"resolver"`
	Version  string `yaml:"version"`
	Error    struct {
		Level int    `yaml:"level"`
		Dir   string `yaml:"dir"`
	} `yaml:"error"`
	Remote struct {
		Public  string `yaml:"public"`
		Private string `yaml:"private"`
		Host    string `yaml:"host"`
	} `yaml:"remote"`
	Collectd []string `yaml:"collectd"`
	Errlog   string   `yaml: errlog`
	Infolog  string   `yaml:"infolog"`
	Health   string   `yaml:"health"`
	SSL      struct {
		Key string `yaml:"key"`
		Pem string `yaml:"pem"`
	} `yaml:"ssl"`
	Debug       int      `yaml:"debug"`
	Num         int      `yaml:"num"`
	Thread      int      `yaml:"thread"`
	KafkaStatus bool     `yaml:"kafkastatus"`
	KafkaAddr   []string `yaml:"kafkaaddr"`
	KafkaTopic  string   `yaml:"kafkatopic"`
	Access      string   `yaml:"access"`
	Limit       int      `yaml:limit`
}

type Server struct {
	Pid      string `yaml:"pid"`
	Resolver string `yaml:"resolver"`
	Version  string `yaml:"version"`
	Public   string `yaml:"public"`
	Private  string `yaml:"private"`
	Host     string `yaml:"host"`
	Error    struct {
		Level int    `yaml:"level"`
		Dir   string `yaml:"dir"`
	} `yaml:"error"`
	Errlog      string   `yaml: errlog`
	Infolog     string   `yaml:"infolog"`
	Health      string   `yaml:"health"`
	Num         int      `yaml:"num"`
	Thread      int      `yaml:"thread"`
	KafkaStatus bool     `yaml:"kafkastatus"`
	KafkaAddr   []string `yaml:"kafkaaddr"`
	KafkaTopic  string   `yaml:"kafkatopic"`
	Access      string   `yaml:"access"`
	Limit       int      `yaml:limit`
}

func LoadConf() error {

	conf := new(Config)
	file, err := ioutil.ReadFile(C)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, conf)
	if err != nil {
		return err
	}
	Conf = *conf
	return nil
}

func (server *Server) ConfToFile() error {

	var f *os.File
	var confNew []byte
	var err error

	if err := LoadConf(); err != nil {
		ERR(DEBUG, "LoadConf ERROR.")
	}
	if server.Pid != "" {
		Conf.Pid = server.Pid
	}
	if server.Version != "" {
		Conf.Version = server.Public
	}
	if server.Resolver != "" {
		Conf.Resolver = server.Resolver
	}

	if server.Public != "" {
		Conf.Remote.Public = server.Public
	}
	if server.Private != "" {
		Conf.Remote.Private = server.Private
	}
	if server.Host != "" {
		Conf.Remote.Host = server.Host
	}
	if server.Error.Dir != "" {
		Conf.Error = server.Error
	}
	if server.Errlog != "" {
		Conf.Errlog = server.Errlog
	}
	if server.Infolog != "" {
		Conf.Infolog = server.Infolog
	}
	if server.Health != "" {
		Conf.Health = server.Health
	}
	if server.Num != 0 {
		Conf.Num = server.Num
	}
	if server.Thread != 0 {
		Conf.Thread = server.Thread
	}
	if server.KafkaStatus == true || server.KafkaStatus == false {
		Conf.KafkaStatus = server.KafkaStatus
	}
	if len(server.KafkaAddr) != 0 {
		Conf.KafkaAddr = server.KafkaAddr
	}
	if len(server.KafkaTopic) != 0 {
		Conf.KafkaTopic = server.KafkaTopic
	}
	if len(server.Access) != 0 {
		Conf.Access = server.Access
	}
	if server.Limit != 0 {
		Conf.Limit = server.Limit
	}

	// 备份原配置
	if err = os.Rename(C, C+".bak"); err != nil {
		return err
	}

	// 生成新配置
	if f, err = os.OpenFile(C, os.O_CREATE|os.O_WRONLY, 0600); err == nil {
		if confNew, err = yaml.Marshal(Conf); err != nil {
			return err
		}

		if _, err = f.Write(confNew); err != nil {
			return err
		}
		defer f.Close()
	} else {
		return err
	}

	return nil
}
