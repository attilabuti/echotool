package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

type server struct {
	config      *config
	sigint      *chan bool
	tcpListener net.Listener
	udpConn     *net.UDPConn
	requests    chan request
}

type request struct {
	err       error
	addr      string
	data      []byte
	timestamp time.Time
}

func (s *server) run() {
	var err error

	s.init()

	if s.config.protocol == "tcp" {
		s.tcpListener, err = net.Listen(s.config.protocol, s.config.addr.local.String())
	} else {
		s.udpConn, err = net.ListenUDP(s.config.protocol, s.config.addr.localUDP)
	}

	if err != nil {
		printNetworkError(err, s.config.debug)
		os.Exit(1)
	}

	fmt.Printf("Waiting for %s connection on port %v\n", strings.ToUpper(s.config.protocol), s.config.port.local)

	if s.config.protocol == "tcp" {
		s.tcp()
	} else {
		s.udp()
	}
}

func (s *server) init() {
	go s.close()

	s.requests = make(chan request)
	go s.printRequests()
}

func (s *server) close() {
	<-*s.sigint

	if s.config.protocol == "tcp" {
		if s.tcpListener != nil {
			s.tcpListener.Close()
		}
	} else {
		s.udpConn.Close()
	}
}

func (s *server) tcp() {
	for {
		conn, err := s.tcpListener.Accept()
		if err != nil {
			printNetworkError(err, s.config.debug)

			if errors.Is(err, net.ErrClosed) || isConnectionReset(err) {
				return
			}

			continue
		}

		defer conn.Close()

		remoteAddr := conn.RemoteAddr().String()
		fmt.Printf("Client %s accepted at %v\n", remoteAddr, time.Now().Format("15:04:05"))

		go s.handleTCPConnection(conn, remoteAddr)
	}
}

func (s *server) handleTCPConnection(conn net.Conn, remoteAddr string) {
	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				s.requests <- request{err: net.ErrClosed}
				return
			}

			s.requests <- request{err: err}

			return
		}

		if n == 0 {
			return
		}

		s.requests <- request{
			addr:      remoteAddr,
			data:      buf[:n],
			timestamp: time.Now(),
		}

		_, err = conn.Write(buf[:n])
		if err != nil {
			s.requests <- request{err: err}
			return
		}
	}
}

func (s *server) udp() {
	buf := make([]byte, 4096)

	for {
		n, addr, err := s.udpConn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			s.requests <- request{err: err}

			continue
		}

		if n > 0 {
			s.requests <- request{
				addr:      addr.String(),
				data:      buf[:n],
				timestamp: time.Now(),
			}

			_, err = s.udpConn.WriteTo(buf[:n], addr)
			if err != nil {
				s.requests <- request{err: err}
			}
		}
	}
}

func (s *server) printRequests() {
	for {
		r := <-s.requests

		if r.err != nil {
			printNetworkError(r.err, s.config.debug)
		} else {
			fmt.Printf("%v received [%s] from %s\n", r.timestamp.Format("15:04:05.000000"), r.data, r.addr)
		}
	}
}
