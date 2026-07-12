package provision

import (
	"context"
	"errors"
	"fmt"
	"os"

	"signal-admin/internal/clientstorage"
)

var ErrDatabaseExists = errors.New("Banco de dados já existente")

// Result summarizes a successful provision run.
type Result struct {
	DBName      string
	DatabaseURL string
	AuthToken   string
	Configs     ConfigPaths
}

// Run provisions a new Turso database for a client.
func Run(ctx context.Context, opts Options) (Result, error) {
	api, err := NewTursoAPI(opts.Org, opts.Token)
	if err != nil {
		return Result{}, err
	}

	exists, err := api.DatabaseExists(ctx, opts.DB)
	if err != nil {
		return Result{}, fmt.Errorf("verificar banco: %w", err)
	}
	if exists {
		return Result{}, ErrDatabaseExists
	}

	created, err := api.CreateDatabase(ctx, opts.DB)
	if err != nil {
		return Result{}, fmt.Errorf("criar banco: %w", err)
	}

	authToken, err := api.CreateDatabaseToken(ctx, opts.DB)
	if err != nil {
		return Result{}, fmt.Errorf("criar token do banco: %w", err)
	}

	databaseURL := LibSQLURL(created.Hostname)

	remoteDB, err := clientstorage.OpenTursoDB(databaseURL, authToken)
	if err != nil {
		return Result{}, fmt.Errorf("conectar ao banco criado (%s): %w", opts.DB, err)
	}
	defer remoteDB.Close()

	if err := clientstorage.InitRemoteSchema(ctx, remoteDB); err != nil {
		return Result{}, fmt.Errorf("aplicar schema remoto: %w", err)
	}

	if err := clientstorage.InsertLicense(ctx, remoteDB, clientstorage.License{
		Nome:     opts.Client,
		ClientID: opts.DB,
		Validade: opts.Limit,
		Status:   "active",
	}); err != nil {
		return Result{}, fmt.Errorf("registrar licença: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return Result{}, fmt.Errorf("obter diretório atual: %w", err)
	}

	configs, err := WriteConfigFiles(cwd, opts.DB, databaseURL, authToken)
	if err != nil {
		return Result{}, err
	}

	return Result{
		DBName:      opts.DB,
		DatabaseURL: databaseURL,
		AuthToken:   authToken,
		Configs:     configs,
	}, nil
}
