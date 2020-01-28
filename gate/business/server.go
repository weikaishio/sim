package business

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/weikaishio/distributed_lib/buffer"
	"github.com/weikaishio/sim/codec"
	"github.com/weikaishio/sim/common/netutil"
	"github.com/weikaishio/sim/gate/config"

	"github.com/mkideal/log"
)

type Server struct {
	seq      int32
	listener net.Listener
	httpSvr  *http.Server
	mux      *http.ServeMux
	once     sync.Once

	bufPool *buffer.Pool

	clients       map[uint32]Client
	clientsLocker sync.RWMutex
	running       int32
}

func NewServer() *Server {
	return &Server{
		clients: map[uint32]Client{},
		bufPool: buffer.NewPool(bufPoolSize, bufMaxPoolSize),
	}
}
func (s *Server) isRunning() bool {
	return atomic.LoadInt32(&s.running) != 0
}
func (s *Server) Start() {
	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		log.Error("Server already start")
		return
	}
	log.Info("svr begin Start")
	s.beginListen()
	err := RPCIMSvrSharedInstance().Init(config.Conf.Etcd, config.Conf.GrpcImCli)
	if err != nil {
		panic(fmt.Sprintf("RPCIMSvrSharedInstance Init err:%v", err))
	}
	go func() {
		for {
			err := RPCIMSvrSharedInstance().MsgCommunication()
			if err != nil {
				log.Error("MsgCommunication err:%v", err)
				RPCIMSvrSharedInstance().retryWait += 3 * time.Second
			}
			time.Sleep(RPCIMSvrSharedInstance().retryWait)
		}
	}()

	log.Info("svr finish Start")
}

func (s *Server) stopListen() {
	if s.httpSvr != nil {
		_ = s.httpSvr.Close()
		log.Info("s.httpSvr Close")
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
}
func (s *Server) makeNewSeq() int32 {
	atomic.AddInt32(&s.seq, 1)
	return s.seq
}
func (s *Server) beginListen() {
	addr := fmt.Sprintf("%s:%d", config.Conf.Net.Host, config.Conf.Net.Port)
	if listener, err := netutil.ListenAndServeTCP(addr, s.handleClientConn, true); err != nil {
		log.Fatal("server start listen:%s,err:%v", addr, err)
	} else {
		s.listener = listener
	}
	log.Trace("beginListen addr:%s", addr)
	go func() {
		s.once.Do(func() {
			s.mux = http.NewServeMux()
			s.mux.HandleFunc("/monitor", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			s.mux.HandleFunc("/online", func(w http.ResponseWriter, r *http.Request) {
				var clientAry []string
				s.clientsLocker.RLock()
				for mid := range s.clients {
					clientAry = append(clientAry, strconv.Itoa(int(mid)))
				}
				s.clientsLocker.RUnlock()
				_, _ = w.Write([]byte(strings.Join(clientAry, ",")))
				w.WriteHeader(http.StatusOK)
			})
			//s.mux.HandleFunc("/debug/pprof/", pprof.Index)
			//s.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			//s.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			//s.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			//s.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		})
		s.httpSvr = &http.Server{
			Addr:           fmt.Sprintf(":%d", config.Conf.Net.Port+1),
			Handler:        s.mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		log.Error("http listen err:%v", s.httpSvr.ListenAndServe())
	}()
}
func (s *Server) handleClientConn(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	if !s.isRunning() {
		_ = conn.Close()
		log.Warn("handleClientConn !s.isRunning() conn:%s.Close", remoteAddr)
		return
	}
	client := newClientCodec(s, uint32(s.makeNewSeq()), remoteAddr)
	uniqueId := fmt.Sprintf("m%d_%s", client.uniqueId, remoteAddr)
	log.Trace("handleClientConn remoteAddr:%s,uniqueId:%s", remoteAddr, uniqueId)
	readStream := netutil.NewReadStream(conn, client.onRecvPacket)
	readStream.SetMagic([]byte(codec.MAGIC))
	readStream.SetTimeout(time.Duration(config.Conf.Net.KeepaliveTime) * time.Second)
	session := netutil.NewRWSession(conn, readStream, uniqueId, maxConWriteSize)
	client.session = session
	session.Run(client.onNewSession, client.onQuitSession)
	session.WSession = nil
	s.removeClient(client)
}
func (s *Server) removeClient(client Client) {
	log.Trace("removeClient:%v", client)
	key := client.GetId()
	s.clientsLocker.Lock()
	if oldClient, ok := s.clients[key]; ok && oldClient.GetUniqueId() == client.GetUniqueId() {
		delete(s.clients, key)
	}
	s.clientsLocker.Unlock()
	log.Info("removeClient:%v, success", client)
}
func (s *Server) quitAllClient() {
	s.clientsLocker.Lock()
	defer s.clientsLocker.Unlock()
	ids := make([]string, 0)
	for _, m := range s.clients {
		m.GetSession().Quit(false)
		ids = append(ids, fmt.Sprintf("%d",m.GetId()))
	}
}
func (s *Server) Quit() {
	log.Info("svr begin Quit")

	s.quitAllClient()
	atomic.StoreInt32(&s.running, 0)
	s.stopListen()
	log.Info("svr finish Quit")
}
