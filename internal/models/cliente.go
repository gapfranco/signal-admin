package models

type Cliente struct {
	ClienteID      string
	Nome           string
	CNPJ           string
	Email          string
	Telefone       string
	SlugTurso      string
	ValidUntil     string
	MaxInstalacoes int
	Status         string
	Observacao     string
	CreatedAt      string
	UpdatedAt      string
}
