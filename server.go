package tftp

import (
	"errors"
	"net"
	"os"
	"time"
)

type serverOpts struct {
	Blocksize  int
	Timeout    time.Duration
	MaxRetries int
}

func defaultServerOpts() *serverOpts {
	return &serverOpts{
		Blocksize:  512,
		Timeout:    3 * time.Second,
		MaxRetries: 5,
	}
}

type Server struct {
	opts *serverOpts
}

func NewServer() *Server {
	o := defaultServerOpts()
	return &Server{
		opts: o,
	}
}

func (s *Server) ListenAndServe(listenAddr string) error {
	pc, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		return err
	}
	defer pc.Close()

	return s.Serve(pc)
}

func (s *Server) Serve(pc net.PacketConn) error {
	buf := make([]byte, 512)
	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			return err
		}

		pkt, err := parsePacket(buf[:n])
		if err != nil {
			continue
		}

		req, ok := pkt.(*request)
		if !ok {
			continue
		}

		go s.handleRequest(req, addr)
	}
}

func (s *Server) handleRequest(req *request, addr net.Addr) {
	conn, err := net.Dial("udp", addr.String())
	if err != nil {
		// TODO: log error
		return
	}
	defer conn.Close()

	pw := newUDPPacketWriter(conn)

	if req.IsWrite {
		pkt := newErrorPacket(errIllegalOp, "operating in read-only mode")
		if err := pw.Write(pkt); err != nil {
			// TODO: Log error
		}
		return
	}

	f, err := os.Open(req.Filename)
	if err != nil {
		var pkt *errorPacket
		switch {
		case errors.Is(err, os.ErrNotExist):
			pkt = newErrorPacket(errNotFound, "file not found")
		case errors.Is(err, os.ErrPermission):
			pkt = newErrorPacket(errPermission, "permission denied")
		default:
			// TODO: Log error
			pkt = newErrorPacket(errUndefined, "internal error")
		}

		if err := pw.Write(pkt); err != nil {
			// TODO: Log error
		}
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		// TODO: Log error
		pkt := newErrorPacket(errUndefined, "internal error")
		if err := pw.Write(pkt); err != nil {
			// TODO: Log error
		}
		return
	}

	if stat.IsDir() {
		pkt := newErrorPacket(errUndefined, "not a file")
		if err := pw.Write(pkt); err != nil {
			// TODO: Log error
		}
		return
	}

	if err := startSender(f, conn); err != nil {
		return
	}
}
