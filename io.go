package tftp

import (
	"fmt"
	"io"
	"os"
)

type packetReader interface {
	Read() (p packet, err error)
}

type udpPacketReader struct {
	r   io.Reader
	buf []byte
}

func newUDPPacketReader(r io.Reader, bufsize int) *udpPacketReader {
	return &udpPacketReader{
		r:   r,
		buf: make([]byte, bufsize),
	}
}

func (r *udpPacketReader) Read() (p packet, err error) {
	n, err := r.r.Read(r.buf)
	if err != nil {
		return nil, err
	}

	return parsePacket(r.buf[:n])
}

type tracingPacketReader struct {
	packetReader
	trace traceFunc
}

func newTracingPacketReader(r packetReader, trace traceFunc) *tracingPacketReader {
	return &tracingPacketReader{
		packetReader: r,
		trace:        trace,
	}
}

func (r *tracingPacketReader) Read() (p packet, err error) {
	p, err = r.packetReader.Read()
	if err == nil {
		r.trace(p)
	}
	return
}

type packetWriter interface {
	Write(p packet) error
}

type udpPacketWriter struct {
	w io.Writer
}

func newUDPPacketWriter(w io.Writer) *udpPacketWriter {
	return &udpPacketWriter{
		w: w,
	}
}

func (w *udpPacketWriter) Write(p packet) error {
	_, err := w.w.Write(p.Bytes())
	return err
}

type tracingPacketWriter struct {
	packetWriter
	trace traceFunc
}

func newTracingPacketWriter(w packetWriter, trace traceFunc) *tracingPacketWriter {
	return &tracingPacketWriter{
		packetWriter: w,
		trace:        trace,
	}
}

func (w *tracingPacketWriter) Write(p packet) error {
	err := w.packetWriter.Write(p)
	if err == nil {
		w.trace(p)
	}
	return err
}

type traceFunc func(p packet)

func traceSend(p packet) {
	fmt.Fprintf(os.Stderr, "sent %s\n", p)
}

func traceReceive(p packet) {
	fmt.Fprintf(os.Stderr, "received %s\n", p)
}
