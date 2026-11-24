package session

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func call(path string, op string) (response, error) {
	conn, err := net.DialTimeout("unix", path, ioTimeout)
	if err != nil {
		return response{}, ErrNoSession
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(ioTimeout))

	if err := json.NewEncoder(conn).Encode(request{Op: op}); err != nil {
		return response{}, ErrNoSession
	}
	var resp response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return response{}, ErrNoSession
	}
	return resp, nil
}

// Get returns the session key and vault salt, or ErrNoSession.
func Get(path string) (key, salt []byte, err error) {
	resp, err := call(path, "get")
	if err != nil {
		return nil, nil, err
	}
	if resp.Err != "" || len(resp.Key) == 0 {
		return nil, nil, ErrNoSession
	}
	return resp.Key, resp.Salt, nil
}

// Lock ends the session. It reports false (and no error) if there was none.
func Lock(path string) (bool, error) {
	resp, err := call(path, "lock")
	if err != nil {
		return false, nil
	}
	if !resp.OK {
		return false, fmt.Errorf("session: cerrar sesión: %s", resp.Err)
	}
	return true, nil
}
