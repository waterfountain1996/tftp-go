package tftp

import (
	"io"
	"log"
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
}

func newTracingPacketReader(r packetReader) *tracingPacketReader {
	return &tracingPacketReader{
		packetReader: r,
	}
}

func (r *tracingPacketReader) Read() (p packet, err error) {
	p, err = r.packetReader.Read()
	if err == nil {
		log.Printf("received %s\n", p)
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
}

func newTracingPacketWriter(w packetWriter) *tracingPacketWriter {
	return &tracingPacketWriter{
		packetWriter: w,
	}
}

func (w *tracingPacketWriter) Write(p packet) error {
	err := w.packetWriter.Write(p)
	if err == nil {
		log.Printf("sent %s\n", p)
	}
	return err
}
