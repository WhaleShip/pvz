package infrastructure

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/whaleship/pvz/internal/metrics"
)

type spyAggregator struct {
	received metrics.MetricsUpdate
}

func (s *spyAggregator) UpdateMetrics(u metrics.MetricsUpdate) {
	s.received = u
}

func TestReportMetrics(t *testing.T) {
	t.Run("sending success", func(t *testing.T) {
		sock := filepath.Join(os.TempDir(), "report.sock")
		os.Remove(sock)
		lnAddr, err := net.ResolveUnixAddr("unix", sock)
		require.NoError(t, err)
		ln, err := net.ListenUnix("unix", lnAddr)
		require.NoError(t, err)
		defer ln.Close()

		m := NewIPCManager(sock, 0, 0, nil)
		update := metrics.MetricsUpdate{Endpoint: "/test", HTTPRequestsDelta: 5}

		ch := make(chan string, 1)
		go func() {
			conn, err := ln.AcceptUnix()
			require.NoError(t, err)
			defer conn.Close()
			scanner := bufio.NewScanner(conn)
			if scanner.Scan() {
				ch <- scanner.Text()
			}
		}()

		m.ReportMetrics(update)
		line := <-ch
		var received metrics.MetricsUpdate
		err = json.Unmarshal([]byte(line), &received)
		require.NoError(t, err)
		require.Equal(t, update, received)
	})

	t.Run("sending error", func(t *testing.T) {
		sock := filepath.Join(os.TempDir(), "report_err.sock")
		os.Remove(sock)
		m := NewIPCManager(sock, 0, 0, nil)
		conn := m.getIPCConn()
		require.Nil(t, conn)

		update := metrics.MetricsUpdate{Endpoint: "/error"}
		m.ReportMetrics(update)
	})

	t.Run("no socket", func(t *testing.T) {
		m := NewIPCManager("/nonexistent.sock", 0, 0, nil)
		m.ReportMetrics(metrics.MetricsUpdate{Endpoint: "/no-sock"})
		require.Nil(t, m.ipcConn)
	})
}

func TestSendTechMetricsUpdate(t *testing.T) {
	m := NewIPCManager("", 1, 0, nil)

	t.Run("enqueues until buffer full", func(t *testing.T) {
		u1 := metrics.MetricsUpdate{Endpoint: "/t1"}
		m.SendTechMetricsUpdate(u1)
		require.Len(t, m.techMetricsCh, 1)
	})

	t.Run("skips when buffer full", func(t *testing.T) {
		u2 := metrics.MetricsUpdate{Endpoint: "/t2"}
		m.SendTechMetricsUpdate(u2)
		require.Len(t, m.techMetricsCh, 1)
	})
}

func TestSendBusinessMetricsUpdateChaining(t *testing.T) {
	t.Run("enqueues until buffer not full", func(t *testing.T) {
		m := NewIPCManager("", 0, 2, nil)
		u1 := metrics.MetricsUpdate{Endpoint: "/b1"}
		u2 := metrics.MetricsUpdate{Endpoint: "/b2"}

		m.SendBusinessMetricsUpdate(u1)
		require.Len(t, m.businessMetricsCh, 1)
		m.SendBusinessMetricsUpdate(u2)
		require.Len(t, m.businessMetricsCh, 2)
	})

	t.Run("fallback when buffer full", func(t *testing.T) {
		sock := filepath.Join(os.TempDir(), "ipc_fb2.sock")
		os.Remove(sock)
		lnAddr, err := net.ResolveUnixAddr("unix", sock)
		require.NoError(t, err)
		ln, err := net.ListenUnix("unix", lnAddr)
		require.NoError(t, err)
		defer func() { ln.Close(); os.Remove(sock) }()

		agg := &spyAggregator{}
		m := NewIPCManager(sock, 0, 1, agg)
		m.SendBusinessMetricsUpdate(metrics.MetricsUpdate{Endpoint: "/keep"})
		ch := make(chan string, 1)
		go func() {
			conn, err := ln.AcceptUnix()
			require.NoError(t, err)
			defer conn.Close()
			scanner := bufio.NewScanner(conn)
			if scanner.Scan() {
				ch <- scanner.Text()
			}
		}()

		m.SendBusinessMetricsUpdate(metrics.MetricsUpdate{Endpoint: "/fallback", HTTPRequestsDelta: 3})
		select {
		case line := <-ch:
			require.Contains(t, line, `"endpoint":"/fallback"`)
			require.Contains(t, line, `"http_requests_delta":3`)
		case <-time.After(50 * time.Millisecond):
			t.Fatal("expected fallback via ReportMetrics")
		}
	})
}

