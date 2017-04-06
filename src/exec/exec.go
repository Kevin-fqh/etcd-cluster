// exec project exec.go
package exec

import (
	"bytes"
	"fmt"
	"os/exec"
)

//Execute a shell command and return the result as a string
func Exec_command(cmd_line string) (string, error) {
	if cmd_line == "" {
		return "", fmt.Errorf("Empty command")
	}
	cmd := exec.Command("/bin/sh", "-c", cmd_line)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Cannot execute %s correctly.", cmd_line)
	}
	return out.String(), nil
}
