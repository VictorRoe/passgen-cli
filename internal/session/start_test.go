package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"os"
	"testing"
	"time"
)

func payloadReader(t *testing.T, key, salt []byte) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(payload{Key: key, Salt: salt})
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(b)
}

func waitGet(t *testing.T, path string) ([]byte, []byte) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if k, s, err := Get(path); err == nil {
			return k, s
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("agent never answered")
	return nil, nil
}

func TestRunAgent_ServesPayload(t *testing.T) {
	path := sockPath(t)
	done := make(chan error, 1)
	go func() { done <- RunAgent(payloadReader(t, []byte("key"), []byte("salt")), path, time.Second) }()
	k, s := waitGet(t, path)
	if string(k) != "key" || string(s) != "salt" {
		t.Errorf("Get() = %q, %q", k, s)
	}
	Lock(path)
	waitDone(t, done)
}

func TestRunAgent_ReplacesStaleSocket(t *testing.T) {
	path := sockPath(t)
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	if ul, ok := ln.(*net.UnixListener); ok {
		ul.SetUnlinkOnClose(false)
	}
	ln.Close()

	done := make(chan error, 1)
	go func() { done <- RunAgent(payloadReader(t, []byte("key"), []byte("salt")), path, time.Second) }()
	waitGet(t, path)
	Lock(path)
	waitDone(t, done)
}

func TestRunAgent_BadPayload(t *testing.T) {
	path := sockPath(t)
	if err := RunAgent(bytes.NewReader(nil), path, time.Second); err == nil {
		t.Error("RunAgent() with empty stdin must fail")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("no socket must be created on bad payload, stat err = %v", err)
	}
}
