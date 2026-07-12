
assumir que esta autenticado com fly.io

onde:
<dbname> - slug do banco turso
<token> - token de conexão ao banco turso. Obter de <dbnane>-signal.conf no folder instals, da linha TURSO_TOKEN=....

fly apps create antena-<dbname> --org personal
fly secrets set TURSO_TOKEN=<token> -a antena-<dbname>
fly secrets set TURSO_URL=libsql://<dbname>-gapfranco.aws-us-east-1.turso.io
fly secrets set ADDR=:4000
fly secrets set SESSION_SECRET=$(openssl rand -hex 32)$(openssl rand -hex 32)
fly deploy --image ghcr.io/gapfranco/antena:latest -a antena-<dbname> --ha=false

criar um fly.toml baseado no fly.toml.example na raiz do projeto, substituindo apenas o <dbname> do nome do app.

