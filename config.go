package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type config struct {
	host struct {
		local  string
		remote string
	}
	port struct {
		local  string
		remote string
	}
	protocol   string
	serverMode bool
	count      int
	infinite   bool
	timeout    time.Duration
	deadline   time.Duration
	echoPeriod time.Duration
	pattern    []byte
	debug      bool
	addr       struct {
		local    net.Addr
		localUDP *net.UDPAddr
		remote   string
	}
}

func (c *config) parse(f cliFlags, args []string) error {
	var err error

	c.debug = f.debug
	c.serverMode = f.serverMode
	c.host.local = f.localHost

	if p := strings.ToLower(f.protocol); p == "tcp" || p == "udp" {
		c.protocol = p
	} else {
		if len(p) == 0 {
			return fmt.Errorf("missing protocol")
		} else {
			return fmt.Errorf("invalid protocol: %v", f.protocol)
		}
	}

	if f.localPort < 0 || f.localPort > 65535 {
		return fmt.Errorf("invalid local port number: %v", f.localPort)
	} else {
		c.port.local = strconv.Itoa(f.localPort)
	}

	if len(f.timeout) > 0 {
		c.timeout, err = time.ParseDuration(f.timeout)
		if err != nil {
			return fmt.Errorf("connection timeout: %s", err)
		}
	} else {
		return fmt.Errorf("connection timeout cannot be empty")
	}

	if c.protocol == "tcp" {
		c.addr.local, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(c.host.local, c.port.local))
	} else {
		c.addr.localUDP, err = net.ResolveUDPAddr("udp", net.JoinHostPort(c.host.local, c.port.local))
		c.addr.local = c.addr.localUDP
	}

	if err != nil {
		return fmt.Errorf("local address: %s", err)
	}

	if !c.serverMode {
		if len(args) > 0 && len(args[0]) > 0 {
			c.host.remote = args[0]
		} else {
			return fmt.Errorf("missing remote host from arguments")
		}

		if f.remotePort < 1 || f.remotePort > 65535 {
			if f.remotePort == -1 {
				return fmt.Errorf("missing remote port number")
			} else {
				return fmt.Errorf("invalid remote port number: %v", f.remotePort)
			}
		} else {
			c.port.remote = strconv.Itoa(f.remotePort)
		}

		c.addr.remote = net.JoinHostPort(c.host.remote, c.port.remote)

		if f.count >= 0 {
			c.count = f.count

			if c.count == 0 {
				c.infinite = true
			}
		} else {
			return fmt.Errorf("number of echo requests cannot be smaller than 0")
		}

		if len(f.echoPeriod) > 0 {
			c.echoPeriod, err = time.ParseDuration(f.echoPeriod)
			if err != nil {
				return fmt.Errorf("echo period: %s", err)
			}
		} else {
			return fmt.Errorf("echo period cannot be empty")
		}

		if len(f.deadline) > 0 {
			c.deadline, err = time.ParseDuration(f.deadline)
			if err != nil {
				return fmt.Errorf("read/write deadline: %s", err)
			}
		} else {
			return fmt.Errorf("read/write deadline cannot be empty")
		}

		if len(f.pattern) > 0 {
			c.pattern = []byte(f.pattern)
		} else {
			return fmt.Errorf("pattern cannot be empty")
		}
	}

	return nil
}
