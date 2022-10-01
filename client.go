package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"strings"
	"time"
)

type client struct {
	config     *config
	sigint     *chan bool
	conn       net.Conn
	ctx        context.Context
	cancel     context.CancelFunc
	ticker     *time.Ticker
	tickerDone chan bool
	stat       struct {
		addr      string
		times     []int64
		sent      int
		received  int
		corrupted int
	}
}

type reply struct {
	err   error
	start int64
	end   int64
	data  []byte
	len   int
}

func (c *client) run() {
	var err error

	c.init()

	dialer := &net.Dialer{
		LocalAddr: c.config.addr.local,
	}

	c.conn, err = dialer.DialContext(c.ctx, c.config.protocol, c.config.addr.remote)
	if err != nil {
		printNetworkError(err, c.config.debug)
		os.Exit(1)
	}

	defer c.conn.Close()

	c.stat.addr = c.conn.RemoteAddr().String()

	remoteAddr, _, _ := net.SplitHostPort(c.conn.RemoteAddr().String())
	fmt.Printf("Hostname %s resolved as %s\n\n", c.config.host.remote, remoteAddr)

	c.startEcho()
	c.printStat()
}

func (c *client) init() {
	c.ticker = time.NewTicker(c.config.echoPeriod)
	c.tickerDone = make(chan bool)

	c.ctx, c.cancel = context.WithTimeout(context.Background(), c.config.timeout)

	go c.close()
}

func (c *client) close() {
	<-*c.sigint

	c.cancel()
	if c.conn != nil {
		c.conn.Close()
	}

	c.ticker.Stop()
	c.tickerDone <- true
}

func (c *client) startEcho() {
	counter := c.config.count

	for {
		select {
		case <-c.tickerDone:
			return
		case <-c.ticker.C:
			if !c.config.infinite {
				if counter == 0 {
					return
				} else {
					counter--
				}
			}

			if c.sendEcho() {
				return
			}
		}
	}
}

func (c *client) sendEcho() bool {
	c.stat.sent++
	buf := make([]byte, 4096)

	c.conn.SetWriteDeadline(time.Now().Add(c.config.deadline))
	_, err := c.conn.Write(c.config.pattern)
	if err != nil {
		c.printReply(reply{err: err})
		return errors.Is(err, net.ErrClosed) || isConnectionReset(err)
	}

	start := time.Now().UnixMilli()

	c.conn.SetReadDeadline(time.Now().Add(c.config.deadline))
	n, err := c.conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			c.printReply(reply{err: net.ErrClosed})
			return true
		}

		c.printReply(reply{err: err})
		return errors.Is(err, net.ErrClosed) || isConnectionReset(err)
	}

	end := time.Now().UnixMilli()
	c.stat.received++

	c.printReply(reply{
		start: start,
		end:   end,
		data:  buf,
		len:   n,
	})

	return false
}

func (c *client) printReply(r reply) {
	if r.err != nil {
		printNetworkError(r.err, c.config.debug)
	} else {
		echoTime := r.end - r.start
		c.stat.times = append(c.stat.times, echoTime)

		status := "OK"
		if !bytes.Equal(c.config.pattern, r.data[:r.len]) {
			c.stat.corrupted++
			status = "CORRUPT"
		}

		fmt.Printf("Reply from %s, time %v ms, %s\n", c.stat.addr, echoTime, status)
	}
}

func (c *client) printStat() {
	if c.stat.sent == 0 {
		return
	}

	min, max, avg := c.getTimeStat()

	lost := c.stat.sent - c.stat.received
	loss := math.Floor(((float64(lost)*float64(100))/float64(c.stat.sent))*100) / 100

	fmt.Printf("\n--- %s %s echo statistics ---\n", c.stat.addr, strings.ToUpper(c.config.protocol))
	fmt.Printf("%v echo request sent, %v received, %v lost (%v%% loss), %v corrupted\n", c.stat.sent, c.stat.received, lost, loss, c.stat.corrupted)
	fmt.Printf("Round-trip time min/avg/max: %v / %v / %v ms\n", min, avg, max)
}

func (c *client) getTimeStat() (int64, int64, float64) {
	var min, max int64
	t := c.stat.times

	if len(t) == 0 {
		return 0, 0, 0
	}

	if len(t) == 1 {
		return t[0], t[0], float64(t[0])
	}

	if len(t) >= 2 {
		min = t[0]
		max = t[0]
	}

	var total int64
	for _, v := range t {
		total += v

		if min > v {
			min = v
		}

		if max < v {
			max = v
		}
	}

	avg := math.Floor((float64(total)/float64(len(t)))*100) / 100

	return min, max, avg
}
