package adapters

import (
	"bytes"
	"io"
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

// StreamCmdStdin ejecuta args[0] con args[1:] con stdin controlado externamente.
// Cierra w al terminar. NUNCA usar exec.Command("sh", "-c", ...).
//
// Usa cmd.StdinPipe() en lugar de cmd.Stdin para que cmd.Wait() no bloquee
// esperando que la goroutine de copia de stdin termine. Esto evita que la
// operación quede colgada cuando sudo tiene credenciales en caché y nunca
// necesita leer de stdin.
func StreamCmdStdin(args []string, stdin io.ReadCloser, w interface {
	Write([]byte) (int, error)
	Close() error
}) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = w
	cmd.Stderr = w

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		w.Close()
		return
	}

	if err := cmd.Start(); err != nil {
		w.Close()
		return
	}

	// Pump stdin: forwarding from our pipe to the process.
	// Runs until stdin is closed (CloseStdin) or the write fails (process exited).
	go func() {
		io.Copy(stdinPipe, stdin)
		stdinPipe.Close()
	}()

	cmd.Wait() // does NOT wait for the stdin pump goroutine above
	stdin.Close()  // unblock the pump goroutine if it's still running
	w.Close()
}
