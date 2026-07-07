# Signal Admin — Diretrizes

Back-office interno para gestão de clientes e instalações Signal. Segue os mesmos padrões do hs-financ.

## Estrutura

- `cmd/signal-admin/` — aplicação desktop (webview + HTTP local)
- `config/` — Viper (`signal-admin.conf`)
- `internal/migrations/` — Goose SQL embed
- `internal/models/` — structs de domínio
- `internal/storage/` — TursoDB (local / remote / sync)
- `ui/` — templates HTMX + Tailwind embed

## Comandos

```bash
make build
make tailwind-build
go test ./...
```

## Configuração local

Use `sample.conf` como base. Para dev offline: `DB_MODE=local`.

## Persistência

Turso sync (`DB_MODE=sync`): nuvem como fonte de verdade, `local.db` como réplica. Após cada write, `syncDB()` executa Push + Pull.

## Escopo atual

- Usuários (auth idêntica ao hs-financ)
- Clientes (signal-ready: slug_turso, valid_until, max_instalacoes, status)

Fora de escopo: provisionamento Turso, tabela installations, integração signal-provision.
