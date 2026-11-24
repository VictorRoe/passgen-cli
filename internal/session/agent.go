package session

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"time"
)

// Serve answers get/lock requests on ln until the session expires (ttl from
// now, checked against the wall clock as well as a timer) or a lock request
// arrives. On return the key is zeroed and the socket is closed and removed.
func Serve(ln net.Listener, key, salt []byte, ttl time.Duration, now func() time.Time) error {
	expiresAt := now().Add(ttl)
	path := ln.Addr().String()
	defer func() {
		for i := range key {
			key[i] = 0
		}
		ln.Close()
		os.Remove(path)
	}()

	// The timer covers the normal case; the ticker catches a suspended machine
	// whose monotonic timer stopped while the wall clock moved on.
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		timer := time.NewTimer(ttl)
		defer timer.Stop()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-timer.C:
				ln.Close()
				return
			case <-ticker.C:
				if !now().Before(expiresAt) {
					ln.Close()
					return
				}
			case <-stop:
				return
			}
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return nil // listener closed: expired or locked
		}
		if done := handle(conn, key, salt, expiresAt, now); done {
			return nil
		}
	}
}

// handle serves one request and reports whether the agent must stop.
func handle(conn net.Conn, key, salt []byte, expiresAt time.Time, now func() time.Time) bool {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(ioTimeout))

	var req request
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&req); err != nil {
		return false
	}
	enc := json.NewEncoder(conn)

	if !now().Before(expiresAt) {
		enc.Encode(response{Err: "expired"})
		return true
	}
	switch req.Op {
	case "get":
		enc.Encode(response{Key: key, Salt: salt, ExpiresAt: expiresAt})
		return false
	case "lock":
		enc.Encode(response{OK: true})
		return true
	default:
		enc.Encode(response{Err: "unknown op"})
		return false
	}
}
