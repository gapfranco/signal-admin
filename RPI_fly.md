
Implementar uma feature para criar um serviço em fly.io para acessar o app de um cliente via internet.

Essa feature será chamda opcionalmente pelo operador mediante um botão na tela de alteração de cliente "Instalar em fly.io".
Esse botão só deve estar ativo caso o campo "onfly" no registro de cliente estiver como falso ou nulo.
Apos instalar sem erro, setar o campo "onfly" na tabela de clientes como true.

Assumindo que esta autenticado com fly.io, criar um fly.toml baseado no fly.toml.example na raiz do projeto, substituindo apenas o <dbname> do nome do app.

onde:
<dbname> - slug do banco turso
<token> - token de conexão ao banco turso. Obter de <dbnane>-signal.conf no folder instals, da linha TURSO_TOKEN=....

fly apps create antena-<dbname> --org personal
fly secrets set TURSO_TOKEN=<token> -a antena-<dbname>
fly secrets set TURSO_URL=libsql://<dbname>-gapfranco.aws-us-east-1.turso.io
fly secrets set ADDR=:4000
fly secrets set SESSION_SECRET=$(openssl rand -hex 32)$(openssl rand -hex 32)
fly deploy --image ghcr.io/gapfranco/antena:latest -a antena-<dbname> --ha=false

