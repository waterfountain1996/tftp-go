package tftp

import (
	"errors"
	"io"
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

	f, err := s.openFile(req.Filename, req.IsWrite)
	if err != nil {
		var (
			pkt     *errorPacket
			pathErr *os.PathError
		)

		if errors.As(err, &pathErr) {
			// TODO: Log err
			pkt = newErrorPacket(errUndefined, "")
		} else {
			pkt, _ = err.(*errorPacket)
		}

		if err := pw.Write(pkt); err != nil {
			// TODO: log error
		}

		return
	}
	defer f.Close()

	opts := &transferOpts{
		Blocksize:  s.opts.Blocksize,
		Timeout:    s.opts.Timeout,
		MaxRetries: s.opts.MaxRetries,
	}

	if req.IsWrite {
		err = startReceiver(f, conn, opts)
	} else {
		err = startSender(f, conn, opts)
	}

	if err != nil {
		// TODO: Log error
	}
}

func (s *Server) openFile(filename string, forWriting bool) (io.ReadWriteCloser, error) {
	var (
		f      *os.File
		err    error
		opener func(string) (*os.File, error)
	)

	if forWriting {
		opener = func(filename string) (*os.File, error) {
			return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
		}
	} else {
		opener = os.Open
	}

	f, err = opener(filename)
	if err != nil {
		switch {
		case os.IsExist(err):
			err = newErrorPacket(errAlreadyExists, "file already exists")
		case os.IsNotExist(err):
			err = newErrorPacket(errNotFound, "file not found")
		case os.IsPermission(err):
			err = newErrorPacket(errPermission, "permission denied")
		}

		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, newErrorPacket(errUndefined, "is a directory")
	}

	return f, err
}
