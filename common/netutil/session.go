package netutil

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/mkideal/log"

	"github.com/weikaishio/distributed_lib/buffer"
)

type CrypterFunc func(dst, src []byte)

type Session interface {
	Id() string
	Send(*buffer.Buffer)
	Run(onNewSession, onQuitSession func())
	Quit(wait bool) chan struct{}
	SetEncrypter(encrypter CrypterFunc)
	SetDecrypter(decrypter CrypterFunc)
}

// 只写的Session
type WSession struct {
	conn   net.Conn
	id     string
	closed int32

	writeQuit chan struct{}
	writeChan chan *buffer.Buffer

	encryptBuf []byte

	encrypter     CrypterFunc
	encrypterChan chan CrypterFunc

	waitQuitNotify int32
	done           chan struct{}
}

func NewWSession(conn net.Conn, id string, conWriteSize int) *WSession {
	if conWriteSize <= 0 {
		conWriteSize = 4096
	}
	return &WSession{
		conn:          conn,
		id:            id,
		writeQuit:     make(chan struct{}),
		writeChan:     make(chan *buffer.Buffer, conWriteSize),
		encrypterChan: make(chan CrypterFunc, 1),
		encryptBuf:    make([]byte, 0),
		done:          make(chan struct{}),
	}
}

func (ws *WSession) Id() string                         { return ws.id }
func (ws *WSession) setClosed()                         { atomic.StoreInt32(&ws.closed, 1) }
func (ws *WSession) getClosed() bool                    { return atomic.LoadInt32(&ws.closed) == 1 }
func (ws *WSession) SetEncrypter(encrypter CrypterFunc) { ws.encrypterChan <- encrypter }

func (ws *WSession) Send(b *buffer.Buffer) {
	if b.Len() > 0 && !ws.getClosed() {
		b.Add(1)

		select {
		case ws.writeChan <- b:
		default:
			log.Error("buffer full")
			ws.setClosed()
		}
	}
}

func (ws *WSession) write(b *buffer.Buffer) (err error) {
	defer b.Done()

	_ = ws.conn.SetWriteDeadline(time.Now().Add(time.Second * 60))

	src := b.Bytes()
	byteNum4Len := DefaultByteNumForLength
	_, err = ws.conn.Write(src[:byteNum4Len])
	dataLength := len(src) - byteNum4Len
	if encrypter := ws.encrypter; encrypter != nil && b.Encrypt {
		if len(ws.encryptBuf) < dataLength {
			ws.encryptBuf = make([]byte, dataLength)
		}
		encrypter(ws.encryptBuf, src[byteNum4Len:])
		_, err = ws.conn.Write(ws.encryptBuf[:dataLength])
	} else {
		_, err = ws.conn.Write(src[byteNum4Len:])
	}
	return
}

func (ws *WSession) startWriteLoop(startWrite, endWrite chan<- struct{}) {
	startWrite <- struct{}{}
	log.Trace("WSession startWriteLoop")
	for {
		if ws.getClosed() {
			log.Trace("WSession getClosed is true")
			break
		}
		select {
		case b := <-ws.writeChan:
			n := b.Len()
			err := ws.write(b)
			if err != nil {
				log.Warn("startWriteLoop id:%d,write err:%v", ws.id, err)
				ws.setClosed()
			} else {
				log.Trace("send %d[%v] bytes to session %s", n, b.Bytes(), ws.Id())
			}
		case encrypter := <-ws.encrypterChan:
			ws.encrypter = encrypter
		case <-time.After(time.Second):
		}
	}

	remain := len(ws.writeChan)
	for i := 0; i < remain; i++ {
		b := <-ws.writeChan
		_ = ws.write(b)
	}

	_ = ws.conn.Close()
	endWrite <- struct{}{}
	log.Trace("WSession startWriteLoop end")
}

func (ws *WSession) Run(onNewSession, onQuitSession func()) {
	startWrite := make(chan struct{})
	endWrite := make(chan struct{})

	go ws.startWriteLoop(startWrite, endWrite)
	<-startWrite

	if onNewSession != nil {
		onNewSession()
	}

	<-endWrite

	if ws.conn != nil {
		_ = ws.conn.Close()
	}

	if onQuitSession != nil {
		onQuitSession()
	}
	if atomic.LoadInt32(&ws.waitQuitNotify) != 0 {
		ws.done <- struct{}{}
	}
}

func (ws *WSession) Quit(wait bool) chan struct{} {
	if wait {
		atomic.StoreInt32(&ws.waitQuitNotify, 1)
	}
	ws.setClosed()
	log.Trace("WSession setClosed")
	return ws.done
}

type RWSession struct {
	*WSession
	rstream StreamReader
}

func NewRWSession(conn net.Conn, rstream StreamReader, id string, conWriteSize int) *RWSession {
	s := new(RWSession)
	s.WSession = NewWSession(conn, id, conWriteSize)
	s.rstream = rstream
	return s
}

func (s *RWSession) SetDecrypter(decrypter CrypterFunc) {
	s.rstream.SetDecrypter(decrypter)
}

func (s *RWSession) startReadLoop(startRead, endRead chan<- struct{}) {
	log.Trace("RWSession startReadLoop")
	startRead <- struct{}{}
	for {
		_, err := s.rstream.Read()
		if err != nil {
			//if err.Error() == "EOF" {
			s.setClosed()
			//}
			log.Warn("startReadLoop id:%v,read err:%v", s.id, err)
		}
		if s.getClosed() {
			break
		}
	}
	endRead <- struct{}{}
	log.Trace("RWSession startReadLoop end")
}

func (s *RWSession) Run(onNewSession, onQuitSession func()) {
	startRead := make(chan struct{})
	startWrite := make(chan struct{})
	endRead := make(chan struct{})
	endWrite := make(chan struct{})

	go s.startReadLoop(startRead, endRead)
	go s.startWriteLoop(startWrite, endWrite)

	<-startRead
	<-startWrite

	if onNewSession != nil {
		onNewSession()
	}

	<-endRead
	<-endWrite

	if s.conn != nil {
		_ = s.conn.Close()
	}

	if onQuitSession != nil {
		onQuitSession()
	}
	if atomic.LoadInt32(&s.waitQuitNotify) != 0 {
		s.done <- struct{}{}
	}
}
