package tftp

import (
	"io"
	"sync/atomic"
	"time"
)

func startReceiver(dst io.Writer, conn io.ReadWriter) error {
	var (
		pw       = newTracingPacketWriter(newUDPPacketWriter(conn))
		pr       = newTracingPacketReader(newUDPPacketReader(conn, 512+4))
		dataCh   = make(chan []byte)
		block    = new(atomic.Uint32)
		timeout  = 3 * time.Second
		maxTries = 5
		atEOF    = false
	)

	go func() {
		for {
			pkt, err := pr.Read()
			if err != nil {
				// TODO: Terminate
				return
			}

			switch pkt := pkt.(type) {
			case *dataPacket:
				if pkt.Block == uint16(block.Load())+1 {
					dataCh <- pkt.Data
				}
			case *errorPacket:
				// TODO: Handle
			}
		}
	}()

Outer:
	for {
		var (
			numTries int
			buf      []byte
		)

	Retransmit:
		for numTries = 0; numTries < maxTries; numTries++ {
			ack := ackPacket{
				Block: uint16(block.Load()),
			}
			if err := pw.Write(ack); err != nil {
				return err
			}

			if atEOF {
				break Outer
			}

			select {
			case buf = <-dataCh:
				break Retransmit
			case <-time.After(timeout):
			}
		}

		if numTries >= maxTries {
			// TODO: Send proper error
			return nil
		}

		if _, err := dst.Write(buf); err != nil {
			return err
		}

		if len(buf) < 512 {
			atEOF = true
		}

		block.Add(1)
	}

	return nil
}
