package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"signal-admin/internal/models"

	"github.com/justinas/nosurf"
)

func isFKError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "foreign key constraint")
}

func (app *application) syncDB(ctx context.Context) {
	if err := app.db.Sync(ctx); err != nil {
		app.logger.Error("db sync failed", "error", err)
	}
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)
	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (app *application) newTemplateData(r *http.Request) templateData {
	menu, submenu := menuFromPath(r.URL.Path)
	data := templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
		ActiveMenu:      menu,
		ActiveSubmenu:   submenu,
		NomeEmpresa:     app.config.NomeEmpresa,
	}
	if data.IsAuthenticated {
		data.CurrentUser = app.sessionManager.GetString(r.Context(), "authenticatedUser")
	}
	return data
}

func menuFromPath(path string) (menu, submenu string) {
	switch {
	case path == "/" || path == "":
		return "", ""
	case strings.HasPrefix(path, "/config/clientes"):
		return "cadastros", "clientes"
	case strings.HasPrefix(path, "/config/usuarios"):
		return "config", "usuarios"
	default:
		return "", ""
	}
}

func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	return app.formDecoder.Decode(dst, r.PostForm)
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

func (app *application) calculatePagination(totalItems, currentPage, pageSize int) models.PaginationMetadata {
	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	return models.PaginationMetadata{
		CurrentPage: currentPage,
		PageSize:    pageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasPrev:     currentPage > 1,
		HasNext:     currentPage < totalPages,
		PrevPage:    currentPage - 1,
		NextPage:    currentPage + 1,
	}
}
