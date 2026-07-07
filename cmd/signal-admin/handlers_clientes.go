package main

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"signal-admin/internal/models"
)

const clientesPageSize = 10

var slugRegex = regexp.MustCompile(`^[a-z0-9]+$`)

type clienteForm struct {
	ClienteID  string `form:"cliente_id"`
	Nome       string `form:"nome"`
	CNPJ       string `form:"cnpj"`
	Email      string `form:"email"`
	Telefone   string `form:"telefone"`
	ValidUntil string `form:"valid_until"`
	Status     string `form:"status"`
	Observacao string `form:"observacao"`
}

func validSlug(s string) bool {
	return s == "" || slugRegex.MatchString(s)
}

func (app *application) clientesList(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	filterID := r.URL.Query().Get("cliente_id")
	filterNome := r.URL.Query().Get("nome")
	filterStatus := r.URL.Query().Get("status")

	clientes, total, err := app.db.ListClientesFilter(filterID, filterNome, filterStatus, clientesPageSize, (page-1)*clientesPageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "clientes"
	data.Data = struct {
		Clientes     []models.Cliente
		Pagination   models.PaginationMetadata
		FilterID     string
		FilterNome   string
		FilterStatus string
	}{
		Clientes:     clientes,
		Pagination:   app.calculatePagination(total, page, clientesPageSize),
		FilterID:     filterID,
		FilterNome:   filterNome,
		FilterStatus: filterStatus,
	}
	app.render(w, r, http.StatusOK, "clientes.html", data)
}

func (app *application) clienteNew(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "clientes"
	data.Data = struct {
		EditMode bool
		Cliente  *models.Cliente
	}{EditMode: false}
	app.render(w, r, http.StatusOK, "cliente_form.html", data)
}

func (app *application) validateClienteForm(form clienteForm, editMode bool) (string, *models.Cliente) {
	form.ClienteID = strings.TrimSpace(form.ClienteID)
	form.Nome = strings.TrimSpace(form.Nome)
	form.Status = strings.TrimSpace(form.Status)

	cliente := &models.Cliente{
		ClienteID:  form.ClienteID,
		Nome:       form.Nome,
		CNPJ:       form.CNPJ,
		Email:      strings.TrimSpace(form.Email),
		Telefone:   strings.TrimSpace(form.Telefone),
		ValidUntil: strings.TrimSpace(form.ValidUntil),
		Status:     form.Status,
		Observacao: strings.TrimSpace(form.Observacao),
	}

	if !editMode {
		if form.ClienteID == "" || form.Nome == "" {
			return "Código e nome são obrigatórios", cliente
		}
		if !validSlug(form.ClienteID) {
			return "Código deve conter apenas letras minúsculas e números", cliente
		}
	} else if form.Nome == "" {
		return "Nome é obrigatório", cliente
	}

	if !validarCNPJ(form.CNPJ) {
		return "CNPJ inválido", cliente
	}
	switch form.Status {
	case "active", "suspended", "inactive":
	default:
		return "Status inválido", cliente
	}

	return "", cliente
}

func (app *application) clienteNewPost(w http.ResponseWriter, r *http.Request) {
	var form clienteForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if form.Status == "" {
		form.Status = "active"
	}

	flashMsg, cliente := app.validateClienteForm(form, false)
	if flashMsg != "" {
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "clientes"
		data.Flash = flashMsg
		data.Data = struct {
			EditMode bool
			Cliente  *models.Cliente
		}{EditMode: false, Cliente: cliente}
		app.render(w, r, http.StatusUnprocessableEntity, "cliente_form.html", data)
		return
	}

	if err := app.db.CreateCliente(*cliente); err != nil {
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "clientes"
		data.Flash = "Erro ao criar cliente (código já existe?)"
		data.Data = struct {
			EditMode bool
			Cliente  *models.Cliente
		}{EditMode: false, Cliente: cliente}
		app.render(w, r, http.StatusUnprocessableEntity, "cliente_form.html", data)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Cliente criado com sucesso.")
	http.Redirect(w, r, "/config/clientes", http.StatusSeeOther)
}

func (app *application) clienteEdit(w http.ResponseWriter, r *http.Request) {
	clienteID := r.PathValue("cliente_id")
	cliente, err := app.db.GetCliente(clienteID)
	if err != nil {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "clientes"
	data.Data = struct {
		EditMode bool
		Cliente  *models.Cliente
	}{EditMode: true, Cliente: cliente}
	app.render(w, r, http.StatusOK, "cliente_form.html", data)
}

func (app *application) clienteEditPost(w http.ResponseWriter, r *http.Request) {
	clienteID := r.PathValue("cliente_id")
	existing, err := app.db.GetCliente(clienteID)
	if err != nil {
		app.notFound(w)
		return
	}

	var form clienteForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.ClienteID = clienteID
	if form.Status == "" {
		form.Status = existing.Status
	}

	flashMsg, cliente := app.validateClienteForm(form, true)
	if flashMsg != "" {
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "clientes"
		data.Flash = flashMsg
		data.Data = struct {
			EditMode bool
			Cliente  *models.Cliente
		}{EditMode: true, Cliente: cliente}
		app.render(w, r, http.StatusUnprocessableEntity, "cliente_form.html", data)
		return
	}

	if err := app.db.UpdateCliente(*cliente); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Cliente atualizado com sucesso.")
	http.Redirect(w, r, "/config/clientes", http.StatusSeeOther)
}

func (app *application) clienteDelete(w http.ResponseWriter, r *http.Request) {
	clienteID := r.PathValue("cliente_id")
	if _, err := app.db.GetCliente(clienteID); err != nil {
		app.notFound(w)
		return
	}
	if err := app.db.DeleteCliente(clienteID); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Cliente excluído.")
	http.Redirect(w, r, "/config/clientes", http.StatusSeeOther)
}
