# Makefile: build + upload + install usando sshpass (compatible con macOS/zsh)

BINARY_NAME = axod
MAIN_FILE   = daemon/main.go

REMOTE_USER = axolotl
REMOTE_HOST = 192.168.64.2
REMOTE_PORT = 22
REMOTE_PATH = /home/axolotl
REMOTE_BIN  = /usr/local/bin

GOOS  = linux
GOARCH = arm64
FLAGS = -ldflags="-s -w"

# ----------------------
# Targets
# ----------------------

.PHONY: all build upload install deploy clean run

# build: compila el binario
build:
	@echo "🏗️ Compilando $(BINARY_NAME) para $(GOOS)/$(GOARCH)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(FLAGS) -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Binario compilado: $(BINARY_NAME)"

# upload: usa sshpass para scp el binario (requiere SSHPASS en entorno)
upload: build
ifndef SSHPASS
	$(error ❌ SSHPASS no definido. Ejecuta: SSHPASS='TuPassword' make upload)
endif
	@echo "📦 Transfiriendo $(BINARY_NAME) a $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)..."
	@env SSHPASS=$(SSHPASS) sshpass -p $$SSHPASS scp -o StrictHostKeyChecking=no -P $(REMOTE_PORT) $(BINARY_NAME) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/
	@echo "✅ Transferencia completada."

# install: mueve el binario a /usr/local/bin (requiere sudo remoto)
install: upload
	@echo "⚙️ Instalando $(BINARY_NAME) en $(REMOTE_BIN)..."
	@sshpass -p $(SSHPASS) ssh -o StrictHostKeyChecking=no -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) \
		"echo $(SSHPASS) | sudo -S mv $(REMOTE_PATH)/$(BINARY_NAME) $(REMOTE_BIN)/ && echo $(SSHPASS) | sudo -S chmod +x $(REMOTE_BIN)/$(BINARY_NAME)"
	@echo "✅ Instalación completada."
# deploy: build + upload + install
deploy: install run
	@echo "🚀 ¡$(BINARY_NAME) desplegado correctamente en $(REMOTE_HOST)!"

run:
	@echo "🚀 Ejecutando axod en modo interactivo..."
	@sshpass -p $(SSHPASS) ssh -o StrictHostKeyChecking=no -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "axod"
# clean: elimina binario local
clean:
	@echo "🧹 Limpiando binarios locales..."
	-rm -f $(BINARY_NAME)
	@echo "✅ Limpieza completada."