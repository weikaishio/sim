package main

import (
	"flag"
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
	"github.com/weikaishio/sim/config"
	"github.com/weikaishio/sim/server"
)

var (
	env       string
	location  string
	pidFile   string
	configDir string
	cfg       *config.Config
	svr       *server.Server
)

func init() {
	flag.StringVar(&env, "env", "local", "The environment the program should run under.")
	flag.StringVar(&location, "location", "", "Server Location. Use only if the env is not local and develop.")
	flag.StringVar(&pidFile, "pid", "sim.pid", "pid filepath")
	flag.StringVar(&configDir, "configDir", "./config/", "config path")
	flag.Parse()
}
func main() {
	flag.Parse()
	var err error
	cfg, err = config.NewConfig(configDir, env, location)
	if err != nil {
		log.Fatal("server.NewConfig(:%s) err:%v", configDir, err)
	}

	if err := pid.New(pidFile); err != nil {
		log.Fatal("pid.New(%s),err:%v", pidFile, err)
	}
	defer func() {
		pid.Remove(pidFile)
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
	cfg, err = cfg.Reload()
	if err != nil {
		log.Fatal("cfg.reload err:%v", err)
	} else {
		log.Info("cfg:%v", cfg)
	}
	log.SetLevelFromString(cfg.LogLevel)

	if svr == nil {
		svr = server.NewServer(cfg)
	} else {
		svr.RefreshCfg(cfg)
	}
	svr.Start()
}
