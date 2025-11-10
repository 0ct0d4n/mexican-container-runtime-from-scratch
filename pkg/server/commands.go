package server

import (
	"axolotl/pkg/model"
	"encoding/json"
	"golang.org/x/crypto/ssh"
	"log"
	"time"
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

	// Simulación: ejecutar proceso
	if _, err := ch.Write([]byte("Axolotl ejecutando...\n")); err != nil {
		log.Printf("⚠️ Error al escribir en canal: %v", err)
		return true
	}

	time.Sleep(1 * time.Second)

	if _, err := ch.Write([]byte("✅ Listo\n")); err != nil {
		log.Printf("⚠️ Error al escribir en canal: %v", err)
		return true
	}

	// Cerrar después de responder
	return true
}
