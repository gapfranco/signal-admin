package provision

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var dbSlugPattern = regexp.MustCompile(`^[a-z0-9]+$`)

// Options holds validated provision inputs.
type Options struct {
	Client string
	DB     string
	Limit  string
	Org    string
	Token  string
}

// ValidateOptions checks provision inputs and Turso Platform credentials.
func ValidateOptions(client, db, limit, org, token string) (Options, error) {
	client = strings.TrimSpace(client)
	db = strings.TrimSpace(db)
	limit = strings.TrimSpace(limit)
	org = strings.TrimSpace(org)
	token = strings.TrimSpace(token)

	if client == "" {
		return Options{}, errors.New("nome do cliente é obrigatório")
	}
	if db == "" {
		return Options{}, errors.New("código do cliente é obrigatório")
	}
	if !dbSlugPattern.MatchString(db) {
		return Options{}, errors.New("código deve ser um slug com letras minúsculas e algarismos, sem espaços")
	}
	if limit != "" {
		if _, err := time.Parse("2006-01-02", limit); err != nil {
			return Options{}, fmt.Errorf("validade deve estar no formato AAAA-MM-DD: %w", err)
		}
	}
	if org == "" {
		return Options{}, errors.New("variável de ambiente TURSO_ORG é obrigatória")
	}
	if token == "" {
		return Options{}, errors.New("variável de ambiente TURSO_TOKEN é obrigatória")
	}

	return Options{
		Client: client,
		DB:     db,
		Limit:  limit,
		Org:    org,
		Token:  token,
	}, nil
}
