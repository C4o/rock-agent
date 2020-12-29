package collect

import (
	//"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"github.com/C4o/rock-agent/common"
	"github.com/C4o/rock-agent/logger"
	"time"

	"github.com/Shopify/sarama"
	"golang.org/x/time/rate"
)

type KafkaAccess struct {
	File  *os.File
	Size  int64
	Path  string
	Chan  chan string
	First int64
	Last  int64 // 上次发送信息
	Count int   // 数据总量
}

var (
	Address []string
	Topic   string
	Num     int
	Thread  int
)

func (access *KafkaAccess) Send(data []*sarama.ProducerMessage, client sarama.SyncProducer) {

	//发送消息
	err := client.SendMessages(data)
	if err != nil {
		for _, e := range err.(sarama.ProducerErrors) {
			logger.ERR(logger.ERROR, "Write to kafka failed: %v", e)
		}
		logger.ERR(logger.ERROR, "send message failed: %v", err)
		return
	}
	//logger.ERR(logger.DEBUG, "Send %d data to kafka in %d", len(data), time.Now().Unix())
}

// 读取channel并发送
func (access *KafkaAccess) RChan(id int) {

	logger.ERR(logger.DEBUG, "id is : %d\n", id)
	//kafka初始化配置
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	//设置keepalive时间
	//config.Net.KeepAlive = 60 * time.Second
	//新增lz4压缩方式
	config.Producer.Compression = sarama.CompressionGZIP
	//生产者
	client, err := sarama.NewSyncProducer(Address, config)
	if err != nil {
		logger.ERR(logger.ERROR, "producer close,err: %v", err)
		return
	}

	defer client.Close()
	s10 := time.NewTicker(20 * time.Second)
	defer func() {
		s10.Stop()
	}()

	// 可从配置文件里读

	count := 0
	//count1 := 0
	access.Last = time.Now().Unix()
	var line string
	data := make([]*sarama.ProducerMessage, Num)

	// rate.Limiter有两个参数，第一个参数为每秒发生多少次事件，第二个参数是其缓存最大可存多少个事件
	l := rate.NewLimiter(rate.Limit(common.Conf.Limit/common.Conf.Thread), 1000)
	c, _ := context.WithCancel(context.TODO())
	for {

		l.Wait(c)
		select {
		case line = <-access.Chan:
			//count1 += 1
			msg := &sarama.ProducerMessage{}
			msg.Topic = Topic
			msg.Value = sarama.StringEncoder(line)
			//data = append(data, msg)
			data[count] = msg
			count += 1

			if count == Num {
				//logger.ERR(logger.ERROR, " main %d  data, time : %d, id : %d", count, time.Now().Unix(), id)
				access.Count += count
				count = 0
				access.Send(data, client)
				data = make([]*sarama.ProducerMessage, Num)

			}

		case <-s10.C:
			//log.Println(count1)
			//logger.ERR(logger.ERROR, "id : %d , left : %d", id, count)
			//logger.ERR(logger.ERROR, "read chan : %d in id : %d", count1, id)
			if count != 0 {
				tmp := count
				count = 0
				//fmt.Printf("%d\n", tmp)
				access.Send(data[:tmp], client)
				access.Count += tmp
				data = make([]*sarama.ProducerMessage, Num)
				access.Last = time.Now().Unix()
			}
		}
	}
}

// 读取文件内容并发送到channel
//func (access *KafkaAccess) Stat() {

//var err error
//var fileinfo os.FileInfo
//access.File, _ = os.Open(access.Path)
//defer access.File.Close()

//if fileinfo, err = os.Stat(access.Path); err == nil {
//if fileinfo.Size() != access.Size {

//count := 0

//scanner := bufio.NewScanner(access.File)
//for scanner.Scan() {
//access.Chan <- scanner.Text()
//count += 1
//}

////logger.ERR(logger.DEBUG, "len of file : %d", count)

//access.Size = fileinfo.Size()
//// 发送完成日志之后修改成.log的格式以后续压缩
//newPath := strings.Split(access.Path, ".log")[0] + ".log"
//if err = os.Rename(access.Path, newPath); err != nil {
//logger.ERR(logger.ERROR, "rename log file : %v", err)
//}
//logger.ERR(logger.DEBUG, "send %s Successfully.", access.Path)
//}
//} else {
//logger.ERR(logger.ERROR, "Get File Stat ERROR : %v", err)
//}
//}

// 判断topic是否存在
func (access *KafkaAccess) Topic() error {

	var err error
	var cli sarama.Client
	var topics []string

	cli, err = sarama.NewClient(Address, sarama.NewConfig())
	defer func() {
		err = cli.Close()
		if err != nil {
			logger.ERR(logger.ERROR, "close client of kafka error : %v", err)
		}
	}()
	if err != nil {
		logger.ERR(logger.ERROR, "create new client err : %v", err)
		return err
	}
	topics, err = cli.Topics()
	if err != nil {
		logger.ERR(logger.ERROR, "get topics err : %v", err)
		return err
	}
	for _, t := range topics {
		if t == Topic {
			return nil
		}
	}
	return errors.New("no topic : " + Topic)
}

// 启动线程
func (access *KafkaAccess) Thread() {

	Thread = common.Conf.Thread

	logger.ERR(logger.DEBUG, "%d new thread start.", Thread)
	for i := 0; i < Thread; i++ {
		go access.RChan(i)
	}
}

// 通过配置文件来获取日志路径生成日志位置
func (access *KafkaAccess) GetPath() {

	if common.Conf.Access == "" {
		s10 := time.NewTicker(10 * time.Second)
		defer func() {
			s10.Stop()
		}()
		for {
			select {
			case <-s10.C:
				if common.Conf.Access != "" {
					break
				}
				logger.ERR(logger.ERROR, "access_log path not found in configure file.")
			}
		}
	}

	// 比如现在11月09日19点07，那就读11-09.(19-1).log.ts
	access.Path = fmt.Sprintf("%saccess.%s.log.ts", GenLog(common.Conf.Access), time.Now().Add(-(3600 * time.Second)).Format("2006-01-02.15"))

}

// 初始化
func (access *KafkaAccess) Start() {

	Address = common.Conf.KafkaAddr
	Topic = common.Conf.KafkaTopic
	Num = common.Conf.Num

	var err error
	s5 := time.NewTicker(5 * time.Second)
	s10 := time.NewTicker(10 * time.Second)
	defer func() {
		s5.Stop()
		s10.Stop()
	}()

	if err = access.Topic(); err != nil {
		for {
			select {
			case <-s10.C:
				if err = access.Topic(); err == nil {
					break
				}
				logger.ERR(logger.ERROR, "kafka get topic err, %v", err)
			}
		}
	}

	access.First = time.Now().Unix()
	access.Last = time.Now().Unix()
	access.Count = 0

	// 获取日志路径并
	access.GetPath()
	//access.Stat()

	// 初始化一个logtail的对象
	lt := LogTail{
		SignalChan:     make(chan os.Signal),
		OffsetFileName: "/etc/rock/agent/offset.restore",
	}

	// 生成要读取的日志路径和生成offset
	lt.SetLogFileName()

	// 监听系统信号
	signal.Notify(lt.SignalChan)

	go lt.SignalCatch()

RESTART:

	if common.Conf.KafkaStatus {
		// 配置更新重新设定
		lt.Tail(access.Chan)
	}

	<-s5.C
	lt.SetLogFileName()
	goto RESTART

}
