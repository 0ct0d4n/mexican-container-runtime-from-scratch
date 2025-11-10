package server

import (
	"axolotl/pkg/model"
	"encoding/json"
	"golang.org/x/crypto/ssh"
	"log"
)

func handleAxorunCommand(ch ssh.Channel, req *ssh.Request) bool {
	decoder := json.NewDecoder(ch)
	var payload model.NamespaceConfig

	if err := decoder.Decode(&payload); err != nil {
		log.Printf("❌ Error decodificando payload: %v", err)
		if err := req.Reply(false, nil); err != nil {
			log.Printf("⚠️ Error al responder: %v", err)
		}
		return true
	}

	log.Printf("🚀 Ejecutando comando con configuración: %+v", payload)

	// Confirmar que se recibió correctamente
	if err := req.Reply(true, nil); err != nil {
		log.Printf("⚠️ Error al responder: %v", err)
		return true
	}

	if _, err := ch.Write([]byte("Axolotl ejecutando...\n")); err != nil {
		log.Printf("⚠️ Error al escribir en canal: %v", err)
		return true
	}

	// configurando recursos
	tmp, err := createCgroup(payload.Cgroup)
	if err != nil {
		log.Printf("❌ [CGROUP] No se pudo crear instancia: %v", err)
		return false
	}
	log.Printf("🦎 [CGROUP] Iniciando configuración en %s", tmp.path)

	// CPU
	if err := tmp.writeCPUMax(); err != nil {
		log.Printf("⚠️ [CGROUP] Error configurando CPU: %v", err)
		return false
	}

	// Memoria
	if err := tmp.writeMemoryMax(); err != nil {
		log.Printf("⚠️ [CGROUP] Error configurando memoria: %v", err)
		return false
	}

	// PIDs
	if err := tmp.writePidsMax(); err != nil {
		log.Printf("⚠️ [CGROUP] Error configurando PIDs: %v", err)
		return false
	}

	if _, err := ch.Write([]byte("✅ Listo\n")); err != nil {
		log.Printf("⚠️ Error al escribir en canal: %v", err)
		return true
	}

	log.Println("✅ [CGROUP] Recursos configurados correctamente")
	// Cerrar después de responder
	return true
}
