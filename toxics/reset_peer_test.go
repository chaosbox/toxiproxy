package toxics_test

import (
	"github.com/Shopify/toxiproxy"
	"github.com/Shopify/toxiproxy/toxics"
	"io"
	"net"
	"os"
	"syscall"
	"testing"
	"time"
)

const msg = "reset toxic payload\n"

func TestResetToxicNoTimeout(t *testing.T) {
	WithEchoProxy(t, func(conn net.Conn, response chan []byte, proxy *toxiproxy.Proxy) {
		addToxicAndWritePayload(t, conn, proxy, toxics.ResetToxic{}, "upstream")
		checkConnectionState(t, conn, false)
	})
}

func TestResetToxicWithTimeout(t *testing.T) {
	WithEchoProxy(t, func(conn net.Conn, response chan []byte, proxy *toxiproxy.Proxy) {
		resetToxic := toxics.ResetToxic{Timeout: 100}
		addToxicAndWritePayload(t, conn, proxy, resetToxic, "upstream")
		start := time.Now()
		checkConnectionState(t, conn, false)
		AssertDeltaTime(t, "Reset after timeout", time.Since(start), time.Duration(resetToxic.Timeout)*time.Millisecond, time.Duration(resetToxic.Timeout+10)*time.Millisecond)
	})
}

func TestResetToxicWithTimeoutDownstream(t *testing.T) {
	WithEchoProxy(t, func(conn net.Conn, response chan []byte, proxy *toxiproxy.Proxy) {
		resetToxic := toxics.ResetToxic{Timeout: 100}
		addToxicAndWritePayload(t, conn, proxy, resetToxic, "downstream")
		start := time.Now()
		checkConnectionState(t, conn, true)
		AssertDeltaTime(t, "Reset after timeout", time.Since(start), time.Duration(resetToxic.Timeout)*time.Millisecond, time.Duration(resetToxic.Timeout+10)*time.Millisecond)
	})
}

func addToxicAndWritePayload(t *testing.T, conn net.Conn, proxy *toxiproxy.Proxy, resetToxic toxics.ResetToxic, stream string) {
	if _, err := proxy.Toxics.AddToxicJson(ToxicToJson(t, "resetconn", "reset_peer", stream, &resetToxic)); err != nil {
		t.Error("AddToxicJson returned error:", err)
	}
	if _, err := conn.Write([]byte(msg)); err != nil {
		t.Error("Failed writing TCP payload", err)
	}
}

func checkConnectionState(t *testing.T, conn net.Conn, downstream bool) {
	tmp := make([]byte, len(msg))
	_, err := conn.Read(tmp)
	if downstream && err != io.EOF {
		t.Fatal("Expected: downstream - returns EOF")
	}
	if opErr, ok := err.(*net.OpError); ok {
		if syscallErr, ok := opErr.Err.(*os.SyscallError); ok {
			if !(syscallErr.Err == syscall.ECONNRESET) {
				t.Error("Expected: upstream - connection reset by peer. Got:", err)
			}
		}
	}
	_, err = conn.Read(tmp)
	if err != io.EOF {
		t.Fatal("expected EOF from closed connection")
	}
}
