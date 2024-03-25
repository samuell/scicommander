package main

import (
	"fmt"
	"os/exec"

	"github.com/abiosoft/ishell"
)

func main() {
	sh := ishell.New()
	sh.Println("SciCommander")
	sh.Println("--------------------------------------------")
	sh.Println("Type '$' and hit enter, to start session")

	sh.AddCmd(&ishell.Cmd{
		Name: "$",
		Help: "Execute command",
		Func: func(c *ishell.Context) {
			for {
				cmd := c.ReadLine()
				c.Println("$ ", cmd)
				execCmd(cmd)
			}
		},
	})

	sh.Run()
}

func execCmd(cmd string) {
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("Command failed!\nCommand:\n%s\n\nOutput:\n%s\nOriginal error:%s", cmd, string(out), err))
	}
	fmt.Printf(string(out))
}
