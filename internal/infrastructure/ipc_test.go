package infrastructure

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	t.Run("writes JSON line to socket", func(t *testing.T) {
		sock := filepath.Join(os.TempDir(), "ipc_test.sock")
		os.Remove(sock)
		lnAddr, err := net.ResolveUnixAddr("unix", sock)
		require.NoError(t, err)
		ln, err := net.ListenUnix("unix", lnAddr)
		require.NoError(t, err)
		defer func() { ln.Close(); os.Remove(sock) }()

		agg := &spyAggregator{}
		m := NewIPCManager(sock, 0, 0, agg)

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

		update := metrics.MetricsUpdate{Endpoint: "/y", HTTPRequestsDelta: 10}
		m.ReportMetrics(update)

		line := <-ch
		require.Contains(t, line, `"endpoint":"/y"`)
		require.Contains(t, line, `"http_requests_delta":10`)
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
