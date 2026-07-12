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
- Clientes (signal-ready: valid_until, status)
- Provisionamento Turso por cliente (`internal/provision` + `internal/clientstorage`)

## Provisionamento Turso

Ao criar um cliente, o admin provisiona automaticamente um banco Turso remoto (nome = `cliente_id`), aplica o schema Signal remoto, registra a licença e gera `{cliente_id}-signal.conf` e `{cliente_id}-antena.conf` na pasta `instals/` na raiz do projeto.

Configuração necessária via **variáveis de ambiente** (não no `signal-admin.conf`):

```bash
export TURSO_ORG=          # slug da organização Turso
export TURSO_TOKEN=        # token Platform API da organização (distinto de DB_TOKEN)
```

Comportamento:

- Banco já existente: cliente é criado, flash de aviso
- Exclusão de cliente: registro removido; banco Turso na nuvem permanece (flash de aviso se existir)
- Sem `TURSO_ORG`/`TURSO_TOKEN` no ambiente: cliente criado sem provisionar

Código duplicado do projeto signal (coexistência temporária até remoção futura no signal).

Fora de escopo: tabela installations, deprovision/delete Turso, deploy Antena.
