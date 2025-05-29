package bytepack

import (
	"encoding/binary"
	"errors"
	"github.com/bytedance/sonic"
	"io"
	"reflect"
)

const (
	cmdLen     = 2
	payloadLen = 4
	HeaderLen  = cmdLen + payloadLen
)

var defaultMaxMsgLength uint32 = 1024 * 1024 * 1024 // 1024MB = 1GMbps的宽带

var (
	ErrMsgType      = errors.New("message type not supported")
	ErrMaxMsgLength = errors.New("message length exceed the limit")
	ErrMsgLength    = errors.New("message length error")
	ErrMsgFormat    = errors.New("message format error")
	ErrPackageNil   = errors.New("package is nil")
)

type cmdType uint16

type Message interface{}

type Package struct {
	Cmd     cmdType
	Payload []byte
}

func NewPackage(cmd cmdType) *Package {
	return &Package{
		Cmd: cmd,
	}
}
func (sel *Package) Len() int {
	return len(sel.Payload) + HeaderLen
}

func (sel *Package) String() string {
	return string(sel.Payload)
}

func (sel *Package) Decode(obj Message) error {
	return sonic.Unmarshal(sel.Payload, obj)
}
func (sel *Package) Encode(obj Message) error {
	data, err := sonic.Marshal(obj)
	if err != nil {
		return err
	}
	sel.Payload = data
	return nil
}

type Packer struct {
	cmdMap     map[cmdType]reflect.Type
	cmdTypeMap map[reflect.Type]cmdType
}

func NewPack() *Packer {
	p := &Packer{
		cmdMap:     make(map[cmdType]reflect.Type),
		cmdTypeMap: make(map[reflect.Type]cmdType),
	}
	p.RegisterMsg(0, "")
	return p
}
func (sel *Packer) writeData(c io.Writer, cmd cmdType, data []byte) (int, error) {
	dataLen := len(data)
	buffer := make([]byte, dataLen+HeaderLen)
	binary.BigEndian.PutUint16(buffer[0:cmdLen], uint16(cmd))
	binary.BigEndian.PutUint32(buffer[cmdLen:HeaderLen], uint32(dataLen))
	copy(buffer[HeaderLen:], data)
	return c.Write(buffer)
}

func (sel *Packer) readData(r io.Reader) (cmd cmdType, buffer []byte, err error) {
	buffer = make([]byte, HeaderLen)
	if _, err = io.ReadFull(r, buffer); err != nil {
		return
	}
	if len(buffer) < HeaderLen {
		err = ErrMsgLength
		return
	}
	cmd = cmdType(binary.BigEndian.Uint16(buffer[0:cmdLen]))
	if _, ok := sel.cmdMap[cmd]; !ok {
		err = ErrMsgType
		return
	}

	length := binary.BigEndian.Uint32(buffer[cmdLen:HeaderLen])
	if length > defaultMaxMsgLength {
		err = ErrMaxMsgLength
		return
	}

	buffer = make([]byte, length)
	n, err := io.ReadFull(r, buffer)
	if err != nil {
		return
	}
	if n != int(length) { // 类型统一比较
		err = ErrMsgFormat
	}
	return
}

func (sel *Packer) RegisterMsg(msgType uint16, msg Message) {
	cmd := cmdType(msgType)
	tp := reflect.TypeOf(msg)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	sel.cmdMap[cmd] = tp
	sel.cmdTypeMap[tp] = cmd
}

func (sel *Packer) ReadPkg(r io.Reader) (*Package, error) {
	cmd, data, err := sel.readData(r)
	if err != nil {
		return nil, err
	}
	return &Package{Cmd: cmd, Payload: data}, nil
}

func (sel *Packer) unpack(typeByte cmdType, buffer []byte, msgIn Message) (msg Message, err error) {
	msg = msgIn
	if msg == nil {
		t, ok := sel.cmdMap[typeByte]
		if !ok {
			err = ErrMsgType
			return
		}
		if t.Kind() == reflect.String {
			msg = string(buffer)
			return
		}
		msg = reflect.New(t).Interface()
	}
	err = sonic.Unmarshal(buffer, &msg)
	return
}

func (sel *Packer) ReadString(r io.Reader) (string, error) {
	_, data, err := sel.readData(r)
	return string(data), err
}

func (sel *Packer) pack(cmd cmdType, msg []byte) []byte {
	buffer := make([]byte, len(msg)+HeaderLen)
	binary.BigEndian.PutUint16(buffer[0:cmdLen], uint16(cmd))
	binary.BigEndian.PutUint32(buffer[cmdLen:HeaderLen], uint32(len(msg)))
	copy(buffer[HeaderLen:], msg)
	return buffer
}

func (sel *Packer) ReadMsgIn(r io.Reader, msg Message) (err error) {
	cmd, data, err := sel.readData(r)
	_, err = sel.unpack(cmd, data, msg)
	return err
}

func (sel *Packer) ReadMsg(r io.Reader) (msg Message, err error) {
	typeByte, buffer, err := sel.readData(r)
	if err != nil {
		return nil, err
	}
	return sel.unpack(typeByte, buffer, nil)
}

func (sel *Packer) decodeMsg(msg Message) (data []byte, err error) {
	switch m := msg.(type) {
	case string:
		data = []byte(m)
	default:
		data, err = sonic.Marshal(m)
	}
	return
}

func (sel *Packer) WriteString(c io.Writer, msg string) error {
	_, err := sel.writeData(c, 0, []byte(msg))
	return err
}

func (sel *Packer) WriteMsg(c io.Writer, msg Message) (err error) {
	tp := reflect.TypeOf(msg)
	if tp.Kind() == reflect.Ptr {
		tp = reflect.TypeOf(msg).Elem()
	}
	cmd, ok := sel.cmdTypeMap[tp]
	if !ok {
		return ErrMsgType
	}

	content, err := sel.decodeMsg(msg)
	if err != nil {
		return err
	}

	if _, err = sel.writeData(c, cmd, content); err != nil {
		return
	}
	return nil
}

func (sel *Packer) WritePkg(c io.Writer, p *Package) error {
	if p == nil {
		return ErrPackageNil
	}
	if _, ok := sel.cmdMap[p.Cmd]; !ok {
		return ErrMsgType
	}
	_, err := sel.writeData(c, p.Cmd, p.Payload)
	return err
}
