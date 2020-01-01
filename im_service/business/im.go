package business

import (
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
	//context from stream
	go func() {
		for msg := range i.newMsg {
			log.Info("new send msg:%v", msg)
			err := stream.Send(msg)
			if err != nil {
				log.Error("stream.Send(%v) err:%v", msg, err)
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
		msg, err := stream.Recv()
		if err != nil {
			log.Error("rev msg:%v", msg)
			break
		} else {
			log.Info("rev msg:%v", msg)
			i.revMsg <- msg
		}
	}
	return nil
}
