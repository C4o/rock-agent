package collect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"github.com/C4o/rock-agent/common"
	"github.com/C4o/rock-agent/logger"
	"github.com/C4o/rock-agent/tcp"
	"time"
)

type Statusd struct {
	A uint64 `json:"total"`
	B uint64 `json:"deny"`
	C uint64 `json:"_200"`
	D uint64 `json:"_30x"`
	E uint64 `json:"_40x"`
	F uint64 `json:"_50x"`

	G uint64 `json:"_100"`
	H uint64 `json:"_1000"`
	I uint64 `json:"_10000"`
	J uint64 `json:"_x0000"`
}

type SMessage struct {
	App       map[string]Statusd
	Host      map[string]Statusd
	Upstream  map[string]Statusd
	TimeStamp time.Time
	Timefive  int
	Timehalf  int
	Timehour  int
}

type Status struct {
	URI    string      //sever path
	Client *tcp.Client // client
	Msg    *SMessage   //message
}

func (status *Status) Get(timefive int) {
	uri := fmt.Sprintf("http://127.0.0.1%s?time=%d", common.Conf.Infolog, timefive-300)
	resp, err := http.Get(uri)
	if err != nil {
		logger.ERR(logger.ERROR, "get info fail %v", err)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ERR(logger.ERROR, "read info fail %v", err)
		return
	}

	err = json.Unmarshal(body, status.Msg)
	if err != nil {
		logger.ERR(logger.ERROR, "json Unmarshal fail %v", err)
		return
	}
}

func (status *Status) Send() {

	nowTime := time.Now()
	if nowTime.Unix()%300 != 0 {
		return
	}

	logger.ERR(logger.DEBUG, "status requests time : %s", nowTime.Format("2006-01-02 15:04:05"))
	status.Msg = new(SMessage)

	status.Msg.TimeStamp = nowTime
	status.Msg.Timefive = int(math.Floor(float64(nowTime.Unix()/300))) * 300
	status.Msg.Timehalf = int(math.Floor(float64(nowTime.Unix()/1800))) * 1800
	status.Msg.Timehour = int(math.Floor(float64(nowTime.Unix()/3600))) * 3600
	status.Get(status.Msg.Timefive)

	session := status.Client.Session
	var result int

	stat := session.Call(status.URI, status.Msg, &result).Status()
	if !stat.OK() {
		logger.ERR(logger.DEBUG, "%v", stat)
		return
	}
}
