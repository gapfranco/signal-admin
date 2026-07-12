# Geração de PDF do manual

## Comando rápido

Na raiz do repositório:

```bash
make manual-pdf
# ou
./recursos/build-pdf.sh
```

Gera:

| Markdown | PDF |
| --- | --- |
| `recursos/manual_admin.md` | `recursos/manual_admin.pdf` |

### Manual específico

```bash
./recursos/build-pdf.sh manual_admin
```

## Caminho padrão (npm + Chromium)

Usa `puppeteer-core` com o Google Chrome/Chromium instalado no sistema e o CSS em `manual.css`. Não exige LaTeX nem pandoc — apenas Node.js, Chrome e `npm install`.

**Requer internet na geração** para baixar as fontes Inter e JetBrains Mono via Google Fonts.

Se o Chrome estiver em caminho não padrão:

```bash
PUPPETEER_EXECUTABLE_PATH=/caminho/para/chrome make manual-pdf
```

## Caminho avançado (LaTeX + Eisvogel)

Para tipografia editorial (sumário nativo, numeração de seções, melhor quebra de página):

1. Instale as dependências:

```bash
sudo apt install pandoc texlive-xetex texlive-fonts-recommended \
                 texlive-lang-portuguese texlive-latex-extra
```

2. Baixe o template [Eisvogel](https://github.com/Wandmalfarbe/pandoc-latex-template/releases/latest) e copie `eisvogel.latex` para:

```text
recursos/pdf/templates/eisvogel.latex
```

3. Rode `make manual-pdf` — o script detecta pandoc/xelatex/Eisvogel e usa esse caminho automaticamente.

## Metadados do Markdown

O manual usa front matter YAML no topo para capa, sumário e metadados do PDF:

```yaml
---
title: "Manual do Signal Admin"
subtitle: "Back-office de clientes e instalações Signal"
author: "Signal Admin"
date: "2026"
lang: pt-BR
toc: true
numbersections: true
---
```

| Chave | Função |
| --- | --- |
| `title` | Título na capa e no header |
| `subtitle` | Subtítulo na capa |
| `author` | Autor na capa |
| `date` | Data/versão na capa |
| `toc` | `false` desabilita o sumário |
| `header` | Texto curto no header de cada página (opcional) |

## Arquivos

| Arquivo | Função |
| --- | --- |
| `build-pdf.sh` | Orquestra os dois caminhos de build |
| `manual.css` | Design system visual (indigo/slate, Inter) |
| `preprocess.mjs` | Sumário HTML e parágrafo introdutório |
| `render-pdf.mjs` | Capa, fontes e conversão para PDF |
| `templates/eisvogel.latex` | Template LaTeX (opcional, não versionado) |
