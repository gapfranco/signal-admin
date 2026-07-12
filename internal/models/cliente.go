package models

type Cliente struct {
	ClienteID  string
	Nome       string
	CNPJ       string
	Email      string
	Telefone   string
	ValidUntil string
	Status     string
	Observacao string
	OnFly      bool
	CreatedAt  string
	UpdatedAt  string
}
