# Signal Admin

Back-office para controle de instalações e clientes do ecossistema Signal.

## Requisitos

- Go 1.25+
- CLI `tailwindcss` (para build de CSS)
- Turso/libSQL (opcional em desenvolvimento com `DB_MODE=local`)

## Configuração

Copie `sample.conf` para `signal-admin.conf` na raiz do projeto:

```env
DB_URL=libsql://signal-admin-registry-<org>.turso.io
DB_TOKEN=<token>
DB_MODE=sync
DB_LOCAL_PATH=local.db
NOME_EMPRESA=Signal Admin
```

### Modos de banco (`DB_MODE`)

| Modo | Descrição |
|------|-----------|
| `sync` | Réplica local + Turso Cloud (padrão). Escritas sincronizadas após mutações. |
| `remote` | Conexão direta ao Turso Cloud. |
| `local` | SQLite local puro, sem rede. Ideal para desenvolvimento. |

## Build

```bash
make build          # desktop (webview)
make tailwind-build # apenas CSS
go test ./...
```

## Execução

```bash
./build/signal-admin
```

Abre janela nativa (webview) em porta efêmera local. No primeiro acesso, o sistema redireciona para `/setup` para criar o usuário administrador.

## Funcionalidades

- Autenticação com sessão, CSRF e bcrypt (mesmo padrão do hs-financ)
- CRUD de clientes (campos preparados para integração Signal/Turso)
- CRUD de usuários
- Menu: Cadastros → Clientes; Configuração → Usuários

## Setup Turso (produção)

1. Criar banco Turso dedicado (ex.: `signal-admin-registry`)
2. Gerar token full-access
3. Configurar `signal-admin.conf` com `DB_MODE=sync`
4. Iniciar aplicação — migrations locais + remotas + sync inicial
5. Acessar `/setup` e criar admin
