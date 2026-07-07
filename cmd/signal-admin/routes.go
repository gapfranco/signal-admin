package main

import (
	"net/http"

	"signal-admin/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.FS(ui.Files))
	mux.Handle("GET /static/", fileServer)

	standard := []func(http.Handler) http.Handler{
		app.recoverPanic,
		app.logRequest,
		commonHeaders,
	}

	dynamic := []func(http.Handler) http.Handler{
		app.sessionManager.LoadAndSave,
		preventCSRF,
		app.requireSetup,
		app.authenticate,
	}

	mux.Handle("GET /setup", app.chain(http.HandlerFunc(app.setup), dynamic...))
	mux.Handle("POST /setup", app.chain(http.HandlerFunc(app.setupPost), dynamic...))
	mux.Handle("GET /login", app.chain(http.HandlerFunc(app.login), dynamic...))
	mux.Handle("POST /login", app.chain(http.HandlerFunc(app.loginPost), dynamic...))

	protected := http.NewServeMux()
	protected.HandleFunc("GET /", app.home)
	protected.HandleFunc("GET /config/clientes", app.clientesList)
	protected.HandleFunc("GET /config/clientes/new", app.clienteNew)
	protected.HandleFunc("POST /config/clientes/new", app.clienteNewPost)
	protected.HandleFunc("GET /config/clientes/{cliente_id}/edit", app.clienteEdit)
	protected.HandleFunc("POST /config/clientes/{cliente_id}/edit", app.clienteEditPost)
	protected.HandleFunc("POST /config/clientes/{cliente_id}/delete", app.clienteDelete)
	protected.HandleFunc("GET /config/usuarios", app.usuariosList)
	protected.HandleFunc("GET /config/usuarios/new", app.usuarioNew)
	protected.HandleFunc("POST /config/usuarios/new", app.usuarioNewPost)
	protected.HandleFunc("GET /config/usuarios/{id}/edit", app.usuarioEdit)
	protected.HandleFunc("POST /config/usuarios/{id}/edit", app.usuarioEditPost)
	protected.HandleFunc("POST /config/usuarios/{id}/delete", app.usuarioDelete)
	protected.HandleFunc("POST /logout", app.logoutPost)

	mux.Handle("/", app.chain(protected, append(dynamic, app.requireAuthentication)...))

	return app.chain(mux, standard...)
}

func (app *application) chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
