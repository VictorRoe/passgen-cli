package session

import (
	"bytes"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func sockPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "pg")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return filepath.Join(dir, "a.sock")
}

// startServe runs Serve in the background and returns a channel with its result.
func startServe(t *testing.T, path string, key, salt []byte, ttl time.Duration, now func() time.Time) <-chan error {
	t.Helper()
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- Serve(ln, key, salt, ttl, now) }()
	return done
}

func waitDone(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not return")
	}
}

func TestAgent_GetReturnsKeyAndSalt(t *testing.T) {
	path := sockPath(t)
	done := startServe(t, path, []byte("k123"), []byte("s456"), time.Hour, time.Now)
	key, salt, err := Get(path)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !bytes.Equal(key, []byte("k123")) || !bytes.Equal(salt, []byte("s456")) {
		t.Errorf("Get() = %q, %q", key, salt)
	}
	Lock(path)
	waitDone(t, done)
}

func TestAgent_LockStopsAgent(t *testing.T) {
	path := sockPath(t)
	done := startServe(t, path, []byte("k"), []byte("s"), time.Hour, time.Now)
	ok, err := Lock(path)
	if err != nil || !ok {
		t.Fatalf("Lock() = %v, %v, want true, nil", ok, err)
	}
	waitDone(t, done)
	if _, _, err := Get(path); !errors.Is(err, ErrNoSession) {
		t.Errorf("Get() after Lock = %v, want ErrNoSession", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("socket must be removed after Lock, stat err = %v", err)
	}
}

func TestAgent_ExpiresAfterTTL(t *testing.T) {
	path := sockPath(t)
	done := startServe(t, path, []byte("k"), []byte("s"), 200*time.Millisecond, time.Now)
	waitDone(t, done)
	if _, _, err := Get(path); !errors.Is(err, ErrNoSession) {
		t.Errorf("Get() after TTL = %v, want ErrNoSession", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("socket must be removed after expiry, stat err = %v", err)
	}
}

func TestAgent_RejectsAfterWallClockExpiry(t *testing.T) {
	path := sockPath(t)
	start := time.Now()
	var offset atomic.Int64
	now := func() time.Time { return start.Add(time.Duration(offset.Load())) }
	done := startServe(t, path, []byte("k"), []byte("s"), time.Hour, now)
	// Simulate a laptop that slept for 2 hours: the timer has not fired yet.
	time.Sleep(50 * time.Millisecond)
	offset.Store(int64(2 * time.Hour))
	if _, _, err := Get(path); !errors.Is(err, ErrNoSession) {
		t.Errorf("Get() after wall-clock expiry = %v, want ErrNoSession", err)
	}
	waitDone(t, done)
}

func TestGet_NoSocket(t *testing.T) {
	if _, _, err := Get(sockPath(t)); !errors.Is(err, ErrNoSession) {
		t.Errorf("Get() = %v, want ErrNoSession", err)
	}
}

func TestGet_StaleSocket(t *testing.T) {
	path := sockPath(t)
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	// Leave the socket file behind with nobody listening.
	if ul, ok := ln.(*net.UnixListener); ok {
		ul.SetUnlinkOnClose(false)
	}
	ln.Close()
	if _, _, err := Get(path); !errors.Is(err, ErrNoSession) {
		t.Errorf("Get() on stale socket = %v, want ErrNoSession", err)
	}
}

func TestLock_NoSession(t *testing.T) {
	ok, err := Lock(sockPath(t))
	if ok || err != nil {
		t.Errorf("Lock() = %v, %v, want false, nil", ok, err)
	}
}
