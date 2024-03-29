package tftp

import (
	"bufio"
	"errors"
	"io"
	"sync/atomic"
	"time"
)

var errClientTimeout = errors.New("tftp: client timeout")

type transferOpts struct {
	Blocksize  int
	Timeout    time.Duration
	MaxRetries int
	Trace      bool
}

func startSender(src io.Reader, conn io.ReadWriter, opts *transferOpts) error {
	var (
		pw    packetWriter = newUDPPacketWriter(conn)
		pr    packetReader = newUDPPacketReader(conn, opts.Blocksize)
		block              = new(atomic.Uint32)
		buf                = make([]byte, opts.Blocksize)
		atEOF              = false
		ackCh              = make(chan bool, 1)
		errCh              = make(chan error, 1)
	)

	if opts.Trace {
		pw = newTracingPacketWriter(pw, traceSend)
		pr = newTracingPacketReader(pr, traceReceive)
	}

	block.Store(1)

	go func() {
		defer func() {
			close(ackCh)
			close(errCh)
		}()

		for {
			pkt, err := pr.Read()
			if err != nil {
				errCh <- err
				return
			}

			switch pkt := pkt.(type) {
			case *ackPacket:
				if pkt.Block == uint16(block.Load()) {
					ackCh <- true
				}
			case *errorPacket:
				errCh <- pkt
				return
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
		for numTries = 0; numTries < opts.MaxRetries; numTries++ {
			if err := pw.Write(pkt); err != nil {
				return err
			}

			select {
			case <-ackCh:
				break Retransmit
			case err := <-errCh:
				return err
			case <-time.After(opts.Timeout):
			}
		}

		if numTries >= opts.MaxRetries {
			return errClientTimeout
		}

		block.Add(1)
	}

	return nil
}

func startReceiver(dst io.Writer, conn io.ReadWriter, opts *transferOpts) error {
	var (
		w                   = bufio.NewWriter(dst)
		pw     packetWriter = newUDPPacketWriter(conn)
		pr     packetReader = newUDPPacketReader(conn, opts.Blocksize+4)
		dataCh              = make(chan []byte, 1)
		errCh               = make(chan error, 1)
		block               = new(atomic.Uint32)
		atEOF               = false
	)

	if opts.Trace {
		pw = newTracingPacketWriter(pw, traceSend)
		pr = newTracingPacketReader(pr, traceReceive)
	}

	go func() {
		defer func() {
			close(dataCh)
			close(errCh)
		}()

		for {
			pkt, err := pr.Read()
			if err != nil {
				errCh <- err
				return
			}

			switch pkt := pkt.(type) {
			case *dataPacket:
				if pkt.Block == uint16(block.Load())+1 {
					dataCh <- pkt.Data
				}
			case *errorPacket:
				errCh <- pkt
				return
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
		for numTries = 0; numTries < opts.MaxRetries; numTries++ {
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
			case err := <-errCh:
				return err
			case <-time.After(opts.Timeout):
			}
		}

		if numTries >= opts.MaxRetries {
			return errClientTimeout
		}

		if _, err := w.Write(buf); err != nil {
			return err
		}

		if len(buf) < opts.Blocksize {
			atEOF = true
		}

		block.Add(1)
	}

	w.Flush()

	return nil
}
