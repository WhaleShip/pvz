//go:build integration
// +build integration

package integration_tests

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	"github.com/whaleship/pvz/internal/app"
	"github.com/whaleship/pvz/internal/database"
)

var (
	handler fasthttp.RequestHandler
)

func runMigrations(ctx context.Context, pool database.PgxIface) error {
	data, err := os.ReadFile("../../migrations/init.sql")
	if err != nil {
		return fmt.Errorf("reading migration file: %w", err)
	}
	if _, err := pool.Exec(ctx, string(data)); err != nil {
		return fmt.Errorf("executing migrations: %w", err)
	}
	return nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	const (
		dbUser     = "testuser"
		dbPassword = "testpass"
		dbName     = "testdb"
		sslMode    = "disabled"
	)

	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPassword,
			"POSTGRES_DB":       dbName,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	dbC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Println("failed to start postgres container:", err)
		os.Exit(1)
	}
	defer dbC.Terminate(ctx)

	host, err := dbC.Host(ctx)
	if err != nil {
		fmt.Println("failed to get container host:", err)
		os.Exit(1)
	}
	port, err := dbC.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Println("failed to get container port:", err)
		os.Exit(1)
	}

	os.Setenv("DB_HOST", host)
	os.Setenv("DB_PORT", port.Port())
	os.Setenv("DB_USER", dbUser)
	os.Setenv("DB_PASSWORD", dbPassword)
	os.Setenv("DB_NAME", dbName)
	os.Setenv("SSL_MODE", sslMode)

	pvzApp := app.New(false)
	pvzApp.InitDBConnection()

	pool := pvzApp.GetDBConn()
	if err := runMigrations(ctx, pool); err != nil {
		fmt.Println("migration error:", err)
		os.Exit(1)
	}

	pvzApp.InitializeMetrics()
	pvzApp.InitializeHTTPServer()
	handler = pvzApp.PVZ.Handler()

	exitCode := m.Run()
	dbC.Terminate(ctx)
	os.Exit(exitCode)
}

func TestFullWorkflow(t *testing.T) {
	ln := fasthttputil.NewInmemoryListener()
	go fasthttp.Serve(ln, handler)
	t.Cleanup(func() { ln.Close() })

	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return ln.Dial()
			},
		},
	}

	expect := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://localhost",
		Reporter: httpexpect.NewAssertReporter(t),
		Client:   httpClient,
	})

	modRaw := expect.POST("/dummyLogin").
		WithJSON(map[string]string{"role": "moderator"}).
		Expect().Status(http.StatusOK).Body().Raw()
	modToken := strings.Trim(modRaw, `"`)
	headersMod := map[string]string{"Authorization": fmt.Sprintf("Bearer %s", modToken)}

	var createdPVZ struct {
		ID string `json:"id"`
	}
	expect.POST("/pvz").WithHeaders(headersMod).
		WithJSON(map[string]string{"city": "Москва"}).
		Expect().Status(http.StatusCreated).
		JSON().Object().Decode(&createdPVZ)

	empRaw := expect.POST("/dummyLogin").
		WithJSON(map[string]string{"role": "employee"}).
		Expect().Status(http.StatusOK).Body().Raw()
	empToken := strings.Trim(empRaw, `"`)
	headersEmp := map[string]string{"Authorization": fmt.Sprintf("Bearer %s", empToken)}

	var createdRec struct {
		ID string `json:"id"`
	}
	expect.POST("/receptions").WithHeaders(headersEmp).
		WithJSON(map[string]string{"pvzId": createdPVZ.ID}).
		Expect().Status(http.StatusCreated).
		JSON().Object().Decode(&createdRec)

	for i := 0; i < 50; i++ {
		expect.POST("/products").WithHeaders(headersEmp).
			WithJSON(map[string]string{"type": "электроника", "pvzId": createdPVZ.ID}).
			Expect().Status(http.StatusCreated)
	}

	closeURL := fmt.Sprintf("/pvz/%s/close_last_reception", createdPVZ.ID)
	var closedRec struct {
		ID string `json:"id"`
	}
	expect.POST(closeURL).WithHeaders(headersEmp).
		Expect().Status(http.StatusOK).
		JSON().Object().Decode(&closedRec)

	assert.Equal(t, createdRec.ID, closedRec.ID)
}
