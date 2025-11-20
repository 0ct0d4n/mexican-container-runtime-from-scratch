package tini

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func StartMainProcess(command string, args ...string) error {
	log.Printf("[INIT] Launching main process: %s %v\n", command, args)

	// 1. Preparar el comando hijo
	cmd := exec.Command(command, args...)

	// Heredar stdio para que se vea en la terminal del contenedor
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 2. Iniciar el proceso hijo
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("[INIT] Failed to start child process: %w", err)
	}

	childPID := cmd.Process.Pid
	log.Printf("[INIT] Child started with PID: %d\n", childPID)

	// 3. Reenviar señales hacia el hijo
	go forwardSignals(childPID)

	// 4. Esperar a que el hijo termine
	err := cmd.Wait()

	// 5. Obtener el exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ProcessState.ExitCode()
		} else {
			log.Printf("[INIT] Error waiting for child: %v\n", err)
			exitCode = 1
		}
	}

	log.Printf("[INIT] Child exited with code %d\n", exitCode)

	// 6. Salir del init con el mismo exit code (así lo hace Docker)
	os.Exit(exitCode)
	return nil
}

func forwardSignals(childPID int) {
	sigs := make(chan os.Signal, 32)

	// Escuchar TODAS las señales, como Tini
	signal.Notify(sigs)

	for sig := range sigs {
		log.Printf("[INIT] Forwarding signal %v to child\n", sig)
		syscall.Kill(childPID, sig.(syscall.Signal))
	}
}
