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
	"github.com/weikaishio/distributed_lib/etcd"
	"github.com/weikaishio/sim/im_service/config"
	"github.com/weikaishio/sim/pb/im_service"
	"google.golang.org/grpc"

	"github.com/mkideal/log"
)

type Server struct {
	seq       int32
	httpSvr   *http.Server
	discovery *etcd.Discovery
	mux       *http.ServeMux
	once      sync.Once

	bufPool *buffer.Pool

	clients       map[uint32]*Client
	clientsLocker sync.RWMutex
	running       int32
}

func NewServer() *Server {
	return &Server{
		clients: map[uint32]*Client{},
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

	log.Info("svr finish Start")
}
func (s *Server) stopListen() {
	if s.httpSvr != nil {
		_ = s.httpSvr.Close()
		log.Info("s.httpSvr Close")
	}
}
func (s *Server) makeNewSeq() int32 {
	atomic.AddInt32(&s.seq, 1)
	return s.seq
}
func (s *Server) RPCSvrRun(auths map[string]string, certFile, keyFile string) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Conf.GrpcSvr.Port))
	if err != nil {
		log.Error("RPCSvrRun net.Listen err:%v", err)
		return err
	}
	//auth := grpc_server.NewRPCSecurity(auths, certFile, keyFile)
	rpcSvr := grpc.NewServer(
		//grpc.Creds(auth.Creds),
		//grpc.UnaryInterceptor(grpc.UnaryServerInterceptor(auth.UnaryInterceptor)),
	)
	svr := NewIMBiz()
	im_service.RegisterImServer(rpcSvr, im_service.ImServer(svr))
	log.Info("gRPC Run")
	s.discovery.Register(config.Conf.GrpcSvr.Target, config.Conf.GrpcSvr.Host, config.Conf.GrpcSvr.Port,
		60*time.Second, "im service gRPC")
	err = rpcSvr.Serve(listen)
	if err != nil {
		log.Error("RPCSvrRun svr.Serve err:%v", err)
	}
	return err
}
func (s *Server) InitDiscovery(endpoints []string, username, password string) error {
	var err error
	s.discovery, err = etcd.NewDiscovery(endpoints, username, password)
	if err != nil {
		log.Error("etcd.NewDiscovery err:%v", err)

	}
	return err
}
func (s *Server) RegForDiscovery(serviceName string, host string, port int) {
	s.discovery.Register(serviceName, host, port, 60*time.Second, "wf gRPC")
}
func (s *Server) UnRegForDiscovery(serviceName string, host string, port int) {
	if s.discovery != nil {
		log.Info("quit discover")
		s.discovery.UnRegister(serviceName, host, port)
	}
}
func (s *Server) beginListen() {
	err := s.InitDiscovery(config.Conf.Etcd.EndpointAry(), config.Conf.Etcd.Username, config.Conf.Etcd.Password)
	if err != nil {
		log.Fatal("etcd.NewDiscovery err:%v", err)
	}
	go func() {
		err := s.RPCSvrRun(config.Conf.GrpcAuths.AuthMap(), config.Conf.GrpcSvr.CertFile, config.Conf.GrpcSvr.KeyFile)
		if err != nil {
			log.Error("RPCSvrRun err:%v", err)
		}
	}()
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
			Addr:           fmt.Sprintf(":%d", config.Conf.GrpcSvr.Port+1),
			Handler:        s.mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		log.Error("http listen err:%v", s.httpSvr.ListenAndServe())
	}()
}

func (s *Server) Quit() {
	log.Info("svr begin Quit")
	atomic.StoreInt32(&s.running, 0)
	s.stopListen()
	log.Info("svr finish Quit")
}