func TestStartServerHandleIPC(t *testing.T) {
	t.Run("full pipeline: StartServer → handle → UpdateMetrics", func(t *testing.T) {
		sock := filepath.Join(os.TempDir(), "ipc_handle.sock")
		os.Remove(sock)
		agg := &spyAggregator{}
		m := NewIPCManager(sock, 0, 0, agg)
		m.StartServer()
		time.Sleep(20 * time.Millisecond)

		addr := &net.UnixAddr{Name: sock, Net: "unix"}
		conn, err := net.DialUnix("unix", nil, addr)
		require.NoError(t, err)
		defer conn.Close()

		update := metrics.MetricsUpdate{Endpoint: "/h", HTTPRequestsDelta: 5}
		payload, err := json.Marshal(update)
		require.NoError(t, err)
		_, err = conn.Write(append(payload, '\n'))
		require.NoError(t, err)

		time.Sleep(20 * time.Millisecond)
		require.Equal(t, update, agg.received)
	})
}
func TestNewIPCManager(t *testing.T) {
	aggregator := &spyAggregator{}
	m := NewIPCManager("/tmp/test.sock", 10, 20, aggregator)

	require.NotNil(t, m.techMetricsCh)
	require.NotNil(t, m.businessMetricsCh)
	require.Equal(t, 10, cap(m.techMetricsCh))
	require.Equal(t, 20, cap(m.businessMetricsCh))

	require.Equal(t, aggregator, m.aggregator)

	if fiber.IsChild() {
		require.Equal(t, 1, m.workerCount)
	} else {
		expected := runtime.NumCPU() / 2
		require.Equal(t, expected, m.workerCount)
	}
}
func TestGetIPCConn(t *testing.T) {
	t.Run("conn success", func(t *testing.T) {
		sock := filepath.Join(os.TempDir(), "test.sock")
		os.Remove(sock)
		lnAddr, err := net.ResolveUnixAddr("unix", sock)
		require.NoError(t, err)
		ln, err := net.ListenUnix("unix", lnAddr)
		require.NoError(t, err)
		defer ln.Close()

		m := NewIPCManager(sock, 0, 0, nil)
		conn := m.getIPCConn()
		require.NotNil(t, conn)
		require.Equal(t, conn, m.ipcConn)

		conn2 := m.getIPCConn()
		require.Equal(t, conn, conn2)
	})

	t.Run("conn error", func(t *testing.T) {
		m := NewIPCManager("/invalid.sock", 0, 0, nil)
		conn := m.getIPCConn()
		require.Nil(t, conn)
	})
}
func TestStartServer(t *testing.T) {
	sock := filepath.Join(os.TempDir(), "server.sock")
	os.Remove(sock)
	aggregator := &spyAggregator{}
	m := NewIPCManager(sock, 0, 0, aggregator)
	m.StartServer()
	time.Sleep(20 * time.Millisecond)

	addr := &net.UnixAddr{Name: sock, Net: "unix"}
	conn, err := net.DialUnix("unix", nil, addr)
	require.NoError(t, err)
	defer conn.Close()

	update := metrics.MetricsUpdate{Endpoint: "/start", HTTPRequestsDelta: 1}
	payload, _ := json.Marshal(update)
	_, err = conn.Write(append(payload, '\n'))
	require.NoError(t, err)

	time.Sleep(20 * time.Millisecond)
	require.Equal(t, update, aggregator.received)
}

