package storage

import (
	"context"
	"database/sql"
	"fmt"
	"signal-admin/internal/models"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"golang.org/x/crypto/bcrypt"
	turso "turso.tech/database/tursogo"
)

type DBConfig struct {
	URL       string
	Token     string
	Mode      string
	LocalPath string
}

type TursoDB struct {
	db     *sql.DB
	pullFn func(ctx context.Context) error
	pushFn func(ctx context.Context) error
}

func enableForeignKeys(db *sql.DB) error {
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	return err
}

func NewTursoDB(cfg DBConfig) (*TursoDB, error) {
	switch cfg.Mode {
	case "local":
		db, err := sql.Open("turso", cfg.LocalPath)
		if err != nil {
			return nil, err
		}
		if err := db.Ping(); err != nil {
			return nil, err
		}
		if err := enableForeignKeys(db); err != nil {
			return nil, err
		}
		return &TursoDB{db: db}, nil

	case "sync":
		ctx := context.Background()
		syncDb, err := turso.NewTursoSyncDb(ctx, turso.TursoSyncDbConfig{
			Path:      cfg.LocalPath,
			RemoteUrl: cfg.URL,
			AuthToken: cfg.Token,
		})
		if err != nil {
			return nil, fmt.Errorf("turso sync: %w", err)
		}
		db, err := syncDb.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("turso sync connect: %w", err)
		}
		if err := enableForeignKeys(db); err != nil {
			return nil, err
		}
		pullFn := func(ctx context.Context) error {
			if _, err := syncDb.Pull(ctx); err != nil {
				return fmt.Errorf("turso pull: %w", err)
			}
			return nil
		}
		pushFn := func(ctx context.Context) error {
			if err := syncDb.Push(ctx); err != nil {
				return fmt.Errorf("turso push: %w", err)
			}
			return nil
		}
		return &TursoDB{db: db, pullFn: pullFn, pushFn: pushFn}, nil

	default:
		var dbURL string
		if cfg.Token != "" {
			dbURL = fmt.Sprintf("%s?authToken=%s", cfg.URL, cfg.Token)
		} else {
			dbURL = cfg.URL
		}
		db, err := sql.Open("libsql", dbURL)
		if err != nil {
			return nil, err
		}
		if err := db.Ping(); err != nil {
			return nil, err
		}
		if err := enableForeignKeys(db); err != nil {
			return nil, err
		}
		return &TursoDB{db: db}, nil
	}
}

// Pull traz alterações remotas para a réplica local. No-op fora do modo sync.
func (t *TursoDB) Pull(ctx context.Context) error {
	if t.pullFn != nil {
		return t.pullFn(ctx)
	}
	return nil
}

// Push envia alterações locais para o Turso Cloud. No-op fora do modo sync.
func (t *TursoDB) Push(ctx context.Context) error {
	if t.pushFn != nil {
		return t.pushFn(ctx)
	}
	return nil
}

// Sync replica escritas locais e puxa o delta remoto (Push → Pull).
func (t *TursoDB) Sync(ctx context.Context) error {
	if t.pushFn == nil {
		return nil
	}
	if err := t.pushFn(ctx); err != nil {
		return err
	}
	return t.pullFn(ctx)
}

// SyncStartup alinha a réplica local com a nuvem antes de aceitar tráfego (Pull → Push).
func (t *TursoDB) SyncStartup(ctx context.Context) error {
	if t.pullFn == nil {
		return nil
	}
	if err := t.pullFn(ctx); err != nil {
		return err
	}
	return t.pushFn(ctx)
}

func (t *TursoDB) DB() *sql.DB {
	return t.db
}

func (t *TursoDB) Close() error {
	return t.db.Close()
}

func (t *TursoDB) HasUsers() (bool, error) {
	var count int
	err := t.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count > 0, err
}

func (t *TursoDB) CreateUser(usuario, nome, senha string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = t.db.Exec("INSERT INTO users (usuario, nome, senha) VALUES (?, ?, ?)", usuario, nome, string(hash))
	return err
}

func (t *TursoDB) Authenticate(usuario, senha string) (*models.User, error) {
	var user models.User
	var hash string
	err := t.db.QueryRow("SELECT id, usuario, nome, senha FROM users WHERE usuario = ?", usuario).
		Scan(&user.ID, &user.Usuario, &user.Nome, &hash)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha)); err != nil {
		return nil, fmt.Errorf("senha inválida")
	}
	return &user, nil
}

