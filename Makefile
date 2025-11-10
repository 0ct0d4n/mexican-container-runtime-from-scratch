export SSHPASS='axolotl'

# Makefile: build + upload + install usando sshpass
# USO SEGURO RECOMENDADO:
#  - export SSHPASS='TuPassword'   # o anteponer SSHPASS='TuPassword' make deploy
#  - NO poner la contraseña en el Makefile ni en el repo.

BINARY_NAME = axod
MAIN_FILE   = daemon/main.go

REMOTE_USER = axolotl
REMOTE_HOST = 192.168.64.2
REMOTE_PORT = 22
REMOTE_PATH = /home/axolotl
REMOTE_BIN  = /usr/local/bin

GOOS  = linux
FLAGS = -ldflags="-s -w"

# Detectar arquitectura remota (intento simple)
ARCH_CMD = $(shell ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "uname -m" 2>/dev/null || echo unknown)

ifeq ($(ARCH_CMD),x86_64)
	GOARCH = amd64
else ifeq ($(ARCH_CMD),aarch64)
	GOARCH = arm64
else
	GOARCH = amd64
endif

# ----------------------
# Targets
# ----------------------

.PHONY: all build upload install deploy clean

# build: compila el binario para la arquitectura detectada
build:
	@echo "🧠 Detectando arquitectura remota: $(ARCH_CMD)"
	@echo "🏗️ Compilando $(BINARY_NAME) para $(GOOS)/$(GOARCH)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(FLAGS) -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Binario compilado: $(BINARY_NAME)"

# upload: usa sshpass para scp el binario (requiere SSHPASS en entorno)
upload: build
ifndef SSHPASS
	$(error SSHPASS no definido. Ejecuta: export SSHPASS='TuPassword'  o SSHPASS='TuPassword' make upload)
endif
	@echo "📦 Transfiriendo $(BINARY_NAME) a $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH) (puerto $(REMOTE_PORT))"
	sshpass -p "$$SSHPASS" scp -P $(REMOTE_PORT) $(BINARY_NAME) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/
	@echo "✅ Transferencia completada."

# install: mueve el binario a /usr/local/bin (requiere sudo en remoto)
install: upload
ifndef SSHPASS
	$(error SSHPASS no definido. Ejecuta: export SSHPASS='TuPassword'  o SSHPASS='TuPassword' make install)
endif
	@echo "⚙️ Instalando $(BINARY_NAME) en $(REMOTE_BIN)..."
	sshpass -p "$$SSHPASS" ssh -o StrictHostKeyChecking=no -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) \
		"sudo mv $(REMOTE_PATH)/$(BINARY_NAME) $(REMOTE_BIN)/ && sudo chmod +x $(REMOTE_BIN)/$(BINARY_NAME)"
	@echo "✅ Instalación completada."

# deploy: pipeline completo
deploy: install
	@echo "🚀 ¡$(BINARY_NAME) desplegado correctamente en $(REMOTE_HOST)!"

# clean: elimina binario local
clean:
	@echo "🧹 Limpiando binarios locales..."
	-rm -f $(BINARY_NAME)
	@echo "✅ Limpieza completada."