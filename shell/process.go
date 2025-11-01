package shell

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StartReading launches goroutines to read from a session's stdout and stderr.
// This enables non-blocking, real-time output streaming.
func (s *ShellSession) StartReading(
	stdoutHandler func(output []byte),
	stderrHandler func(output []byte),
) {
	// Goroutine for stdout
	go func() {
		defer s.Stdout.Close()
		buf := make([]byte, 1024)
		for {
			n, err := s.Stdout.Read(buf)
			if n > 0 {
				stdoutHandler(buf[:n])
			}
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading stdout from session %s: %v\n", s.ID, err)
				}
				return
			}
		}
	}()

	// Goroutine for stderr
	if s.Stderr != nil {
		go func() {
			defer s.Stderr.Close()
			buf := make([]byte, 1024)
			for {
				n, err := s.Stderr.Read(buf)
				if n > 0 {
					stderrHandler(buf[:n])
				}
				if err != nil {
					if err != io.EOF {
						fmt.Printf("Error reading stderr from session %s: %v\n", s.ID, err)
					}
					return
				}
			}
		}()
	}
}

// SendCommand writes a command string to the shell's standard input.
func (s *ShellSession) SendCommand(command string) error {
	if s.Stdin == nil {
		return fmt.Errorf("session %s has no stdin", s.ID)
	}
	_, err := s.Stdin.Write([]byte(command + "\n"))
	return err
}

// SendCommandAndWait sends a command and blocks until a unique delimiter is found in the output.
func (s *ShellSession) SendCommandAndWait(command string) (string, error) {
	s.mu.Lock()
	s.OutputBuf.Reset()
	s.mu.Unlock()

	delimiter := uuid.New().String()
	fullCommand := fmt.Sprintf("%s\n%s\n", command, "echo "+delimiter)

	if err := s.SendCommand(fullCommand); err != nil {
		return "", fmt.Errorf("failed to write command to session %s: %w", s.ID, err)
	}

	timeout := time.After(300 * time.Second)
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timed out waiting for command delimiter in session %s", s.ID)
		case <-tick.C:
			s.mu.Lock()
			output := s.OutputBuf.String()
			s.mu.Unlock()

			if strings.Contains(output, delimiter) {
				content := strings.Split(output, delimiter)[0]
				return content, nil
			}
		}
	}
}

// OutputHandler is a default handler that processes raw byte output from the shell.
// It appends the output to the session's buffer and prints it to the console.
func (s *ShellSession) OutputHandler(output []byte) {
	// Convert the raw bytes to a string for display
	s.mu.Lock()
	s.OutputBuf.Write(output)
	s.mu.Unlock()

	// Print the output clearly labeled with the session ID
	fmt.Print(string(output))
}

// ErrorOutputHandler is a handler that processes raw byte output from the shell's stderr.
// It appends the output to the session's StderrBuf and prints it to the console as an error.
func (s *ShellSession) ErrorOutputHandler(output []byte) {
	s.mu.Lock()
	s.StderrBuf.Write(output)
	s.mu.Unlock()

	// Print the error output clearly labeled with the session ID
	fmt.Print(string(output))
}

// Close gracefully terminates the shell session.
func (s *ShellSession) Close(gracePeriod time.Duration) error {
	if s.Cmd == nil || s.Cmd.Process == nil {
		return nil // Nothing to close
	}

	if s.Stdin != nil {
		_ = s.Stdin.Close() // Signal process to exit
	}

	done := make(chan error, 1)
	go func() { done <- s.Cmd.Wait() }()

	select {
	case <-time.After(gracePeriod):
		// The grace period expired. Force-kill the process.
		if err := s.Cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process after timeout: %w", err)
		}
		// Wait for the original Wait() call to return with the kill error.
		return <-done
	case err := <-done:
		// Process exited gracefully within the grace period.
		return err
	}
}
