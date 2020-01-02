package business

import (
	"io"
	"time"

	"github.com/mkideal/log"
	"github.com/weikaishio/sim/pb/im_service"
)

type IMBiz struct {
	newMsg chan *im_service.MsgModel
	revMsg chan *im_service.MsgModel
}

func NewIMBiz() *IMBiz {
	return &IMBiz{
		newMsg: make(chan *im_service.MsgModel, maxMsgChanSize),
		revMsg: make(chan *im_service.MsgModel, maxMsgChanSize),
	}
}

func (i *IMBiz) NewMsg(stream im_service.Im_NewMsgServer) error {
	ctx := stream.Context()
	go func() {
		for msg := range i.newMsg {
			select {
			case <-ctx.Done():
				log.Warn("the sent go routinue : client close conn by context, err:%v", ctx.Err())
				return
			default:
				err := stream.Send(msg)
				if err != nil {
					log.Error("stream.Send(%v) err:%v", msg, err)
				}
			}
		}
	}()
	go func() {
		tick := time.Tick(15 * time.Second)
		for {
			select {
			case <-tick:
				i.newMsg <- &im_service.MsgModel{
					GateId:     0,
					Uid:        1,
					MsgType:    1,
					MsgContent: "from im_service content",
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
			msg, err := stream.Recv()
			if err == io.EOF {
				log.Warn("client stream end")
				return nil
			} else if err != nil {
				log.Error("rev msg:%v", msg)
				break
			} else {
				log.Info("rev msg:%v", msg)
				i.revMsg <- msg
			}
		}
	}
}
