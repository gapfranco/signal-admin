# Variáveis de Configuração
BINARY_NAME=signal-admin
SRC_PATH=./cmd/signal-admin
OUT_DIR=./build
TAILWIND_BIN=tailwindcss
TW_INPUT=./ui/input.css
TW_OUTPUT=./ui/static/css/styles.css

.PHONY: all clean build tailwind-build tailwind-watch manual-pdf

all: build

tailwind-build:
	@echo "Gerando CSS com Tailwind..."
	$(TAILWIND_BIN) -i $(TW_INPUT) -o $(TW_OUTPUT) --minify

tailwind-watch:
	$(TAILWIND_BIN) -i $(TW_INPUT) -o $(TW_OUTPUT) --watch

build: tailwind-build
	@echo "Construindo signal-admin (desktop)..."
	@mkdir -p $(OUT_DIR)
	go build -o $(OUT_DIR)/$(BINARY_NAME) $(SRC_PATH)

windows: tailwind-build
	@mkdir -p $(OUT_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui" -o $(OUT_DIR)/$(BINARY_NAME).exe $(SRC_PATH)

linux: tailwind-build
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(OUT_DIR)/$(BINARY_NAME)_linux $(SRC_PATH)

darwin: tailwind-build
	@mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(OUT_DIR)/$(BINARY_NAME)_mac $(SRC_PATH)

manual-pdf:
	@./recursos/build-pdf.sh

clean:
	rm -rf $(OUT_DIR)
	rm -f $(TW_OUTPUT)
