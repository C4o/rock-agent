package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/C4o/rock-agent/collect"
	"github.com/C4o/rock-agent/common"
	//"github.com/C4o/rock-agent/debug"
	"github.com/C4o/rock-agent/debug"
	"github.com/C4o/rock-agent/logger"
	"github.com/C4o/rock-agent/tcp"
	"strings"

	tp "github.com/henrylee2cn/erpc/v6"
	"github.com/henrylee2cn/erpc/v6/plugin/auth"
	//tp "github.com/henrylee2cn/teleport"
	//"github.com/henrylee2cn/teleport/plugin/auth"
)

var agentInfo = `[*] 更新说明

[+] 2020年5月19日更新

`

func main() {

	var err error

	v := flag.Bool("v", false, "version")
	e := flag.Bool("env", false, "env")
	c := flag.String("c", "/etc/rock/agent/conf.yaml", "configure path")
	a := flag.Bool("a", false, "agent info")

	flag.Parse()

	if *a {
		fmt.Println(agentInfo)
		return
	}

	if *v {
		fmt.Printf("rock agent version: %s\n", common.Conf.Version)
		return
	}

	if *e {
		fmt.Printf("%s\n", common.Conf.Rock)
		return
	}

	if *c == "" {
		fmt.Println("configure file path not found.")
		return
	}

	common.New(*c)

	err = logger.NewLogger()
	if err != nil {
		fmt.Printf("init logger fail %v\n", err)
		return
	}

	err = tcp.New(newPeer())
	if err != nil {
		fmt.Printf("init tcp fail %v\n", err)
		return
	}

	defer tcp.Cli.Stop()

	go tcp.Cli.Health()

	logger.ERR(logger.DEBUG, "pid for now : %d", os.Getpid())

	collect.Do(tcp.Cli)

}

func newPeer() tp.Peer {

	var verifyAuthInfo = auth.NewBearerPlugin(
		func(sess auth.Session, fn auth.SendOnce) (stat *tp.Status) {
			var ret string
			data := tcp.Data{
				Host:   common.Conf.Remote.Host,
				UUID:   common.Conf.UUID,
				IP:     strings.Split(sess.LocalAddr().String(), ":")[0],
				Status: "offline",
			}
			stat = fn(data.String(), &ret)
			if !stat.OK() {
				return
			}
			tp.Infof("auth info: %s, result: %s", common.Conf.UUID, ret)
			return
		},
		tp.WithBodyCodec('s'),
	)

	agent := new(tcp.Agent)
	tp.SetLoggerLevel("OFF")
	config := tp.PeerConfig{PrintDetail: false}
	peer := tp.NewPeer(config, verifyAuthInfo)
	e := peer.SetTLSConfigFromFile(common.Conf.SSL.Pem, common.Conf.SSL.Key, true)
	if e != nil {
		tp.Fatalf("%v", e)
	}

	peer.RoutePush(agent)
	peer.RoutePush(new(collect.Access))
	peer.RoutePush(new(debug.Debug))

	return peer

}
