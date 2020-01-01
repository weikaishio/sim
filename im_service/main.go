package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/mkideal/log"
	"github.com/weikaishio/sim/common/osutil/pid"
	"github.com/weikaishio/sim/im_service/business"
	"github.com/weikaishio/sim/im_service/config"
)

var (
	svr *business.Server
)

func main() {
	var err error
	err = config.Init()
	if err != nil {
		log.Fatal("config.Init() err:%v", err)
	}

	if err := pid.New(config.PidFile); err != nil {
		log.Fatal("pid.New(%s),err:%v", config.PidFile, err)
	}
	defer func() {
		_ = pid.Remove(config.PidFile)
		defer log.Uninit(log.InitFile("./log/sim.log"))
	}()

	reload()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR1)
	for {
		s := <-c
		log.Info("sim get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			svr.Quit()
			//client连接状态置离线
			time.Sleep(time.Second)
			log.Warn("sim exit")
			return
		case syscall.SIGHUP:
			reload()
		case syscall.SIGUSR1:
			filename := filepath.Base(os.Args[0]) + ".dump"
			dumpOut, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
			if err == nil {
				for _, name := range []string{"goroutine", "heap", "block"} {
					p := pprof.Lookup(name)
					if p == nil {
						continue
					}
					name = strings.ToUpper(name)
					fmt.Fprintf(dumpOut, "-----BEGIN %s-----\n", name)
					p.WriteTo(dumpOut, 2)
					fmt.Fprintf(dumpOut, "\n-----END %s-----\n", name)
				}
				dumpOut.Close()
			}
		default:
			log.Error("unhandled signal:%v", s.String())
			return
		}
	}
}

func reload() {
	log.Warn("reload")
	if svr != nil {
		svr.Quit()
		time.Sleep(time.Second)
	}
	var err error
	err = config.Init()
	if err != nil {
		log.Fatal("config.Init() err:%v", err)
	} else {
		log.Info("cfg:%v", config.Conf)
	}
	//log.SetLevelFromString(cfg.LogLevel)

	if svr == nil {
		svr = business.NewServer()
	}
	svr.Start()
}
