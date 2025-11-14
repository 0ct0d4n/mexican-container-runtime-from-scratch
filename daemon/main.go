package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init-container" {
		// Run the container initialization sequence
		if err := runContainerInit(); err != nil {
			fmt.Println("❌ init-container failed:", err)
			os.Exit(1)
		}

		// If runContainerInit() does execve(), we never return here.
		return
	}

	// Otherwise this is the daemon process
	startDaemonMode()
}
