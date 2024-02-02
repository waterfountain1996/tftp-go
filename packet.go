package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	errInvalidPacket = errors.New("tftp: invalid packet")
)

type packet interface {
	Op() opcode
	Bytes() []byte
	String() string
}

type requestOpt struct {
	Name  string
	Value string
}

func (o requestOpt) String() string {
	return fmt.Sprintf("%s: %s", o.Name, o.Value)
}

type request struct {
	Filename string
	Mode     string
	IsWrite  bool
	Opts     []requestOpt
}

func (r request) Op() opcode {
	if r.IsWrite {
		return opWRQ
	} else {
		return opRRQ
	}
}

func (r request) Bytes() []byte {
	var b bytes.Buffer
	binary.Write(&b, binary.BigEndian, r.Op())
	b.WriteString(r.Filename)
	b.WriteByte(0)
	b.WriteString(r.Mode)
	b.WriteByte(0)
	return b.Bytes()
}

func (r request) String() string {
	s := fmt.Sprintf("%s <file: %s, mode: %s", r.Op(), r.Filename, r.Mode)

	if len(r.Opts) > 0 {
		opts := make([]string, len(r.Opts))
		for i, o := range r.Opts {
			opts[i] = o.String()
		}

		optString := strings.Join(opts, ", ")
		s += fmt.Sprintf(", opts: <%s>", optString)
	}

	s += ">"

	return s
}

type dataPacket struct {
	Block uint16
	Data  []byte
}

func (p dataPacket) Op() opcode {
	return opDATA
}

func (p dataPacket) Bytes() []byte {
	var b bytes.Buffer
	binary.Write(&b, binary.BigEndian, p.Op())
	binary.Write(&b, binary.BigEndian, p.Block)
	b.Write(p.Data)
	return b.Bytes()
}

func (p dataPacket) String() string {
	return fmt.Sprintf("%s <block: %d, size: %d>", p.Op(), p.Block, len(p.Data))
}

type ackPacket struct {
	Block uint16
}

func (p ackPacket) Op() opcode {
	return opACK
}

func (p ackPacket) Bytes() []byte {
	var b bytes.Buffer
	binary.Write(&b, binary.BigEndian, p.Op())
	binary.Write(&b, binary.BigEndian, p.Block)
	return b.Bytes()
}

func (p ackPacket) String() string {
	return fmt.Sprintf("%s <block: %d>", p.Op(), p.Block)
}

type errorPacket struct {
	Code    uint16
	Message string
}

func newErrorPacket(code uint16, message string) *errorPacket {
	return &errorPacket{
		Code:    code,
		Message: message,
	}
}

func (p errorPacket) Op() opcode {
	return opERROR
}

func (p errorPacket) Bytes() []byte {
	var b bytes.Buffer
	binary.Write(&b, binary.BigEndian, p.Op())
	binary.Write(&b, binary.BigEndian, p.Code)
	b.WriteString(p.Message)
	b.WriteByte(0)
	return b.Bytes()
}

func (p errorPacket) String() string {
	return fmt.Sprintf("%s <code: %d, message: %s>", p.Op(), p.Code, p.Message)
}

func parsePacket(p []byte) (packet, error) {
	if len(p) < 2 {
		return nil, fmt.Errorf("%w: too short", errInvalidPacket)
	}

	op := opcode(binary.BigEndian.Uint16(p))
	p = p[2:]

	switch op {
	case opRRQ, opWRQ:
		filename, idx, err := readString(p)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid request", errInvalidPacket)
		}
		p = p[idx:]

		mode, idx, err := readString(p)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid request", errInvalidPacket)
		}
		p = p[idx:]

		opts, err := readOpts(p)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid request", errInvalidPacket)
		}

		req := &request{
			Filename: filename,
			Mode:     strings.ToLower(mode),
			IsWrite:  op == opWRQ,
			Opts:     opts,
		}
		return req, nil
	case opDATA:
		if len(p) < 2 {
			return nil, fmt.Errorf("%w: too short", errInvalidPacket)
		}

		pkt := &dataPacket{
			Block: binary.BigEndian.Uint16(p),
			Data:  p[2:],
		}
		return pkt, nil
	case opACK:
		if len(p) < 2 {
			return nil, fmt.Errorf("%w: too short", errInvalidPacket)
		}

		pkt := &ackPacket{
			Block: binary.BigEndian.Uint16(p),
		}
		return pkt, nil
	case opERROR:
		if len(p) < 2 {
			return nil, fmt.Errorf("%w: too short", errInvalidPacket)
		}

		errCode := binary.BigEndian.Uint16(p)
		message, _, err := readString(p[2:])
		if err != nil {
			return nil, fmt.Errorf("%w: invalid error message", errInvalidPacket)
		}

		pkt := &errorPacket{
			Code:    errCode,
			Message: message,
		}
		return pkt, err
	default:
		return nil, fmt.Errorf("%w: unknown op: %d", errInvalidPacket, op)
	}
}

func readOpts(p []byte) ([]requestOpt, error) {
	if len(p) == 0 {
		return nil, nil
	}

	opts := []requestOpt{}

	for len(p) > 0 {
		name, idx, err := readString(p)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid request", errInvalidPacket)
		}
		p = p[idx:]

		value, idx, err := readString(p)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid request", errInvalidPacket)
		}
		p = p[idx:]

		o := requestOpt{
			Name:  name,
			Value: value,
		}
		opts = append(opts, o)
	}

	return opts, nil
}

func readString(p []byte) (string, int, error) {
	if len(p) == 0 {
		return "", 0, io.EOF
	}

	idx := bytes.IndexByte(p, 0)
	if idx < 0 {
		return "", 0, io.ErrUnexpectedEOF
	}

	return string(p[:idx]), idx + 1, nil
}
