package bytepack

import "io"

var (
	packer = NewPacker()
)

func Register(cmd uint16, msg Message) {
	packer.RegisterMsg(cmd, msg)
}

func ReadFrame(c io.Reader) (msg *CmdFrame, err error) {
	return packer.ReadFrame(c)
}

func WriteFrame(c io.Writer, pkg *CmdFrame) (err error) {
	return packer.WriteFrame(c, pkg)
}

func ReadMsg(c io.Reader) (msg Message, err error) {
	return packer.ReadMsg(c)
}

func ReadMsgIn(c io.Reader, msg Message) (err error) {
	return packer.ReadMsgIn(c, msg)
}

func WriteMsg(c io.Writer, msg Message) (err error) {
	return packer.WriteMsg(c, msg)
}

func Write(c io.Writer, data []byte) error {
	return packer.Write(c, data)
}

func Read(c io.Reader) ([]byte, error) {
	return packer.Read(c)
}
