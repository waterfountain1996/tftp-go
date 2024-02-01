package tftp

import (
	"io"
	"sync/atomic"
	"time"
)

func startSender(src io.Reader, conn io.ReadWriter) error {
	var (
		pw       = newUDPPacketWriter(conn)
		block    = new(atomic.Uint32)
		buf      = make([]byte, 512)
		atEOF    = false
		ackCh    = make(chan bool)
		timeout  = 3 * time.Second
		maxTries = 5
	)

	block.Store(1)

	go func() {
		b := make([]byte, 4)
		for {
			n, err := conn.Read(b)
			if err != nil {
				// TODO: Handle
				return
			}

			pkt, err := parsePacket(b[:n])
			if err != nil {
				// TODO: Terminate
				return
			}

			switch pkt := pkt.(type) {
			case *ackPacket:
				if pkt.Block == uint16(block.Load()) {
					ackCh <- true
				}
			case *errorPacket:
				// TODO: Handle
			}
		}
	}()

	for !atEOF {
		n, err := io.ReadFull(src, buf)
		if err != nil {
			if !(err == io.EOF || err == io.ErrUnexpectedEOF) {
				return err
			}

			atEOF = true
		}

		pkt := dataPacket{
			Block: uint16(block.Load()),
			Data:  buf[:n],
		}

		var numTries int
	Retransmit:
		for numTries = 0; numTries < maxTries; numTries++ {
			if err := pw.Write(pkt); err != nil {
				return err
			}

			select {
			case <-ackCh:
				break Retransmit
			case <-time.After(timeout):
			}
		}

		if numTries >= maxTries {
			return nil
		}
	}

	return nil
}
