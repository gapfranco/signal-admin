package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"signal-admin/config"

	webview "github.com/webview/webview_go"
)

func runDesktop(app *application, cfg config.Config, logger *slog.Logger) {
	setDesktopLocale()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		logger.Error("failed to bind port", "error", err)
		os.Exit(1)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	srv := &http.Server{Handler: app.routes()}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()
	defer srv.Close()

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	logger.Info("desktop server listening", "url", url)

	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle(cfg.NomeEmpresa + " — Signal Admin")
	w.SetSize(1280, 800, webview.HintNone)
	maximizeDesktopWindow(w)
	w.Navigate(url)
	w.Run()
}

func setDesktopLocale() {
	const locale = "pt_BR.UTF-8"
	for _, key := range []string{"LANG", "LC_ALL", "LC_TIME", "LC_NUMERIC", "LC_MONETARY"} {
		_ = os.Setenv(key, locale)
	}
	_ = os.Setenv("LANGUAGE", "pt_BR:pt")
}
