package server

import (
	"axolotl/pkg/model"
	"encoding/json"
	"golang.org/x/crypto/ssh"
	"log"
	"time"
)

func HandleConnection(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "solo sesiones")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("❌ Error al aceptar canal: %v", err)
			continue
		}

		go func(ch ssh.Channel, reqs <-chan *ssh.Request) {
			defer func() {
				log.Println("🔚 Cerrando canal SSH")
				ch.Close()
			}()

			for req := range reqs {
				log.Printf("📦 Request: tipo=%s, payload=%x", req.Type, req.Payload)

				switch req.Type {
				case "exec":
					var payload model.RunRequest
					if err := json.Unmarshal(req.Payload, &payload); err != nil {
						log.Printf("❌ Error decodificando payload: %v", err)
						req.Reply(false, nil)
						return
					}

					log.Printf("🚀 Ejecutando comando con configuración: %+v", payload)

					// Confirmar que se recibió correctamente
					req.Reply(true, nil)

					// Simulación: ejecutar proceso
					ch.Write([]byte("Axolotl ejecutando...\n"))
					time.Sleep(1 * time.Second)
					ch.Write([]byte("✅ Listo\n"))

					// Cerrar después de responder
					return

				default:
					req.Reply(false, nil)
				}
			}
		}(channel, requests)
	}
}
