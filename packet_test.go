package tftp

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePacket(t *testing.T) {
	tests := []struct {
		raw         []byte
		want        packet
		errContains string
	}{
		{
			raw: []byte("\x00\x01foo.txt\x00OcTeT\x00"),
			want: &request{
				Filename: "foo.txt",
				Mode:     "octet",
				IsWrite:  false,
			},
		},
		{
			raw: []byte("\x00\x02bar.txt\x00nEtAsCiI\x00"),
			want: &request{
				Filename: "bar.txt",
				Mode:     "netascii",
				IsWrite:  true,
			},
		},
		{
			raw:         []byte("\x00\x01xxx\x00yyy"),
			errContains: "invalid request",
		},
		{
			raw:         []byte("\x00\x02xxx"),
			errContains: "invalid request",
		},
		{
			raw:         []byte("\x00\x01yyy\x00"),
			errContains: "invalid request",
		},
		{
			raw:         []byte("\x00\x02"),
			errContains: "invalid request",
		},
		{
			raw: []byte("\x00\x03\x00\xABfoobar"),
			want: &dataPacket{
				Block: 171,
				Data:  []byte("foobar"),
			},
		},
		{
			raw: []byte("\x00\x03\xFF\xFF"),
			want: &dataPacket{
				Block: 65535,
				Data:  []byte{},
			},
		},
		{
			raw:         []byte("\x00\x03\x00"),
			errContains: "too short",
		},
		{
			raw: []byte("\x00\x04\xDE\xAD"),
			want: &ackPacket{
				Block: 57005,
			},
		},
		{
			raw:         []byte("\x00\x04\x00"),
			errContains: "too short",
		},
		{
			raw: []byte("\x00\x05\x00\x01error\x00"),
			want: &errorPacket{
				Code:    1,
				Message: "error",
			},
		},
		{
			raw:         []byte("\x00\x05\x00"),
			errContains: "too short",
		},
		{
			raw:         []byte("\x00\x05\x00\x01"),
			errContains: "invalid error message",
		},
	}

	a := assert.New(t)

	for _, test := range tests {
		packet, err := parsePacket(test.raw)
		if test.errContains == "" {
			a.Nil(err)
			a.Equal(packet, test.want)
		} else {
			a.Nil(packet)
			a.NotNil(err)
			a.Contains(err.Error(), test.errContains)
		}
	}

	p := make([]byte, 2)
	for op := uint16(opOACK + 1); op < 1<<16-1; op++ {
		binary.BigEndian.PutUint16(p, op)
		packet, err := parsePacket(p)

		a.Nil(packet)
		a.NotNil(err)
		a.Contains(err.Error(), "unknown op")
	}
}
