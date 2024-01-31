package tftp

import (
	"io"
	"sync/atomic"
	"time"
)

func startSender(src io.Reader, conn io.ReadWriter) error {
	var (
		block = new(atomic.Uint32)
		buf   = make([]byte, 512)
		atEOF = false
		ackCh = make(chan bool)
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

	Retransmit:
		for {
			if _, err := conn.Write(pkt.Bytes()); err != nil {
				return err
			}

			select {
			case <-ackCh:
				break Retransmit
			case <-time.After(3 * time.Second):
			}
		}
	}

	return nil
}
