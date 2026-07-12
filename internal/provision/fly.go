package provision

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	flyOrg          = "personal"
	flyImage        = "ghcr.io/gapfranco/antena:latest"
	flyTomlExample  = "fly.toml.example"
	flyPlaceholder  = "<dbname>"
)

// AntenaConf holds connection settings read from an antena conf file.
type AntenaConf struct {
	TursoURL   string
	TursoToken string
}

// ParseAntenaConf reads KEY=VALUE pairs from path and returns TURSO_URL / TURSO_TOKEN.
func ParseAntenaConf(path string) (AntenaConf, error) {
	f, err := os.Open(path)
	if err != nil {
		return AntenaConf{}, fmt.Errorf("abrir %s: %w", path, err)
	}
	defer f.Close()

	var conf AntenaConf
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "TURSO_URL":
			conf.TursoURL = val
		case "TURSO_TOKEN":
			conf.TursoToken = val
		}
	}
	if err := sc.Err(); err != nil {
		return AntenaConf{}, fmt.Errorf("ler %s: %w", path, err)
	}
	if conf.TursoURL == "" {
		return AntenaConf{}, fmt.Errorf("%s: TURSO_URL ausente", path)
	}
	if conf.TursoToken == "" {
		return AntenaConf{}, fmt.Errorf("%s: TURSO_TOKEN ausente", path)
	}
	return conf, nil
}

// MaterializeFlyToml replaces <dbname> in the example template and returns the content.
func MaterializeFlyToml(example []byte, dbName string) []byte {
	return []byte(strings.ReplaceAll(string(example), flyPlaceholder, dbName))
}

func flyAppName(dbName string) string {
	return "antena-" + dbName
}

func generateSessionSecret() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func findFlyCLI() (string, error) {
	if p, err := exec.LookPath("fly"); err == nil {
		return p, nil
	}
	if p, err := exec.LookPath("flyctl"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("fly/flyctl não encontrado no PATH")
}

func runFly(ctx context.Context, flyBin, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, flyBin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		if text != "" {
			return text, fmt.Errorf("%w: %s", err, text)
		}
		return text, err
	}
	return text, nil
}

func appAlreadyExists(output string) bool {
	lower := strings.ToLower(output)
	return strings.Contains(lower, "already exists") ||
		strings.Contains(lower, "name has already been taken") ||
		strings.Contains(lower, "app already exists")
}

// DeployFly creates the Fly app for the client Antena instance and deploys the image.
func DeployFly(ctx context.Context, dbName string) error {
	if dbName == "" {
		return fmt.Errorf("nome do banco vazio")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("obter diretório atual: %w", err)
	}

	confPath := filepath.Join(cwd, "instals", dbName+"-antena.conf")
	conf, err := ParseAntenaConf(confPath)
	if err != nil {
		return err
	}

	examplePath := filepath.Join(cwd, flyTomlExample)
	example, err := os.ReadFile(examplePath)
	if err != nil {
		return fmt.Errorf("ler %s: %w", flyTomlExample, err)
	}
	tomlContent := MaterializeFlyToml(example, dbName)

	tmpDir, err := os.MkdirTemp("", "signal-admin-fly-*")
	if err != nil {
		return fmt.Errorf("criar diretório temporário: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tomlPath := filepath.Join(tmpDir, "fly.toml")
	if err := os.WriteFile(tomlPath, tomlContent, 0o644); err != nil {
		return fmt.Errorf("escrever fly.toml: %w", err)
	}

	flyBin, err := findFlyCLI()
	if err != nil {
		return err
	}

	app := flyAppName(dbName)
	if out, err := runFly(ctx, flyBin, tmpDir, "apps", "create", app, "--org", flyOrg); err != nil {
		if !appAlreadyExists(out + err.Error()) {
			return fmt.Errorf("criar app fly: %w", err)
		}
	}

	sessionSecret, err := generateSessionSecret()
	if err != nil {
		return fmt.Errorf("gerar SESSION_SECRET: %w", err)
	}

	if _, err := runFly(ctx, flyBin, tmpDir,
		"secrets", "set",
		"TURSO_TOKEN="+conf.TursoToken,
		"TURSO_URL="+conf.TursoURL,
		"ADDR=:4000",
		"SESSION_SECRET="+sessionSecret,
		"-a", app,
	); err != nil {
		return fmt.Errorf("definir secrets fly: %w", err)
	}

	if _, err := runFly(ctx, flyBin, tmpDir,
		"deploy",
		"--image", flyImage,
		"-a", app,
		"--ha=false",
	); err != nil {
		return fmt.Errorf("deploy fly: %w", err)
	}

	return nil
}

// FlyInstallFlashMessage returns a user-facing flash for a successful Fly install.
func FlyInstallFlashMessage(clienteID string) string {
	return "Antena instalada no fly.io (app " + flyAppName(clienteID) + ")."
}

// FlyInstallErrorFlashMessage returns a user-facing flash for a failed Fly install.
func FlyInstallErrorFlashMessage(errDetail string) string {
	if strings.TrimSpace(errDetail) == "" {
		return "Falha ao instalar no fly.io."
	}
	return "Falha ao instalar no fly.io. Detalhe: " + errDetail
}

// FlyAlreadyInstalledFlashMessage is shown when onfly is already true.
func FlyAlreadyInstalledFlashMessage() string {
	return "Cliente já está instalado no fly.io."
}
