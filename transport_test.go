package gofast

import "testing"
import "fmt"
import "compress/flate"
import "net"
import "time"
import "sync"

func TestTransport(t *testing.T) {
	st, end := tagOpaqueStart, tagOpaqueStart+10
	config := newconfig("testtransport", st, end)
	config["tags"] = "gzip"
	tconn := newTestConnection(nil, true)
	trans, err := NewTransport(tconn, testVersion(1), nil, config)
	if err != nil {
		t.Error(err)
	}
	trans.VersionHandler(testVerhandler).Handshake()
	if _, ok := trans.tagenc[tagGzip]; !ok && len(trans.tagenc) != 1 {
		t.Errorf("expected gzip, got %v", trans.tagenc)
	}
	if ver := trans.peerver.Value().(int); ver != 1 {
		t.Errorf("expected 1, got %v", ver)
	}
	trans.Close()
	time.Sleep(1 * time.Second)
}

func newconfig(name string, start, end int) map[string]interface{} {
	return map[string]interface{}{
		"name":         name,
		"buffersize":   1024 * 1024 * 10,
		"chansize":     1,
		"batchsize":    1,
		"tags":         "",
		"opaque.start": start,
		"opaque.end":   end,
		"log.level":    "error",
		"gzip.file":    flate.BestSpeed,
	}
}

type testConnection struct {
	roff  int
	woff  int
	buf   []byte
	read  bool
	mu    sync.Mutex
	laddr netAddr
	raddr netAddr
}

func newTestConnection(buf []byte, read bool) *testConnection {
	tconn := &testConnection{
		laddr: netAddr("127.0.0.1:9998"),
		raddr: netAddr("127.0.0.1:9999"),
		read:  read,
	}
	if tconn.buf = buf; buf == nil {
		tconn.buf = make([]byte, 100000)
	}
	return tconn
}

func (tc *testConnection) Write(b []byte) (n int, err error) {
	do := func() (int, error) {
		tc.mu.Lock()
		defer tc.mu.Unlock()
		newoff := tc.woff + len(b)
		copy(tc.buf[tc.woff:newoff], b)
		tc.woff = newoff
		return len(b), nil
	}
	n, err = do()
	//fmt.Println("write ...", n, err)
	return
}

func (tc *testConnection) Read(b []byte) (n int, err error) {
	do := func() (int, error) {
		tc.mu.Lock()
		defer tc.mu.Unlock()
		if newoff := tc.roff + len(b); newoff <= tc.woff {
			copy(b, tc.buf[tc.roff:newoff])
			tc.roff = newoff
			return len(b), nil
		}
		return 0, nil
	}
	for err == nil && n == 0 {
		if tc.read {
			n, err = do()
		}
		time.Sleep(100 * time.Millisecond)
	}
	//fmt.Println("read ...", n, err)
	return n, err
}

func (tc *testConnection) LocalAddr() net.Addr {
	return tc.laddr
}

func (tc *testConnection) RemoteAddr() net.Addr {
	return tc.raddr
}

func (tc *testConnection) reset() *testConnection {
	tc.woff, tc.roff = 0, 0
	return tc
}

func (tx *testConnection) Close() error {
	return nil
}

type netAddr string

func (addr netAddr) Network() string {
	return "tcp"
}

func (addr netAddr) String() string {
	return string(addr)
}

type testVersion int

func (v testVersion) Less(ver Version) bool {
	return v < ver.(testVersion)
}

func (v testVersion) Equal(ver Version) bool {
	return v == ver.(testVersion)
}

func (v testVersion) String() string {
	return fmt.Sprintf("%v", int(v))
}

func (v testVersion) Value() interface{} {
	return int(v)
}

func testVerhandler(val interface{}) Version {
	if ver, ok := val.(uint64); ok {
		return testVersion(ver)
	} else if ver, ok := val.(int); ok {
		return testVersion(ver)
	}
	return nil
}