func (t *TursoDB) ListUsersFilter(usuario, nome string, limit, offset int) ([]models.User, int, error) {
	like := func(s string) string { return "%" + s + "%" }
	var total int
	if err := t.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE usuario LIKE ? AND nome LIKE ?",
		like(usuario), like(nome),
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := t.db.Query(
		"SELECT id, usuario, nome FROM users WHERE usuario LIKE ? AND nome LIKE ? ORDER BY usuario LIMIT ? OFFSET ?",
		like(usuario), like(nome), limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Usuario, &u.Nome); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (t *TursoDB) GetUser(id int) (*models.User, error) {
	var u models.User
	err := t.db.QueryRow("SELECT id, usuario, nome FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Usuario, &u.Nome)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (t *TursoDB) UpdateUserNome(id int, nome string) error {
	_, err := t.db.Exec("UPDATE users SET nome = ? WHERE id = ?", nome, id)
	return err
}

func (t *TursoDB) DeleteUser(id int) error {
	_, err := t.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func scanCliente(row interface {
	Scan(dest ...any) error
}) (models.Cliente, error) {
	var c models.Cliente
	var validUntil sql.NullString
	err := row.Scan(
		&c.ClienteID, &c.Nome, &c.CNPJ, &c.Email, &c.Telefone,
		&validUntil, &c.Status, &c.Observacao, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return models.Cliente{}, err
	}
	if validUntil.Valid {
		c.ValidUntil = validUntil.String
	}
	return c, nil
}

func (t *TursoDB) CreateCliente(c models.Cliente) error {
	now := time.Now().UTC().Format(time.RFC3339)
	c.CreatedAt = now
	c.UpdatedAt = now
	var validUntil any
	if c.ValidUntil != "" {
		validUntil = c.ValidUntil
	}
	_, err := t.db.Exec(`
		INSERT INTO clientes (
			cliente_id, nome, cnpj, email, telefone, valid_until,
			status, observacao, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ClienteID, c.Nome, c.CNPJ, c.Email, c.Telefone, validUntil,
		c.Status, c.Observacao, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (t *TursoDB) GetCliente(clienteID string) (*models.Cliente, error) {
	row := t.db.QueryRow(`
		SELECT cliente_id, nome, cnpj, email, telefone, valid_until,
		       status, observacao, created_at, updated_at
		FROM clientes WHERE cliente_id = ?`, clienteID)
	c, err := scanCliente(row)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (t *TursoDB) UpdateCliente(c models.Cliente) error {
	c.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	var validUntil any
	if c.ValidUntil != "" {
		validUntil = c.ValidUntil
	}
	_, err := t.db.Exec(`
		UPDATE clientes SET
			nome = ?, cnpj = ?, email = ?, telefone = ?, valid_until = ?,
			status = ?, observacao = ?, updated_at = ?
		WHERE cliente_id = ?`,
		c.Nome, c.CNPJ, c.Email, c.Telefone, validUntil,
		c.Status, c.Observacao, c.UpdatedAt, c.ClienteID,
	)
	return err
}

func (t *TursoDB) DeleteCliente(clienteID string) error {
	_, err := t.db.Exec("DELETE FROM clientes WHERE cliente_id = ?", clienteID)
	return err
}

func (t *TursoDB) ListClientesFilter(clienteID, nome, status string, limit, offset int) ([]models.Cliente, int, error) {
	like := func(s string) string { return "%" + s + "%" }
	where := "WHERE cliente_id LIKE ? AND nome LIKE ?"
	args := []any{like(clienteID), like(nome)}
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM clientes " + where
	if err := t.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `
		SELECT cliente_id, nome, cnpj, email, telefone, valid_until,
		       status, observacao, created_at, updated_at
		FROM clientes ` + where + ` ORDER BY nome LIMIT ? OFFSET ?`
	listArgs := append(append([]any{}, args...), limit, offset)
	rows, err := t.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clientes []models.Cliente
	for rows.Next() {
		c, err := scanCliente(rows)
		if err != nil {
			return nil, 0, err
		}
		clientes = append(clientes, c)
	}
	return clientes, total, rows.Err()
}
