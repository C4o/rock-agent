package common

import (
	//"errors"
	//"io/ioutil"
	"net"
	"time"
)

var (
	Conf   Config
	Info   string
	HCInfo string
	Errlog string
	Err    string
)

func dial(remote string) bool {
	if remote == "" {
		return false
	}
	conn, err := net.DialTimeout("tcp", remote, time.Second*3)
	if err != nil {
		return false
	}

	conn.Close()

	return true

}

func Net() string {

	if dial(Conf.Remote.Private) {
		return Conf.Remote.Private
	} else {
		ERR(ERROR, "remote Private %s conn fail", Conf.Remote.Private)
	}

	if dial(Conf.Remote.Public) {
		return Conf.Remote.Public
	} else {
		ERR(ERROR, "remote Public %s conn fail", Conf.Remote.Public)
	}

	if dial(Conf.Remote.Host) {
		return Conf.Remote.Host
	} else {
		ERR(ERROR, "remote Host %s conn fail", Conf.Remote.Host)
	}

	return ""
}

// unused function
//func Cert() (string, string, error) {

//var err error
//var pem, key []byte
//ERR(DEBUG, "log ssl %v", Conf)
//if pem, err = ioutil.ReadFile(Conf.SSL.Pem); err == nil {
//if key, err = ioutil.ReadFile(Conf.SSL.Key); err == nil {
//return string(key), string(pem), nil
//}
//return "", "", errors.New("key file open error.")
//}
//return "", "", errors.New("pem file open error.")
//}

func New(conf string) {
	C = conf
	if err := LoadConf(); err != nil {
		ERR(DEBUG, "LoadConf ERROR.")
	}
}
