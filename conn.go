// Copyright 2012, 2013 Gary Burd & Zhang Peihao
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package websocket implements the WebSocket protocol defined in RFC 6455.
//
// The websocket package passes UTF-8 text to and from the network without
// validation. It is the application's responsibility to validate the contents
// of text messages.

package zwebsocket

import (
	"github.com/garyburd/go-websocket/websocket"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Implement the net.Conn interface.
// All data are transfered in binary stream.
type Conn struct {
	ws *websocket.Conn
	r  io.Reader
}

// Connect a web socket hosr, and upgrade to web socket.
//
// Examples:
//	Connect("http://localhost:8081/websocket", 1024, 1024)
func Connect(urlstr string, readBufSize, writeBufSize int) (conn *Conn, resp *http.Response, err error) {
	var u *url.URL
	var ws *websocket.Conn
	var c net.Conn
	if u, err = url.Parse(urlstr); err != nil {
		return
	}
	if c, err = net.Dial("tcp", u.Host); err != nil {
		return
	}
	if ws, resp, err = websocket.NewClient(c, u, http.Header{"Origin": {urlstr}},
		readBufSize, writeBufSize); err != nil {
		c.Close()
		return
	}
	conn = &Conn{
		ws: ws,
	}
	return
}

// Create a server side connection.
func NewConn(w http.ResponseWriter, r *http.Request, responseHeader http.Header,
	readBufSize, writeBufSize int) (conn *Conn, err error) {
	var ws *websocket.Conn
	if ws, err = websocket.Upgrade(w, r.Header, responseHeader, readBufSize,
		writeBufSize); err != nil {
		return
	}
	conn = &Conn{
		ws: ws,
	}
	return
}

// Read reads data from the connection.
// Read can be made to time out and return a Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (conn *Conn) Read(b []byte) (n int, err error) {
	var opCode int
	if conn.r == nil {
		// New message
		var r io.Reader
		for {
			if opCode, r, err = conn.ws.NextReader(); err != nil {
				return
			}
			if opCode != websocket.OpBinary && opCode != websocket.OpText {
				continue
			}

			conn.r = r
			break
		}
	}

	n, err = conn.r.Read(b)
	if err != nil {
		if err == io.EOF {
			// Message finished
			conn.r = nil
			err = nil
		}
	}
	return
}

// Write writes data to the connection.
// Write can be made to time out and return a Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (conn *Conn) Write(b []byte) (n int, err error) {
	var w io.WriteCloser
	if w, err = conn.ws.NextWriter(websocket.OpBinary); err != nil {
		return
	}
	if n, err = w.Write(b); err != nil {
		return
	}
	err = w.Close()
	return
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (conn *Conn) Close() error {
	return conn.ws.Close()
}

// LocalAddr returns the local network address.
func (conn *Conn) LocalAddr() net.Addr {
	return conn.ws.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (conn *Conn) RemoteAddr() net.Addr {
	return conn.ws.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future I/O, not just
// the immediately following call to Read or Write.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (conn *Conn) SetDeadline(t time.Time) (err error) {
	if err = conn.ws.SetReadDeadline(t); err != nil {
		return
	}
	return conn.ws.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls.
// A zero value for t means Read will not time out.
func (conn *Conn) SetReadDeadline(t time.Time) error {
	return conn.ws.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (conn *Conn) SetWriteDeadline(t time.Time) error {
	return conn.ws.SetWriteDeadline(t)
}
