package server

import (
	"axolotl/pkg/model"
	"encoding/json"
	"golang.org/x/crypto/ssh"
	"log"
)

func handleAxoRunCommand(ch ssh.Channel, req *ssh.Request) bool {
	log.Println("🦎 [AXO_RUN] Exec request recibido")

	// 1️⃣ Primero validar que el comando es AXO_RUN
	var args struct{ Command string }
	if err := ssh.Unmarshal(req.Payload, &args); err != nil {
		log.Printf("❌ No se pudo leer exec payload: %v", err)
		_ = req.Reply(false, nil)
		return true
	}

	if args.Command != "AXO_RUN" {
		log.Printf("❌ Comando no reconocido: %s", args.Command)
		_ = req.Reply(false, nil)
		return true
	}

	// 2️⃣ Antes de aceptar el exec request, LEER el JSON DESDE STDIN (ch)
	runRequest, err := decodePayloadFromChannel(ch)
	if err != nil {
		log.Printf("❌ Error leyendo payload JSON: %v", err)
		_ = sendMessage(ch, "❌ Payload inválido\n")
		_ = req.Reply(false, nil)
		return true
	}

	if runRequest.Namespace == nil {
		log.Printf("❌ Error: namespace config es nil en el request")
		_ = sendMessage(ch, "❌ Configuración de namespace vacía\n")
		_ = req.Reply(false, nil)
		return true
	}

	payload := runRequest.Namespace
	log.Printf("📦 Payload decodificado correctamente: %+v", payload)

	// 3️⃣ Aceptar el request exec AHORA
	if err := req.Reply(true, nil); err != nil {
		log.Printf("⚠️ Error respondiendo a exec-request: %v", err)
		return true
	}

	// 4️⃣ Respuesta al cliente
	_ = sendMessage(ch, "🦎 Procesando configuración...\n")

	// 5️⃣ Crear cgroup
	tmp, err := createCgroup(payload.Cgroup)
	if err != nil {
		log.Printf("❌ [CGROUP] No se pudo crear instancia: %v", err)
		_ = sendMessage(ch, "❌ [CGROUP] No se pudo crear instancia\n")
		return true
	}
	log.Printf("🦎 [CGROUP] Iniciando configuración en %s", tmp.path)

	// CPU
	if err := tmp.writeCPUMax(); err != nil {
		log.Printf("⚠️ [CGROUP] Error configurando CPU: %v", err)
		_ = sendMessage(ch, "❌ [CGROUP] Error configurando CPU\n")
		return true
	}

	// Memoria
	if err := tmp.writeMemoryMax(); err != nil {
		log.Printf("⚠️ [CGROUP] Error configurando memoria: %v", err)
		_ = sendMessage(ch, "❌ [CGROUP] Error configurando memoria\n")
		return true
	}

	// PIDs
	if err := tmp.writePidsMax(); err != nil {
		log.Printf("⚠️ [CGROUP] Error configurando PIDs: %v", err)
		_ = sendMessage(ch, "❌ [CGROUP] Error configurando PIDs\n")
		return true
	}

	_ = sendMessage(ch, "✅ Configuración completada exitosamente\n")
	log.Println("✅ [CGROUP] Recursos configurados correctamente")

	return true
}

// -----------------------------------------------------------------------------
// ✔️ Lee el JSON desde EL CANAL (stdin del cliente)
// El cliente envía un RunRequest completo con el campo "namespaces"
// -----------------------------------------------------------------------------
func decodePayloadFromChannel(ch ssh.Channel) (*model.RunRequest, error) {
	var runRequest model.RunRequest

	decoder := json.NewDecoder(ch)
	if err := decoder.Decode(&runRequest); err != nil {
		log.Printf("❌ Error leyendo JSON desde stdin: %v", err)
		return nil, err
	}

	log.Printf("🔍 RunRequest raw: Command=%v, Namespace=%+v", runRequest.Command, runRequest.Namespace)
	return &runRequest, nil
}

// -----------------------------------------------------------------------------
// ✔️ Escribe mensajes en el canal SSH
// -----------------------------------------------------------------------------
func sendMessage(ch ssh.Channel, msg string) error {
	if _, err := ch.Write([]byte(msg)); err != nil {
		log.Printf("⚠️ Error al escribir en canal SSH: %v", err)
		return err
	}
	return nil
}
