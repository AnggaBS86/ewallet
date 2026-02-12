package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"ewallet/internal/config"
	"ewallet/internal/db"
	"ewallet/internal/router"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
)

func TestFullFlow(t *testing.T) {
	testDSN := resolveTestDSN()
	if testDSN == "" {
		t.Skip("TEST_DB_DSN or TEST_DB_* not set")
	}

	migrationsPath := projectRoot(t, "migrations")
	resetDB(t, testDSN, migrationsPath)

	cfg := config.Config{
		Port:      "0",
		JWTSecret: "testsecret",
		DBDSN:     testDSN,
	}

	database, err := db.New(cfg)
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}
	sqlDB, _ := database.DB()
	defer sqlDB.Close()

	e := echo.New()
	router.Register(e, database, cfg)

	register(t, e, "Alice", "alice@example.com", "password123")
	register(t, e, "Bob", "bob@example.com", "password123")

	token := login(t, e, "alice@example.com", "password123")

	topUp(t, e, token, 50000)
	transfer(t, e, token, "bob@example.com", 10000)

	history := getHistory(t, e, token)
	assertHistory(t, history)
}

func resolveTestDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("TEST_DB_DSN")); dsn != "" {
		return dsn
	}

	host := strings.TrimSpace(os.Getenv("TEST_DB_HOST"))
	port := strings.TrimSpace(os.Getenv("TEST_DB_PORT"))
	user := strings.TrimSpace(os.Getenv("TEST_DB_USER"))
	password := os.Getenv("TEST_DB_PASSWORD")
	name := strings.TrimSpace(os.Getenv("TEST_DB_NAME"))
	sslmode := strings.TrimSpace(os.Getenv("TEST_DB_SSLMODE"))
	if sslmode == "" {
		sslmode = "disable"
	}

	if host == "" || port == "" || user == "" || name == "" {
		return ""
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, name, sslmode)
}

func projectRoot(t *testing.T, rel string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	root := filepath.Dir(filepath.Dir(filename))
	return filepath.Join(root, rel)
}

func resetDB(t *testing.T, dsn, migrationsPath string) {
	t.Helper()
	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		t.Fatalf("migrate init failed: %v", err)
	}
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate down failed: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up failed: %v", err)
	}
}

func register(t *testing.T, e *echo.Echo, name, email, password string) {
	t.Helper()
	body := map[string]string{"name": name, "email": email, "password": password}
	resp := doJSON(t, e, http.MethodPost, "/api/auth/register", "", body)
	if resp.Status != "OK" {
		t.Fatalf("register failed: %v", resp.Message)
	}
}

func login(t *testing.T, e *echo.Echo, email, password string) string {
	t.Helper()
	body := map[string]string{"email": email, "password": password}
	resp := doJSON(t, e, http.MethodPost, "/api/auth/login", "", body)
	if resp.Status != "OK" {
		t.Fatalf("login failed: %v", resp.Message)
	}
	token, _ := resp.Data.(map[string]interface{})["token"].(string)
	if token == "" {
		t.Fatalf("missing token in login response")
	}
	return token
}

func topUp(t *testing.T, e *echo.Echo, token string, amount int64) {
	t.Helper()
	body := map[string]int64{"amount": amount}
	resp := doJSON(t, e, http.MethodPost, "/api/wallets/topup", token, body)
	if resp.Status != "OK" {
		t.Fatalf("topup failed: %v", resp.Message)
	}
}

func transfer(t *testing.T, e *echo.Echo, token, receiverEmail string, amount int64) {
	t.Helper()
	body := map[string]interface{}{"receiver_email": receiverEmail, "amount": amount}
	resp := doJSON(t, e, http.MethodPost, "/api/transactions/transfer", token, body)
	if resp.Status != "OK" {
		t.Fatalf("transfer failed: %v", resp.Message)
	}
}

func getHistory(t *testing.T, e *echo.Echo, token string) responseEnvelope {
	t.Helper()
	return doJSON(t, e, http.MethodGet, "/api/transactions/history?limit=50", token, nil)
}

func assertHistory(t *testing.T, resp responseEnvelope) {
	t.Helper()
	if resp.Status != "OK" {
		t.Fatalf("history failed: %v", resp.Message)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("history data invalid")
	}

	list, ok := data["transactions"].([]interface{})
	if !ok || len(list) == 0 {
		t.Fatalf("history transactions empty")
	}

	item, ok := list[0].(map[string]interface{})
	if !ok {
		t.Fatalf("history item invalid")
	}

	if _, ok := item["sender_id"]; ok {
		t.Fatalf("sender_id should not be present")
	}
	if _, ok := item["receiver_id"]; ok {
		t.Fatalf("receiver_id should not be present")
	}

	sender, ok := item["sender"].(map[string]interface{})
	if !ok || sender["email"] == "" {
		t.Fatalf("sender info missing")
	}
	receiver, ok := item["receiver"].(map[string]interface{})
	if !ok || receiver["email"] == "" {
		t.Fatalf("receiver info missing")
	}
}

type responseEnvelope struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func doJSON(t *testing.T, e *echo.Echo, method, path, token string, body interface{}) responseEnvelope {
	t.Helper()
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json marshal failed: %v", err)
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var resp responseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response json: %v", err)
	}
	return resp
}
