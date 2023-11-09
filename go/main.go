package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/creack/pty" // You need to get this package for pty functionality
)

func main() {
	// Start a bash session
	c := exec.Command("bash")
	f, err := pty.Start(c)
	if err != nil {
		panic(err)
	}

	// This is a simple read-eval-print loop (REPL)
	reader := bufio.NewReader(os.Stdin)

	for {
		// Read the command from the user
		command, _ := reader.ReadString('\n')
		fmt.Printf("You executed: %s\n", command)

		// Send the command to bash
		_, err = f.Write([]byte(command))
		if err != nil {
			panic(err)
		}

		// Read the output from bash
		buf := make([]byte, 1024)
		_, err = f.Read(buf)
		if err != nil {
			panic(err)
		}

		// Print the output to the user
		fmt.Printf("You received: %s\n", buf)
	}
}
