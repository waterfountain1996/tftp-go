package tftp

import "io"

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
