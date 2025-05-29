package bytepack

import (
	"encoding/binary"
	"io"
)

type SimPacker struct {
}

func (sel *SimPacker) Write(writer io.Writer, data []byte) (int, error) {
	size := len(data)
	buffer := make([]byte, payloadLen+size)
	binary.BigEndian.PutUint32(buffer[0:payloadLen], uint32(size))
	copy(buffer[payloadLen:], data)
	_, err := writer.Write(buffer)
	return size, err
}

func (sel *SimPacker) Read(reader io.Reader) ([]byte, error) {
	sizeBuf := make([]byte, payloadLen)
	if _, err := io.ReadFull(reader, sizeBuf); err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(sizeBuf)
	if size == 0 {
		return []byte{}, nil
	}
	data := make([]byte, size)
	_, err := io.ReadFull(reader, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
