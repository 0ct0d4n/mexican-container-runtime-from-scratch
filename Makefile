# === CONFIGURACIÓN ===
BINARY_NAME = axod
TARGET_PATH = /usr/local/bin/$(BINARY_NAME)
SERVICE_NAME = axod.service

REMOTE_USER = axolotl
REMOTE_HOST = 192.168.64.2
REMOTE_PORT = 22
SSHPASS ?= axolotl

GOOS = linux
GOARCH = arm64

# === TARGETS ===

.PHONY: all build upload install service start stop restart status deploy

all: build

build:
	@echo "⚙️ Compilando $(BINARY_NAME) para Linux/ARM64..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o $(BINARY_NAME) daemon/main.go
	@echo "✅ Binario compilado: $(BINARY_NAME)"

upload:
	@echo "📤 Transfiriendo $(BINARY_NAME) a $(REMOTE_USER)@$(REMOTE_HOST)..."
	@sshpass -p "$(SSHPASS)" scp -P $(REMOTE_PORT) $(BINARY_NAME) $(REMOTE_USER)@$(REMOTE_HOST):/home/$(REMOTE_USER)/
	@echo "✅ Transferencia completada."

install:
	@echo "⚙️ Instalando $(BINARY_NAME) en $(TARGET_PATH)..."
	@sshpass -p "$(SSHPASS)" ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "\
		echo '$(SSHPASS)' | sudo -S mv /home/$(REMOTE_USER)/$(BINARY_NAME) $(TARGET_PATH) && \
		echo '$(SSHPASS)' | sudo -S chmod +x $(TARGET_PATH)"
	@echo "✅ Instalación completada."

service:
	@echo "🪄 Creando servicio systemd en $(REMOTE_HOST)..."
	@sshpass -p "$(SSHPASS)" ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "\
		echo '$(SSHPASS)' | sudo -S bash -c '\
			printf \"%s\n\" \
			\"[Unit]\" \
			\"Description=Axolotl SSH Daemon\" \
			\"After=network.target\" \
			\"\" \
			\"[Service]\" \
			\"ExecStart=$(TARGET_PATH)\" \
			\"Restart=always\" \
			\"Slice=axolotl.slice\" \
			\"RestartSec=5\" \
			\"User=root\" \
			\"Delegate=cpu cpuset io memory pids\" \
			\"CapabilityBoundingSet=CAP_SYS_ADMIN CAP_NET_ADMIN CAP_SYS_PTRACE CAP_DAC_OVERRIDE CAP_CHOWN\" \
			\"AmbientCapabilities=CAP_SYS_ADMIN CAP_NET_ADMIN CAP_SYS_PTRACE CAP_DAC_OVERRIDE CAP_CHOWN\" \
			\"WorkingDirectory=/home/$(REMOTE_USER)\" \
			\"StandardOutput=journal\" \
			\"StandardError=journal\" \
			\"\" \
			\"[Install]\" \
			\"WantedBy=multi-user.target\" \
			> /etc/systemd/system/$(SERVICE_NAME)'; \
		echo '$(SSHPASS)' | sudo -S systemctl daemon-reload && \
		echo '$(SSHPASS)' | sudo -S systemctl enable $(SERVICE_NAME)"
	@echo "✅ Servicio systemd configurado."

start:
	@echo "🚀 Iniciando servicio..."
	@sshpass -p "$(SSHPASS)" ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "echo '$(SSHPASS)' | sudo -S systemctl start $(SERVICE_NAME)"
	@echo "✅ Servicio iniciado."

stop:
	@echo "🛑 Deteniendo servicio..."
	@sshpass -p "$(SSHPASS)" ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "echo '$(SSHPASS)' | sudo -S systemctl stop $(SERVICE_NAME)"
	@echo "✅ Servicio detenido."

restart:
	@echo "♻️ Reiniciando servicio..."
	@sshpass -p "$(SSHPASS)" ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "echo '$(SSHPASS)' | sudo -S systemctl restart $(SERVICE_NAME)"
	@echo "✅ Servicio reiniciado."

status:
	@sshpass -p "$(SSHPASS)" ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "echo '$(SSHPASS)' | sudo -S systemctl status $(SERVICE_NAME) --no-pager"

deploy: build upload install service restart status