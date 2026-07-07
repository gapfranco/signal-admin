package main

import "testing"

func TestValidarCNPJ(t *testing.T) {
	tests := []struct {
		cnpj  string
		valid bool
	}{
		{"", true},
		{"11222333000181", true},
		{"11.222.333/0001-81", true},
		{"00000000000000", false},
		{"123", false},
		{"abcdefghijklmn", false},
	}

	for _, tt := range tests {
		got := validarCNPJ(tt.cnpj)
		if got != tt.valid {
			t.Errorf("validarCNPJ(%q) = %v, want %v", tt.cnpj, got, tt.valid)
		}
	}
}

func TestValidSlug(t *testing.T) {
	tests := []struct {
		slug  string
		valid bool
	}{
		{"", true},
		{"hospital123", true},
		{"Hospital", false},
		{"hospital-sj", false},
		{"abc", true},
	}

	for _, tt := range tests {
		got := validSlug(tt.slug)
		if got != tt.valid {
			t.Errorf("validSlug(%q) = %v, want %v", tt.slug, got, tt.valid)
		}
	}
}

func TestMenuFromPath(t *testing.T) {
	tests := []struct {
		path        string
		wantMenu    string
		wantSubmenu string
	}{
		{"/", "", ""},
		{"/config/clientes", "cadastros", "clientes"},
		{"/config/clientes/new", "cadastros", "clientes"},
		{"/config/usuarios", "config", "usuarios"},
	}

	for _, tt := range tests {
		menu, submenu := menuFromPath(tt.path)
		if menu != tt.wantMenu || submenu != tt.wantSubmenu {
			t.Errorf("menuFromPath(%q) = (%q, %q), want (%q, %q)", tt.path, menu, submenu, tt.wantMenu, tt.wantSubmenu)
		}
	}
}
