package main

import (
	"flag"
	"github.com/mkideal/log/logger"
	"github.com/weikaishio/sim/codec"
	"github.com/weikaishio/sim/common/osutil/pid"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mkideal/log"

	"github.com/weikaishio/sim/common/netutil"
	"strconv"
)

const (
	bufPoolSize     = 100
	bufMaxPoolSize  = 1000
	maxConWriteSize = 1000
)

var (
	machineId  = flag.Uint("mid", 1, "machine Id")
	machineKey = flag.String("mkey", "washcartestkey123", "machine key")
	svrAddr    = flag.String("svr_addr", "127.0.0.1:8910", "server addr")
	aliveTime  = flag.Int("alive_time", 180, "keep alive time(s)")
	pidFile    = flag.String("pid", "client_test.pid", "pid filepath")
	client    *Client
)

func main() {
	flag.Parse()

	log.SetLevel(logger.MustParseLevel("TRACE"))
	if err := pid.New(*pidFile); err != nil {
		log.Fatal("pid.New(%s),err:%v", pidFile, err)
	}
	defer func() {
		pid.Remove(*pidFile)
		defer log.Uninit(nil)
	}()

	client = newClient(uint32(*machineId), *machineKey)
	BeginConnect()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT)
	for {
		sig := <-c
		switch sig {
		case syscall.SIGQUIT:
			log.Info("exit")
		}
	}
}
func BeginConnect() {
	log.Info("BeginConnect")
	var conn *net.TCPConn
	var err error
	for {
		var tcpAddr *net.TCPAddr
		tcpAddr, err = net.ResolveTCPAddr("tcp", *svrAddr)
		if err != nil {
			log.Error("resolveTCPAddr err:%v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		conn, err = net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			log.Error("DialTCP err:%v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	readStream := netutil.NewReadStream(conn, client.onRecvPacket)
	//readStream.SetTimeout(time.Duration(*aliveTime) * time.Second)
	readStream.SetMagic([]byte(codec.MAGIC))
	session := netutil.NewRWSession(conn, readStream, strconv.Itoa(int(*machineId)), maxConWriteSize)
	client.session = session
	client.start()
	go session.Run(client.onBeginHandle, func() {
		log.Warn("session closed")
		client.stop()
		BeginConnect()
	})
}
