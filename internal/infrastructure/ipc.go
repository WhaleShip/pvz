package infrastructure

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
	"runtime"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/whaleship/pvz/internal/metrics"
)

type MetricsAggregator interface {
	UpdateMetrics(update metrics.MetricsUpdate)
}

type IPCManager struct {
	techMetricsCh     chan metrics.MetricsUpdate
	businessMetricsCh chan metrics.MetricsUpdate

	ipcConn       *net.UnixConn
	ipcConnMutex  sync.Mutex
	workerCount   int
	ipcSocketPath string

	aggregator MetricsAggregator
}

func NewIPCManager(socketPath string, techBuffer, businessBuffer int, aggregator MetricsAggregator) *IPCManager {
	manager := &IPCManager{
		techMetricsCh:     make(chan metrics.MetricsUpdate, techBuffer),
		businessMetricsCh: make(chan metrics.MetricsUpdate, businessBuffer),
		ipcSocketPath:     socketPath,
		aggregator:        aggregator,
	}
	manager.workerCount = runtime.NumCPU() / 2
	if fiber.IsChild() {
		manager.workerCount = 1
	}
	return manager
}

func (m *IPCManager) getIPCConn() *net.UnixConn {
	m.ipcConnMutex.Lock()
	defer m.ipcConnMutex.Unlock()
	if m.ipcConn == nil {
		addr, err := net.ResolveUnixAddr("unix", m.ipcSocketPath)
		if err != nil {
			log.Printf("ipc: error getting address: %v", err)
			return nil
		}
		conn, err := net.DialUnix("unix", nil, addr)
		if err != nil {
			log.Printf("ipc: establishing connection error: %v", err)
			return nil
		}
		m.ipcConn = conn
		log.Println("ipc: connection established")
	}
	return m.ipcConn
}

func (m *IPCManager) ReportMetrics(update metrics.MetricsUpdate) {
	data, err := json.Marshal(update)
	if err != nil {
		log.Printf("ipc: JSON serialization error: %v", err)
		return
	}
	data = append(data, '\n')

	conn := m.getIPCConn()
	if conn == nil {
		return
	}

	m.ipcConnMutex.Lock()
	defer m.ipcConnMutex.Unlock()
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("ipc: error sending data (%v), reconnetcting", err)
		conn.Close()
		m.ipcConn = nil
	}
}

func (m *IPCManager) StartServer() {
	os.Remove(m.ipcSocketPath)
	addr, err := net.ResolveUnixAddr("unix", m.ipcSocketPath)
	if err != nil {
		log.Fatalf("ipc: error getting address: %v", err)
	}
	ln, err := net.ListenUnix("unix", addr)
	if err != nil {
		log.Fatalf("ipc: listening error: %v", err)
	}
	log.Printf("ipc server started on %s", m.ipcSocketPath)
	go func() {
		for {
			conn, err := ln.AcceptUnix()
			if err != nil {
				log.Printf("ipc: connection acceptance error: %v", err)
				continue
			}
			go m.handleIPCConnection(conn)
		}
	}()
}

func (m *IPCManager) handleIPCConnection(conn *net.UnixConn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		var update metrics.MetricsUpdate
		if err := json.Unmarshal(line, &update); err != nil {
			log.Printf("ipc: JSON parsing error: %v", err)
			continue
		}
		m.aggregator.UpdateMetrics(update)
		// log.Printf("ipc: received an update %+v", update)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("ipc: error reading from connection: %v", err)
	}
}

func (m *IPCManager) StartSender() {
	for i := 0; i < m.workerCount; i++ {
		go func() {
			for update := range m.techMetricsCh {
				m.ReportMetrics(update)
			}
		}()
	}
	for i := 0; i < m.workerCount; i++ {
		go func() {
			for update := range m.businessMetricsCh {
				m.ReportMetrics(update)
			}
		}()
	}
	go m.getIPCConn()
}

func (m *IPCManager) SendTechMetricsUpdate(update metrics.MetricsUpdate) {
	select {
	case m.techMetricsCh <- update:
	default:
		log.Printf("skipping updating technical metrics: %+v", update)
	}
}

func (m *IPCManager) SendBusinessMetricsUpdate(update metrics.MetricsUpdate) {
	select {
	case m.businessMetricsCh <- update:
	default:
		go m.ReportMetrics(update) // немного кринж но иначе я не придумал как
	}
}
