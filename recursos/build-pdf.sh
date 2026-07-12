#!/usr/bin/env bash
# Gera PDFs a partir dos manuais Markdown em recursos/
#
# Uso:
#   ./recursos/build-pdf.sh                    # todos os manuais
#   ./recursos/build-pdf.sh manual_admin       # um manual específico
#
# Caminhos de build (em ordem de preferência):
#   1. pandoc + xelatex + template Eisvogel  (melhor tipografia)
#   2. puppeteer-core + Chromium + CSS local  (npm, sem LaTeX)
#
# Dependências LaTeX (opcional, caminho 1):
#   sudo apt install pandoc texlive-xetex texlive-fonts-recommended \
#                    texlive-lang-portuguese texlive-latex-extra
#
# Template Eisvogel (opcional, caminho 1):
#   Baixe em https://github.com/Wandmalfarbe/pandoc-latex-template/releases/latest
#   e copie eisvogel.latex para recursos/pdf/templates/eisvogel.latex

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PDF_DIR="$ROOT/recursos/pdf"
EISVOGEL="$PDF_DIR/templates/eisvogel.latex"

ALL_MANUALS=(manual_admin)

log() { printf 'build-pdf: %s\n' "$*"; }
die() { printf 'build-pdf: erro: %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<EOF
uso: build-pdf.sh [manual ...]

Sem argumentos, gera todos os manuais:
  ${ALL_MANUALS[*]}

Saída: recursos/<manual>.pdf
EOF
}

ensure_npm_deps() {
  if [[ ! -d "$ROOT/node_modules/puppeteer-core" ]]; then
    log "instalando dependências npm (puppeteer-core, marked, gray-matter)..."
    (cd "$ROOT" && npm install >/dev/null)
  fi
}

latex_available() {
  command -v pandoc >/dev/null 2>&1 \
    && command -v xelatex >/dev/null 2>&1 \
    && [[ -f "$EISVOGEL" ]]
}

build_with_latex() {
  local md_src="$1"
  local pdf_out="$2"

  log "gerando $(basename "$pdf_out") com pandoc + xelatex + Eisvogel..."
  pandoc "$md_src" \
    -o "$pdf_out" \
    --from markdown \
    --template "$EISVOGEL" \
    --pdf-engine=xelatex \
    -V lang=pt-BR \
    -V mainfont="DejaVu Serif" \
    -V sansfont="DejaVu Sans" \
    -V monofont="DejaVu Sans Mono" \
    --toc \
    --toc-depth=3 \
    --number-sections \
    --highlight-style=tango
}

build_with_chromium() {
  local name="$1"
  local md_src="$2"
  local md_tmp="$3"
  local pdf_out="$4"

  node "$PDF_DIR/preprocess.mjs" "$md_src" >"$md_tmp"
  log "gerando $(basename "$pdf_out") com Chromium + CSS..."
  node "$PDF_DIR/render-pdf.mjs" "$md_tmp" "$pdf_out"
  rm -f "$md_tmp"
}

build_manual() {
  local name="$1"
  local md_src="$ROOT/recursos/${name}.md"
  local md_tmp="$PDF_DIR/.${name}.build.md"
  local pdf_out="$ROOT/recursos/${name}.pdf"

  [[ -f "$md_src" ]] || die "arquivo não encontrado: $md_src"

  if latex_available; then
    build_with_latex "$md_src" "$pdf_out"
    log "concluído: $pdf_out (LaTeX/Eisvogel)"
    return
  fi

  if ! command -v node >/dev/null 2>&1; then
    die "Node.js não encontrado. Instale Node ou pandoc+texlive para gerar PDFs."
  fi

  ensure_npm_deps
  build_with_chromium "$name" "$md_src" "$md_tmp" "$pdf_out"
  log "concluído: $pdf_out (Chromium/CSS)"
}

main() {
  if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    usage
    exit 0
  fi

  mkdir -p "$PDF_DIR/templates"

  local targets=()
  if [[ $# -eq 0 ]]; then
    targets=("${ALL_MANUALS[@]}")
  else
    targets=("$@")
  fi

  for name in "${targets[@]}"; do
    local found=false
    for known in "${ALL_MANUALS[@]}"; do
      if [[ "$name" == "$known" ]]; then
        found=true
        break
      fi
    done
    [[ "$found" == true ]] || die "manual desconhecido: $name (válidos: ${ALL_MANUALS[*]})"
  done

  if ! latex_available; then
    ensure_npm_deps
  fi

  for name in "${targets[@]}"; do
    build_manual "$name"
  done

  if ! latex_available; then
    log "dica: para PDF com LaTeX, instale pandoc/texlive e coloque eisvogel.latex em recursos/pdf/templates/"
  fi
}

main "$@"
