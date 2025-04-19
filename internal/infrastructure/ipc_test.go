package infrastructure

import (
	"bufio"
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

func TestSendBusinessMetricsFallback(t *testing.T) {
	t.Run("falls back to ReportMetrics when buffer full", func(t *testing.T) {
		sock := filepath.Join(os.TempDir(), "ipc_fb.sock")
		os.Remove(sock)
		lnAddr, err := net.ResolveUnixAddr("unix", sock)
		require.NoError(t, err)
		ln, err := net.ListenUnix("unix", lnAddr)
		require.NoError(t, err)
		defer func() { ln.Close(); os.Remove(sock) }()

		m := NewIPCManager(sock, 0, 1, &spyAggregator{})
		m.SendBusinessMetricsUpdate(metrics.MetricsUpdate{Endpoint: "/b1"})

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

		m.SendBusinessMetricsUpdate(metrics.MetricsUpdate{Endpoint: "/b2", HTTPRequestsDelta: 2})
		select {
		case line := <-ch:
			require.Contains(t, line, `"endpoint":"/b2"`)
		case <-time.After(50 * time.Millisecond):
			t.Fatal("expected fallback ReportMetrics to write to socket")
		}
	})
}
