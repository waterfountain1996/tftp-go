package tftp

import (
	"io"
	"sync/atomic"
	"time"
)

func startSender(src io.Reader, conn io.ReadWriter) error {
	var (
		pw       = newTracingPacketWriter(newUDPPacketWriter(conn))
		pr       = newTracingPacketReader(newUDPPacketReader(conn, 512))
		block    = new(atomic.Uint32)
		buf      = make([]byte, 512)
		atEOF    = false
		ackCh    = make(chan bool)
		timeout  = 3 * time.Second
		maxTries = 5
	)

	block.Store(1)

	go func() {
		for {
			pkt, err := pr.Read()
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

		block.Add(1)
	}

	return nil
}

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
