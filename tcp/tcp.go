package tcp

import (
	//"context"
	//"errors"
	"github.com/C4o/rock-agent/common"
	"github.com/C4o/rock-agent/logger"
	"time"

	tp "github.com/henrylee2cn/erpc/v6"
	//tp "github.com/henrylee2cn/teleport"
)

var Cli *Client

type Client struct {
	Session tp.Session
	Peer    tp.Peer
	Remote  string
	//是否初始化成功
	Status bool
}

func New(peer tp.Peer) error {

	remote := common.Net()
	if remote == "" {
		time.Sleep(3 * time.Second)
	} else {
		Cli = &Client{Peer: peer, Status: false, Remote: remote}
	}
	Cli.Conn()

	return nil
}

func (c *Client) Conn() {

	for {
		// 探测时，探测了几次就会出现几个未连接的状态
		remote := common.Net()
		if remote == "" {
			time.Sleep(3 * time.Second)
		} else {
			session, err := c.Peer.Dial(remote)
			if err == nil {
				c.Remote = remote
				c.Session = session
				c.Status = true
				logger.ERR(logger.INFO, "session conn successful")
				return
			} else {
				logger.ERR(logger.ERROR, "dial remote %s error: %v", remote, err)
				time.Sleep(3 * time.Second)
			}

		}
	}
}

func (c *Client) Health() {

	logger.ERR(logger.ERROR, " time : %d", time.Now().Unix())
	tk := time.NewTicker(5 * time.Second)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			//hc := c.Session.Health()
			//logger.ERR(logger.ERROR, " res of health : %v, c.status: %v", hc, c.RStatus)
			if !c.Session.Health() {
				c.Status = false
				logger.ERR(logger.INFO, "tcp session down")
				c.Conn()
			}
		}
	}
}

func (c *Client) Stop() {
	c.Status = false
	tp.FlushLogger()
	c.Peer.Close()
}
