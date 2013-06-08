// Copyright 2013 Gary Burd & Zhang Peihao
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

package zwebsocket

import (
	"bufio"
	"github.com/garyburd/go-websocket/websocket"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type wsBinaryHandler struct {
	*testing.T
}

var (
	g_remote, g_local net.Addr
)

func (t wsBinaryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		t.Logf("bad method: %s", r.Method)
		return
	}
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		t.Logf("bad origin: %s", r.Header.Get("Origin"))
		return
	}

	conn, err := NewConn(w, r, http.Header{"Set-Cookie": {"sessionId=1234"}}, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		t.Logf("bad handshake: %v", err)
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		t.Logf("upgrade error: %v", err)
		return
	}
	defer conn.Close()
	g_remote = conn.RemoteAddr()
	g_local = conn.LocalAddr()
	for {
		br := bufio.NewReader(conn)
		str, err := br.ReadString('\n')
		if err != nil {
			t.Logf("server ReadString error: %v", err)
		}

		if _, err = conn.Write([]byte(str)); err != nil {
			t.Logf("conn.Write error: %v", err)
			return
		}
	}
}

func TestBinaryConn(t *testing.T) {
	s := httptest.NewServer(wsBinaryHandler{t})
	defer s.Close()
	conn, resp, err := Connect(s.URL, 1024, 1024)
	if err != nil {
		t.Fatal("Connect err:", err)
	}

	defer conn.Close()

	var sessionId string
	for _, c := range resp.Cookies() {
		if c.Name == "sessionId" {
			sessionId = c.Value
		}
	}
	if sessionId != "1234" {
		t.Error("Set-Cookie not received from the server.")
	}

	if _, err = conn.Write([]byte("HELLO\n")); err != nil {
		t.Error("Write err:", err)
	}

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	var str string
	br := bufio.NewReader(conn)
	str, err = br.ReadString('\n')
	if err != nil {
		t.Fatalf("client ReadString: %v", err)
	}
	if str != "HELLO\n" {
		t.Fatalf("message=%s, want %s", str, "HELLO\n")
	}

	if g_remote == nil {
		t.Error("g_remote is nil")
	} else if g_remote.String() != conn.LocalAddr().String() || g_remote.Network() != conn.LocalAddr().Network() {
		t.Errorf("remote address: %v != %v\n", g_remote, conn.LocalAddr())
	}
	t.Log("remote OK")

	if g_local == nil {
		t.Error("g_local is nil")
	} else if g_local.String() != conn.RemoteAddr().String() || g_local.Network() != conn.RemoteAddr().Network() {
		t.Errorf("remote address: %v != %v\n", g_local, conn.RemoteAddr())
	}

}
