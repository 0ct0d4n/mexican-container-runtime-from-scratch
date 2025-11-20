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

	// 1. Prepare child command
	cmd := exec.Command(command, args...)

	// Inherit stdio so output is visible in container terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 2. Start child process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("[INIT] Failed to start child process: %w", err)
	}

	childPID := cmd.Process.Pid
	log.Printf("[INIT] Child started with PID: %d\n", childPID)

	// 3. Forward signals to child
	go forwardSignals(childPID)

	// 4. Wait for child to exit
	err := cmd.Wait()

	// 5. Get exit code
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

	// 6. Exit init with same exit code (Docker behavior)
	os.Exit(exitCode)
	return nil
}

func forwardSignals(childPID int) {
	sigs := make(chan os.Signal, 32)

	// Listen to ALL signals, like Tini
	signal.Notify(sigs)

	for sig := range sigs {
		log.Printf("[INIT] Forwarding signal %v to child\n", sig)
		syscall.Kill(childPID, sig.(syscall.Signal))
	}
}
