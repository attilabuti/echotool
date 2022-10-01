package main

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

type networkError struct {
	errno    syscall.Errno
	errnoWin uint
	message  string
}

var networkErrors = []networkError{
	{syscall.EACCES, 10013, "Permission denied"},
	{syscall.EWOULDBLOCK, 10035, "Resource temporarily unavailable"},
	{syscall.EMSGSIZE, 10040, "Message too long"},
	{syscall.EADDRINUSE, 10048, "Address already in use"},
	{syscall.ENETDOWN, 10050, "Network is down"},
	{syscall.ENETUNREACH, 10051, "Network is unreachable"},
	{syscall.ENETRESET, 10052, "Network dropped connection on reset"},
	{syscall.ECONNABORTED, 10053, "Software caused connection abort"},
	{syscall.ECONNRESET, 10054, "Connection reset by peer"},
	{syscall.ETIMEDOUT, 10060, "Connection timed out"},
	{syscall.ECONNREFUSED, 10061, "Connection refused"},
	{syscall.EHOSTDOWN, 10064, "Host is down"},
	{syscall.EHOSTUNREACH, 10065, "No route to host"},
}

func printNetworkError(err error, debug bool) {
	var errMsg string

	oErr := err
	if err == nil {
		return
	}

	for errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}

	if cause, ok := err.(*net.DNSError); ok && cause.IsNotFound {
		errMsg = fmt.Sprintf("Host not found: %s", cause.Name)
	} else if cause, ok := err.(syscall.Errno); ok {
		for _, v := range networkErrors {
			if syscall.Errno(v.errnoWin) == cause || v.errno == cause {
				errMsg = v.message
				break
			}
		}
	} else if cause, ok := err.(net.Error); ok {
		if cause.Timeout() {
			errMsg = "Connection timed out"
		} else if errors.Is(cause, net.ErrClosed) {
			errMsg = "Connection closed"
		}
	}

	if len(errMsg) == 0 {
		errMsg = fmt.Sprintf("Unknown network error: %s", err)
	}

	fmt.Println(errMsg)
	if debug {
		fmt.Printf("Error: %s\n", oErr)
	}
}

func isConnectionReset(err error) bool {
	for errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}

	if errno, ok := err.(syscall.Errno); ok {
		if syscall.Errno(10054) == errno || syscall.ECONNRESET == errno {
			return true
		}
	}

	return false
}
