package server

import (
	"axolotl/pkg/model"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"path/filepath"
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

	// creando cgroup
	base := "/sys/fs/cgroup/"
	path := filepath.Join(base, payload.Cgroup.Path)
	memoryMaxPath := filepath.Join(path, "memory.max")
	cpuMaxPath := filepath.Join(path, "cpu.max")
	pidMaxPath := filepath.Join(path, "pids.max")

	os.MkdirAll(path, 0755)
	fmt.Printf("🦎 Creando cgroup en %s: Mem=%d, CPU=%d, PIDs=%d\n", path, payload.Cgroup.MemoryMax, payload.Cgroup.CPUMax, payload.Cgroup.PidsMax)

	// configurando recursos

	if _, err := ch.Write([]byte("✅ Listo\n")); err != nil {
		log.Printf("⚠️ Error al escribir en canal: %v", err)
		return true
	}

	// Cerrar después de responder
	return true
}
