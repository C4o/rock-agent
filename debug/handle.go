package debug

import (
	//"bytes"
	"os"
	"os/exec"
	"github.com/C4o/rock-agent/logger"

	tp "github.com/henrylee2cn/erpc/v6"
)

const (
	ClearNgxErr  = 3306
	StopFirewall = 3389
)

type Debug struct {
	tp.PushCtx
}

func (debug *Debug) Operation(op *int) *tp.Status {

	logger.ERR(logger.DEBUG, "/debug/operation req : %d ", *op)
	var cmd *exec.Cmd
	//var stdout bytes.Buffer
	//var stderr bytes.Buffer
	var err error

	switch *op {
	case ClearNgxErr:
		if err = os.Truncate("/usr/local/rock/logs/error.log", 0); err != nil {
			logger.ERR(logger.ERROR, "error in clear ngxerrlog : %v", err)
			return nil
		}
	case StopFirewall:
		cmd = exec.Command("systemctl", "stop", "iptables", "&&", "systemctl", "stop", "firewalld")
	default:
		return nil
	}

	//cmd.Stdout = &stdout
	//cmd.Stderr = &stderr

	if err = cmd.Start(); err != nil {
		logger.ERR(logger.ERROR, " error in stop firewall : %v", err)
		return nil
	}

	//if err = cmd.Wait(); err != nil {
	//return tp.NewStatus(500, stderr.String(), err)
	//}

	return nil

}
