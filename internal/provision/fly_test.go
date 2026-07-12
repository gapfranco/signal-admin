package provision_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"signal-admin/internal/provision"
)

func TestParseAntenaConf(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "narnia-antena.conf")
	content := "ADDR=:4000\nTURSO_URL=libsql://narnia-gapfranco.aws-us-east-1.turso.io\nTURSO_TOKEN=secret-token\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	conf, err := provision.ParseAntenaConf(path)
	if err != nil {
		t.Fatalf("ParseAntenaConf: %v", err)
	}
	if conf.TursoURL != "libsql://narnia-gapfranco.aws-us-east-1.turso.io" {
		t.Fatalf("TursoURL = %q", conf.TursoURL)
	}
	if conf.TursoToken != "secret-token" {
		t.Fatalf("TursoToken = %q", conf.TursoToken)
	}
}

func TestParseAntenaConfMissingToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.conf")
	if err := os.WriteFile(path, []byte("TURSO_URL=libsql://x\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := provision.ParseAntenaConf(path); err == nil {
		t.Fatal("expected error for missing TURSO_TOKEN")
	}
}

func TestMaterializeFlyToml(t *testing.T) {
	example := []byte(`app = "antena-<dbname>"
primary_region = "iad"

[build]
  image = "ghcr.io/gapfranco/antena:latest"
`)
	got := string(provision.MaterializeFlyToml(example, "narnia"))
	if strings.Contains(got, "<dbname>") {
		t.Fatalf("placeholder not replaced: %s", got)
	}
	if !strings.Contains(got, `app = "antena-narnia"`) {
		t.Fatalf("app name missing: %s", got)
	}
	if strings.Count(got, "antena-narnia") != 1 {
		t.Fatalf("unexpected replacements: %s", got)
	}
}

func TestFlyFlashMessages(t *testing.T) {
	if got := provision.FlyInstallFlashMessage("narnia"); !strings.Contains(got, "antena-narnia") {
		t.Fatalf("success flash = %q", got)
	}
	if got := provision.FlyInstallErrorFlashMessage("boom"); !strings.Contains(got, "boom") {
		t.Fatalf("error flash = %q", got)
	}
	if got := provision.FlyAlreadyInstalledFlashMessage(); got == "" {
		t.Fatal("empty already-installed flash")
	}
}