func TestStartSender(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sock := filepath.Join(os.TempDir(), "sender.sock")
		os.Remove(sock)
		lnAddr, err := net.ResolveUnixAddr("unix", sock)
		require.NoError(t, err)
		ln, err := net.ListenUnix("unix", lnAddr)
		require.NoError(t, err)
		defer ln.Close()

		aggregator := &spyAggregator{}
		m := NewIPCManager(sock, 1, 1, aggregator)
		m.StartSender()

		update := metrics.MetricsUpdate{Endpoint: "/sender", HTTPRequestsDelta: 3}
		m.techMetricsCh <- update

		conn, err := ln.AcceptUnix()
		require.NoError(t, err)
		defer conn.Close()

		scanner := bufio.NewScanner(conn)
		require.True(t, scanner.Scan())
		line := scanner.Text()
		var received metrics.MetricsUpdate
		err = json.Unmarshal([]byte(line), &received)
		require.NoError(t, err)
		require.Equal(t, update, received)
	})
	t.Run("buisness metric chanal", func(t *testing.T) {
		dir := os.TempDir()
		sock := filepath.Join(dir, fmt.Sprintf("snd_%d.sock", time.Now().UnixNano()))
		lnAddr, err := net.ResolveUnixAddr("unix", sock)
		require.NoError(t, err)
		ln, err := net.ListenUnix("unix", lnAddr)
		require.NoError(t, err)
		defer ln.Close()

		m := NewIPCManager(sock, 1, 1, nil)
		m.StartSender()

		upd := metrics.MetricsUpdate{Endpoint: "/biz", HTTPRequestsDelta: 9}
		m.businessMetricsCh <- upd

		conn, err := ln.AcceptUnix()
		require.NoError(t, err)
		defer conn.Close()

		sc := bufio.NewScanner(conn)
		require.True(t, sc.Scan())
		var got metrics.MetricsUpdate
		require.NoError(t, json.Unmarshal(sc.Bytes(), &got))
		require.Equal(t, upd, got)
	})
}

func TestHandleIPCConnection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		sockPath := filepath.Join(dir, "ipc.sock")
		addr := &net.UnixAddr{Name: sockPath, Net: "unix"}

		listener, err := net.ListenUnix("unix", addr)
		require.NoError(t, err)
		defer listener.Close()

		aggregator := &spyAggregator{}
		m := NewIPCManager("", 0, 0, aggregator)

		go func() {
			conn, err := listener.AcceptUnix()
			require.NoError(t, err)
			m.handleIPCConnection(conn)
		}()

		client, err := net.DialUnix("unix", nil, addr)
		require.NoError(t, err)
		defer client.Close()

		update := metrics.MetricsUpdate{Endpoint: "/handle", HTTPRequestsDelta: 2}
		payload, _ := json.Marshal(update)
		_, err = client.Write(append(payload, '\n'))
		require.NoError(t, err)

		time.Sleep(20 * time.Millisecond)
		require.Equal(t, update, aggregator.received)
	})
	t.Run("error UnixConn deserialize", func(t *testing.T) {
		dir := os.TempDir()
		sockPath := filepath.Join(dir, fmt.Sprintf("ipc_%d.sock", time.Now().UnixNano()))
		addr := &net.UnixAddr{Name: sockPath, Net: "unix"}

		listener, err := net.ListenUnix("unix", addr)
		require.NoError(t, err)
		defer listener.Close()

		aggregator := &spyAggregator{}
		m := NewIPCManager("", 0, 0, aggregator)

		go func() {
			conn, err := listener.AcceptUnix()
			require.NoError(t, err)
			m.handleIPCConnection(conn)
		}()

		client, err := net.DialUnix("unix", nil, addr)
		require.NoError(t, err)
		defer client.Close()

		_, err = client.Write([]byte("invalid json\n"))
		require.NoError(t, err)

		time.Sleep(20 * time.Millisecond)
		require.Equal(t, metrics.MetricsUpdate{}, aggregator.received)
	})

}
