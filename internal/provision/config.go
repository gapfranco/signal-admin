package provision

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigPaths holds the paths of generated configuration files.
type ConfigPaths struct {
	Signal string
	Antena string
}

// WriteConfigFiles creates {db}-signal.conf and {db}-antena.conf in dir.
func WriteConfigFiles(dir, db, databaseURL, authToken string) (ConfigPaths, error) {
	signalPath := filepath.Join(dir, db+"-signal.conf")
	antenaPath := filepath.Join(dir, db+"-antena.conf")

	signalContent := fmt.Sprintf(`DATABASE_URL=./data/signal.db
PORT=9000
TURSO_DATABASE_URL=%s
TURSO_AUTH_TOKEN=%s
`, databaseURL, authToken)

	antenaContent := fmt.Sprintf(`ADDR=:4000
TURSO_URL=%s
TURSO_TOKEN=%s
`, databaseURL, authToken)

	if err := os.WriteFile(signalPath, []byte(signalContent), 0o644); err != nil {
		return ConfigPaths{}, fmt.Errorf("escrever %s: %w", signalPath, err)
	}
	if err := os.WriteFile(antenaPath, []byte(antenaContent), 0o644); err != nil {
		return ConfigPaths{}, fmt.Errorf("escrever %s: %w", antenaPath, err)
	}

	return ConfigPaths{
		Signal: signalPath,
		Antena: antenaPath,
	}, nil
}
