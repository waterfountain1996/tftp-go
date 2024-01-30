package tftp

import "net"

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
	return nil
}
