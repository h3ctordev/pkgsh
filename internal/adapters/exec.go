package adapters

import (
	"bytes"
	"os/exec"
)

// RunCmd ejecuta args[0] con args[1:] sin shell. Devuelve stdout+stderr combinado.
// NUNCA usar exec.Command("sh", "-c", ...) — usar siempre este helper.
func RunCmd(args []string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}

// StreamCmd ejecuta args[0] con args[1:] y escribe stdout+stderr en w.
// Cierra w al terminar.
func StreamCmd(args []string, w interface {
	Write([]byte) (int, error)
	Close() error
}) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = w
	cmd.Stderr = w
	err := cmd.Run()
	w.Close()
	_ = err
}
