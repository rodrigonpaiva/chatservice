package configs

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfig_ParsesEmptyStopAsEmptySlice(t *testing.T) {
	dir := writeTestEnv(t, "STOP=[]\n")
	resetViper()

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if !reflect.DeepEqual(cfg.Stop, []string{}) {
		t.Fatalf("expected empty stop slice, got %v", cfg.Stop)
	}
}

func TestLoadConfig_ParsesMultipleStopsFromJSON(t *testing.T) {
	dir := writeTestEnv(t, "STOP=[\"END\",\"User:\"]\n")
	resetViper()

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	expected := []string{"END", "User:"}
	if !reflect.DeepEqual(cfg.Stop, expected) {
		t.Fatalf("expected stop %v, got %v", expected, cfg.Stop)
	}
}

func TestLoadConfig_ReturnsErrorForInvalidStopFormat(t *testing.T) {
	dir := writeTestEnv(t, "STOP=A,B\n")
	resetViper()

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected invalid STOP format to return an error")
	}

	if !strings.Contains(err.Error(), "invalid STOP config") {
		t.Fatalf("expected invalid STOP error, got %v", err)
	}
}

func writeTestEnv(t *testing.T, extra string) string {
	t.Helper()

	dir := t.TempDir()
	content := strings.Join([]string{
		"DB_DRIVER=mysql",
		"DB_HOST=localhost",
		"DB_PORT=3306",
		"DB_USER=root",
		"DB_PASSWORD=root",
		"DB_NAME=chat_app",
		"WEB_SERVER_PORT=8080",
		"GRPC_SERVER_PORT=50051",
		"INITIAL_CHAT_MESSAGE=You are a helpful assistant.",
		"OPENAI_API_KEY=test-key",
		"MODEL=gpt-3.5-turbo",
		"MODEL_MAX_TOKENS=4096",
		"TEMPERATURE=0.7",
		"TOP_P=1",
		"N=1",
		"MAX_TOKENS=512",
		"AUTH_TOKEN=test-token",
		strings.TrimSpace(extra),
		"",
	}, "\n")

	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	return dir
}

func resetViper() {
	viper.Reset()
}
