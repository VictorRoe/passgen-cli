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
	expiresAt := expiryFrom(now(), ttl)
	path := ln.Addr().String()

	// Remove the socket on exit only if it is still ours: a newer agent may
	// have replaced it, and deleting its socket would make it unreachable
	// (even for "passgen lock") while it still holds a key.
	if ul, ok := ln.(*net.UnixListener); ok {
		ul.SetUnlinkOnClose(false)
	}
	own, ownErr := os.Stat(path)
	defer func() {
		for i := range key {
			key[i] = 0
		}
		ln.Close()
		if ownErr == nil {
			if cur, err := os.Stat(path); err == nil && os.SameFile(own, cur) {
				os.Remove(path)
			}
		}
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
				if expired(now(), expiresAt) {
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

	if expired(now(), expiresAt) {
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

// expiryFrom returns now+ttl as a wall-clock time. The monotonic reading is
// dropped because Go's monotonic clock stops while Linux and macOS sleep,
// which would let a suspended machine extend the session.
func expiryFrom(now time.Time, ttl time.Duration) time.Time {
	return now.Round(0).Add(ttl)
}

// expired compares wall-clock times only.
func expired(now, expiresAt time.Time) bool {
	return !now.Round(0).Before(expiresAt.Round(0))
}
