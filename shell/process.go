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
		// In a PTY, stdout and stderr are the same file. The stderr handler
		// will be called for all output in this case.
		handler := stdoutHandler
		if s.Stdout == s.Stderr {
			handler = stderrHandler
		}

		for {
			buf := make([]byte, 1024)
			n, err := s.Stdout.Read(buf)
			if n > 0 {
				handler(buf[:n])
			}
			if err != nil {
				// Don't print an error if the file is intentionally closed.
				if err != io.EOF && !strings.Contains(err.Error(), "file already closed") {
					fmt.Printf("Error reading from session %s: %v\n", s.ID, err)
				}
				return
			}
		}
	}()

	// Only start a separate stderr goroutine if it's a different pipe.
	if s.Stderr != nil && s.Stderr != s.Stdout {
		go func() {
			defer s.Stderr.Close()
			for {
				buf := make([]byte, 1024)
				n, err := s.Stderr.Read(buf)
				if n > 0 {
					stderrHandler(buf[:n])
				}
				if err != nil {
					if err != io.EOF && !strings.Contains(err.Error(), "file already closed") {
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
	// Use `echo -n` to prevent a trailing newline from the delimiter itself.
	// Combine with the user's command using '&&' to run sequentially.
	// This prevents an intermediate shell prompt from leaking into the output.
	fullCommand := fmt.Sprintf("%s && echo -n %s", command, delimiter)

	if err := s.SendCommand(fullCommand + "\n"); err != nil {
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
				// Split the output by the delimiter and return only what came before it.
				// This cleanly separates the command's output from our delimiter.
				content := strings.SplitN(output, delimiter, 2)[0]
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
