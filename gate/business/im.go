package business

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/mkideal/log"
	"github.com/weikaishio/distributed_lib/etcd"
	"github.com/weikaishio/distributed_lib/grpc_pool"
	"github.com/weikaishio/sim/common/confutil"
	"github.com/weikaishio/sim/pb/im_service"
	"google.golang.org/grpc"
)

type IMRPC struct {
	rpcPool   *grpc_pool.RPCPool
	newMsg    chan *im_service.MsgModel
	revMsg    chan *im_service.MsgModel
	retryWait time.Duration
}

var (
	sharedInstance *IMRPC
	once           sync.Once
)

func RPCIMSvrSharedInstance() *IMRPC {
	once.Do(func() {
		sharedInstance = &IMRPC{}
	})
	return sharedInstance
}

func (p *IMRPC) Init(etcdConf *confutil.Etcd, grpcImCliConf *confutil.GRPCCliConf) error {
	discover, err := etcd.NewDiscovery(etcdConf.EndpointAry(), etcdConf.Username, etcdConf.Password)
	if err != nil {
		log.Error("etcd.NewDiscovery p.etcdEndPoints:%v,err:%v", etcdConf.EndpointAry(), err)
		return err
	}
	makeConn := func() (grpc_pool.Connection, error) {
		return discover.DialWithAuth(grpcImCliConf.Target, grpcImCliConf.AuthName, grpcImCliConf.AuthPwd, grpcImCliConf.CertServerName, grpcImCliConf.CertFile)
	}
	p.rpcPool = grpc_pool.NewRPCPool(makeConn)
	p.rpcPool.InitPool()
	p.newMsg = make(chan *im_service.MsgModel, maxMsgChanSize)
	p.revMsg = make(chan *im_service.MsgModel, maxMsgChanSize)
	return nil
}
func (p *IMRPC) Close() {
	p.rpcPool.Close()
}

func (p *IMRPC) MsgCommunication() error {
	conn := p.rpcPool.Borrow()
	if conn == nil {
		return errors.New("执行出错，无可用连接")
	} else {
		defer p.rpcPool.Return(conn)
	}

	cli := im_service.NewImClient(conn.(*grpc.ClientConn))
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	stream, err := cli.NewMsg(context.Background())
	if err != nil {
		return err
	} else {
		p.retryWait = 0
	}
	ctx := stream.Context()
	go func() {
		for msg := range p.newMsg {
			select {
			case <-ctx.Done():
				log.Warn("the sent go routinue : client close conn by context, err:%v", ctx.Err())
				return
			default:
				_ = stream.Send(msg)
			}
		}
	}()
	go func() {
		tick := time.Tick(15 * time.Second)
		for {
			select {
			case <-tick:
				p.newMsg <- &im_service.MsgModel{
					GateId:     1,
					Uid:        2,
					MsgType:    1,
					MsgContent: "from gate content",
				}
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			log.Warn("client close conn by context, err:%v", ctx.Err())
			return ctx.Err()
		default:
			//stream.Trailer()
			msg, err := stream.Recv()
			if err == io.EOF {
				log.Warn("client stream end")
				return err
			} else if err != nil {
				log.Error("rev msg:%v", msg)
				break
			} else {
				log.Info("rev msg:%v", msg)
				p.revMsg <- msg
			}
		}
	}
}
