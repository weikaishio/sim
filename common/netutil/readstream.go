package netutil

import (
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/mkideal/log"
)

const (

	DefaultByteNumForLength = 2
	DefaultByteNumForType   = 2
	DefaultByteNumForSeq    = 2
	DefaultByteNumForMagic  = 0
)

var (
	ErrLengthTooBig = errors.New("length too big")
)

func DecodeLength(buf []byte) int {
	n := 0
	end := len(buf) - 1
	for i := range buf {
		n |= int(buf[end-i]) << (uint32(i) << 3)
	}
	return n
}

func EncodeLength(length int, buf []byte) ([]byte, error) {
	end := len(buf) - 1
	for i := range buf {
		b := length & 0xFF
		buf[end-i] = byte(b)
		length = length >> 8
	}
	if length != 0 {
		return buf, ErrLengthTooBig
	}
	return buf, nil
}

type PacketHandler func(packet []byte)

// 序列化后的数据流的读取
type StreamReader interface {
	Conn() net.Conn
	Read() (n int, err error)
	SetDecrypter(decrypter CrypterFunc)
	SetTimeout(d time.Duration)
}

type ReadStream struct {
	conn             net.Conn
	timeout          time.Duration
	bufForLength     []byte
	buf              []byte
	byteNumForLength int
	onRecvPacket     PacketHandler
	decodeLengthFunc func([]byte) int

	decrypterLocker sync.RWMutex
	decrypter       CrypterFunc
	magic           []byte
	magicLen        int
}

// 通用的流读取
func NewReadStream(conn net.Conn, onRecvPacket PacketHandler) *ReadStream {
	return &ReadStream{
		conn:             conn,
		buf:              make([]byte, 4096),
		byteNumForLength: DefaultByteNumForLength,
		onRecvPacket:     onRecvPacket,
		decodeLengthFunc: DecodeLength,
	}
}

func (r *ReadStream) SetMagic(magic []byte) {
	r.magic = magic
	r.magicLen = len(magic)
}
func (r *ReadStream) SetDecrypter(decrypter CrypterFunc) {
	r.decrypterLocker.Lock()
	defer r.decrypterLocker.Unlock()
	r.decrypter = decrypter
}

func (r *ReadStream) SetTimeout(d time.Duration) { r.timeout = d }

func (r *ReadStream) SetByteNumForLength(n int) {
	r.byteNumForLength = n
	if len(r.buf) < n {
		r.buf = make([]byte, n)
	}
}

func (r *ReadStream) SetDecodeLengthFunc(decodeLengthFunc func([]byte) int) {
	r.decodeLengthFunc = decodeLengthFunc
}

func (r *ReadStream) Conn() net.Conn { return r.conn }

func (r *ReadStream) Read() (total int, err error) {
	defer func() {
		//if e := recover(); e != nil {
		//	log.Error("ReadStream.Read panic: %v", e)
		//}
		//var err error
		if e := recover(); e != nil {
			buf := make([]byte, 1<<16)
			buf = buf[:runtime.Stack(buf, true)]
			switch x := e.(type) {
			case error:
				err = x
			case string:
				err = errors.New(x)
			default:
				err = fmt.Errorf("%v", e)
			}
			err = fmt.Errorf("==== STACK TRACE BEGIN ====\npanic: %v\n%s\n===== STACK TRACE END =====", err, string(buf))
			log.Error("ReadStream.Read panic: %v", err)
		}
	}()

	total = 0
	// 读取 2 个字节以得到数据包的大小
	if r.timeout > 0 {
		r.conn.SetReadDeadline(time.Now().Add(r.timeout))
	}
	if r.magicLen > 0 {
		n, err := r.conn.Read(r.buf[:r.magicLen])
		if err != nil {
			log.Warn("read packet magic: %v", err)
			return total, err
		}
		if n > 0 {
			if string(r.magic) != string(r.buf[:r.magicLen]) {
				log.Warn("string(r.magic):%s,%d != string(r.buf[:r.magicLen]):%s", r.magic, r.magicLen, string(r.buf[:r.magicLen]))
			}
		}
		total += n
	}
	n, err := r.conn.Read(r.buf[r.magicLen : r.magicLen+r.byteNumForLength])
	total += n
	if err != nil {
		log.Info("read packet length: %v", err)
		return total, err
	}
	log.Trace("read packet length: %v", r.buf[r.magicLen:r.magicLen+r.byteNumForLength])

	// 取得当前的 decrypter
	r.decrypterLocker.RLock()
	decrypter := r.decrypter
	r.decrypterLocker.RUnlock()

	// 取得包的大小 包大小不加密
	//if decrypter != nil {
	//	decrypter(r.buf[:r.byteNumForLength], r.buf[:r.byteNumForLength])
	//}
	packetLength := r.decodeLengthFunc(r.buf[r.magicLen : r.magicLen+r.byteNumForLength])
	log.Trace("packet length: %d", packetLength)
	if packetLength < r.byteNumForLength {
		err = fmt.Errorf("packet length: %d less than %d", packetLength, r.byteNumForLength)
		log.Warn("%v", err)
		return total, err
	}
	readedSize := r.byteNumForLength + r.magicLen
	if len(r.buf) < packetLength {
		r.buf = make([]byte, packetLength)
	}

	// 然后按包大小读取数据包 长度不对 会被认为拆包 等待下一个包到超时
	for readedSize < packetLength {
		//log.Trace("read package readedSize:%d < packetLength:%d",readedSize , packetLength)
		n, err := r.conn.Read(r.buf[readedSize:packetLength])
		total += n
		readedSize += n
		if err != nil {
			log.Debug("read packet body error: %v", err)
			return total, err
		}
	}
	if decrypter != nil {
		decrypter(r.buf[r.byteNumForLength:packetLength], r.buf[r.byteNumForLength:packetLength])
	}
	// 处理数据包
	r.onRecvPacket(r.buf[:packetLength])
	return total, nil
}

// UDP 包读取
type UDPReadStream struct {
	conn         *net.UDPConn
	onRecvPacket func([]byte)
	buf          []byte
	timeout      time.Duration
}

func NewUDPReadStream(conn *net.UDPConn, onRecvPacket PacketHandler) *UDPReadStream {
	return &UDPReadStream{
		conn:         conn,
		onRecvPacket: onRecvPacket,
		buf:          make([]byte, 4096),
	}
}

func (r *UDPReadStream) Conn() net.Conn             { return r.conn }
func (r *UDPReadStream) SetDecrypter(CrypterFunc)   { panic("UDPReadStream unsupport decrypter") }
func (r *UDPReadStream) SetTimeout(d time.Duration) { r.timeout = d }

func (r *UDPReadStream) Read(CrypterFunc) (int, error) {
	total := 0
	if r.timeout > 0 {
		r.conn.SetReadDeadline(time.Now().Add(r.timeout))
	}
	n, _, err := r.conn.ReadFromUDP(r.buf)
	total += n
	if err != nil && err != io.EOF {
		return total, err
	}
	if n > 0 {
		r.onRecvPacket(r.buf[:n])
	}
	return total, nil
}
