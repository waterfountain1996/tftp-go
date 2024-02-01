package tftp

import "io"

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
