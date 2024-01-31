package tftp

import (
	"net"
	"os"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
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
	if req.IsWrite {
		// Only handle reads for now
		return
	}

	f, err := os.Open(req.Filename)
	if err != nil {
		return
	}
	defer f.Close()

	conn, err := net.Dial("udp", addr.String())
	if err != nil {
		return
	}
	defer conn.Close()

	if err := startSender(f, conn); err != nil {
		return
	}
}
