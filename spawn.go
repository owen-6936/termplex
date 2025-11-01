package termplex

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	uuid "github.com/google/uuid"
)

// ShellSession represents an active shell process, which can be either a standalone
// process managed by exec.Cmd or a shell running inside a tmux pane. It holds
// references to the process's I/O streams and buffers for capturing output.
type ShellSession struct {
	ID        string         // Unique identifier for the session.
	Cmd       *exec.Cmd      // The underlying command process, nil for pane-based shells.
	PaneRef   *Pane          // Reference to the tmux pane if the shell is running inside one.
	Stdin     io.WriteCloser // Pipe for writing to the shell's standard input.
	Stdout    io.ReadCloser  // Pipe for reading from the shell's standard output.
	Stderr    io.ReadCloser  // Pipe for reading from the shell's standard error.
	CreatedAt time.Time      // Timestamp of when the session was created.
	OutputBuf bytes.Buffer   // Buffer to capture stdout.
	StderrBuf bytes.Buffer   // Buffer to capture stderr.
	mu        sync.Mutex     // Mutex to protect concurrent access to session buffers.
}

var (
	mu       sync.Mutex
	sessions = make(map[string]*ShellSession)
)

// NewShell creates, starts, and registers a new standalone interactive bash session
// using `os/exec`. It returns a unique session ID for future interactions.
// This is suitable for processes that do not need to be managed within a tmux layout.
func NewShell() (string, error) {
	cmd := exec.Command("bash", "-i")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	id := uuid.New().String()

	session := &ShellSession{
		ID:        id,
		Cmd:       cmd,
		Stdin:     stdin,
		Stdout:    stdout,
		CreatedAt: time.Now(),
		OutputBuf: *bytes.NewBuffer(nil),
		// Stderr is not captured for basic shells, only for command shells.
		StderrBuf: *bytes.NewBuffer(nil),
	}

	mu.Lock()
	sessions[session.ID] = session
	mu.Unlock()

	fmt.Printf("🧠 New shell started: %s\n", id)
	return id, nil
}

// NewPaneShell seeds a new shell inside an existing tmux pane and returns a
// `ShellSession` to represent it. This allows the shell to be tracked and
// interacted with via the session management system, even though it's managed
// by tmux. The `shellType` defaults to "bash" if empty.
func NewPaneShell(pane *Pane, shellType string) (*ShellSession, error) {
    if shellType == "" {
        shellType = "bash"
    }

    // Send exec shellType to the pane
    err := pane.SendKeys("exec " + shellType)
    if err != nil {
        return nil, fmt.Errorf("failed to seed shell in pane: %w", err)
    }

    // Create a ShellSession identity to track this shell
    id := uuid.New().String()
    session := &ShellSession{
        ID:        id,
        CreatedAt: time.Now(),
        PaneRef:   pane, // Link back to the pane
    }

    // Register the shell identity
    mu.Lock()
    sessions[id] = session
    mu.Unlock()

    // Append to pane's ShellSessions slice
    pane.ShellSessions = append(pane.ShellSessions, session)

    fmt.Printf("🧠 Pane shell seeded: %s in %s\n", shellType, pane.Target())
    return session, nil
}

// NewShellWithCommand creates, starts, and registers a new standalone session with a
// custom command (e.g., `python`, `node`). This is useful for launching and
// managing long-running background processes or scripts. It captures both stdout
// and stderr and returns a unique session ID.
func NewShellWithCommand(command ...string) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("NewShellWithCommand requires a command to execute")
	}
	cmd := exec.Command(command[0], command[1:]...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	// Create a separate pipe for stderr to distinguish errors from normal output.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	id := uuid.New().String()
	session := &ShellSession{
		ID:        id,
		Cmd:       cmd,
		Stdin:     stdin,
		Stdout:    stdout,
		Stderr:    stderr,
		CreatedAt: time.Now(),
		OutputBuf: *bytes.NewBuffer(nil),
		StderrBuf: *bytes.NewBuffer(nil),
	}

	mu.Lock()
	sessions[id] = session
	mu.Unlock()

	fmt.Printf("🧠 New command shell started: %s with command %v\n", id, command)
	return id, nil
}

// SendCommand writes a command string to the standard input of a specific shell
// session. It automatically appends a newline character to execute the command.
// This function is non-blocking.
func SendCommand(sessionID string, command string) error {
	mu.Lock()
	session, ok := sessions[sessionID] // Use package-level sessions
	mu.Unlock()

	if !ok {
		return fmt.Errorf("shell session %s not found", sessionID)
	}

	// Append a newline to ensure the shell executes the command
	commandWithNewline := command + "\n"

	// Write the command bytes to the shell's input pipe
	_, err := session.Stdin.Write([]byte(commandWithNewline))

	return err
}

// SendCommandAndWait sends a command to a session and blocks until the command
// has finished executing. It achieves this by appending a second command that
// prints a unique delimiter, and then waiting for that delimiter to appear in
// the standard output. It returns all output generated by the original command.
func SendCommandAndWait(sessionID string, command string) (string, error) {
	mu.Lock()
	session, ok := sessions[sessionID]
	mu.Unlock()

	if !ok {
		return "", fmt.Errorf("SendCommandAndWait: shell session %s not found", sessionID)
	}

	// Clear the buffer before sending a new command to ensure we only capture the new output.
	session.mu.Lock()
	session.OutputBuf.Reset()
	session.mu.Unlock()

	// Generate a unique delimiter for this specific command execution.
	delimiter := uuid.New().String()
	commandToProduceDelimiter := "echo " + delimiter

	// Combine the user's command with our delimiter-producing command.
	// This ensures the delimiter is printed only after the user's command completes.
	fullCommand := fmt.Sprintf("%s\n%s\n", command, commandToProduceDelimiter)

	// Write the combined command to the shell's input pipe.
	_, err := session.Stdin.Write([]byte(fullCommand))
	if err != nil {
		return "", fmt.Errorf("failed to write command to session %s: %w", sessionID, err)
	}

	// Wait for the response by polling the buffer until the delimiter is found.
	timeout := time.After(300 * time.Second) // 5-minute timeout for the command to respond
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timed out waiting for command delimiter in session %s", sessionID)
		case <-tick.C:
			if output, ok := checkBufferForDelimiter(session, delimiter); ok {
				return output, nil
			}
		}
	}
}

// SendCommandFromReady sends a command to a session that is already in a "ready"
// state (e.g., after printing an initial prompt) and waits for a `delimiter` in
// the subsequent output. Unlike `SendCommandAndWait`, it does not clear the
// output buffer first, making it suitable for interacting with processes that
// provide an initial startup message before accepting commands.
func SendCommandFromReady(sessionID string, command string, delimiter string) (string, error) {
	mu.Lock()
	session, ok := sessions[sessionID]
	mu.Unlock()

	if !ok {
		return "", fmt.Errorf("shell session %s not found", sessionID)
	}

	// Do not reset the buffer, as it contains the initial ready prompt.
	// Instead, we will capture the output from this point forward.

	commandWithNewline := command + "\n"
	_, err := session.Stdin.Write([]byte(commandWithNewline))
	if err != nil {
		return "", err
	}

	// Wait for the response by polling the buffer until the delimiter is found.
	timeout := time.After(180 * time.Second) // 3-minute timeout for the command to respond
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timed out waiting for response delimiter: %s", delimiter)
		case <-tick.C:
			if output, ok := checkBufferForDelimiter(session, delimiter); ok {
				return output, nil
			}
		}
	}
}

// OutputHandler is a callback function for processing stdout from a shell session.
// It appends the output to the session's `OutputBuf` and prints it to the
// console, providing real-time feedback.
func OutputHandler(output []byte, sessionID string, session *ShellSession) {
	// Convert the raw bytes to a string for display
	session.mu.Lock()
	session.OutputBuf.Write(output)
	session.mu.Unlock()

	// Print the output clearly labeled with the session ID
	fmt.Print(string(output))
}

// ErrorOutputHandler is a callback function for processing stderr from a shell
// session. It appends the error output to the session's `StderrBuf` and prints
// it to the console.
func ErrorOutputHandler(output []byte, sessionID string, session *ShellSession) {
	session.mu.Lock()
	session.StderrBuf.Write(output)
	session.mu.Unlock()

	// Print the error output clearly labeled with the session ID
	fmt.Print(string(output))
}

// InfoOutputHandler is a callback for processing stderr streams that may contain
// non-error information, such as progress updates or logs. It appends the data
// to the `StderrBuf` and prints it to the console.
func InfoOutputHandler(output []byte, sessionID string, session *ShellSession) {
	session.mu.Lock()
	session.StderrBuf.Write(output)
	session.mu.Unlock()

	fmt.Print(string(output))
}

// StartReading launches goroutines to continuously read from a session's stdout
// and stderr pipes. It uses the provided handler functions to process the data
// as it arrives. This enables non-blocking, real-time output streaming from the
// underlying process.
func StartReading(sessionID string, stdoutHandler func(output []byte, sessionID string, session *ShellSession), stderrHandler func(output []byte, sessionID string, session *ShellSession)) error {
	// Use a buffer for efficient reading
	mu.Lock()
	session, ok := sessions[sessionID]
	mu.Unlock()

	if !ok {
		return fmt.Errorf("session %s not found for starting reader", sessionID)
	}

	buf := make([]byte, 1024)
	// Launch the dedicated reader goroutine
	go func() {
		defer session.Stdout.Close()
		for {
			// Read blocks until data is available or the pipe closes
			n, err := session.Stdout.Read(buf)
			if n > 0 {
				stdoutHandler(buf[:n], session.ID, session)
			}

			if err != nil {
				// io.EOF is expected when the shell exits normally
				if err != io.EOF {
					fmt.Printf("Error reading from session %s: %v\n", session.ID, err)
				} else {
					fmt.Printf("Shell session %s finished (EOF).\n", session.ID)
					return
					// EOF is expected when the shell exits normally.
					// The process finishing will be logged by the stderr handler if it exits with an error.
				}
				// Once the pipe closes (EOF or error), the goroutine exits
				return
			}
		}
	}()

	// Launch a dedicated reader goroutine for stderr if it exists
	if session.Stderr != nil {
		go func() {
			defer session.Stderr.Close()
			errBuf := make([]byte, 1024)
			for {
				n, err := session.Stderr.Read(errBuf)
				if n > 0 {
					stderrHandler(errBuf[:n], session.ID, session)
				}

				if err != nil {
					if err != io.EOF {
						fmt.Printf("Error reading from session stderr %s: %v\n", session.ID, err)
					} else {
						// EOF is expected when the process closes its stderr.
					}
					return
				}
			}
		}()
	}
	return nil
}

// CloseSession gracefully terminates a shell session. For pane-based shells, it
// sends an "exit" command via tmux. For standalone `exec.Cmd` shells, it closes
// the stdin pipe, which typically causes the shell to exit, and then waits for
// the process to terminate.
func CloseSession(sessionID string) error {
    session, ok := GetSession(sessionID)
    if !ok {
        return fmt.Errorf("shell session %s not found", sessionID)
    }

    // If it's a tmux-pane shell, send 'exit' via SendKeys
    if session.PaneRef != nil {
        err := session.PaneRef.SendKeys("exit")
        if err != nil {
            return fmt.Errorf("failed to send exit to pane shell: %w", err)
        }
        // Optionally: wait for command to change or buffer to flush
        return nil
    }

    // External shell: close stdin and wait for termination
    if session.Stdin != nil {
        _ = session.Stdin.Close()
    }
    if session.Cmd != nil {
        return session.Cmd.Wait()
    }

    return nil
}

// ForceTerminateSession immediately kills the underlying process of a session using
// `os.Process.Kill()`. This should be used as a last resort when a graceful
// `CloseSession` fails or is not possible. It does not apply to pane-based
// shells, which are managed by tmux.
func ForceTerminateSession(sessionID string) error {
    session, ok := GetSession(sessionID)
    if !ok || session.Cmd == nil || session.Cmd.Process == nil {
        return fmt.Errorf("cannot force terminate: session not found or invalid")
    }
    return session.Cmd.Process.Kill()
}

// CloseSessionWithFallback attempts to gracefully close a session and falls back to
// force-termination if the process does not exit within the specified grace period.
// This is particularly useful for ensuring that long-running child processes are
// properly cleaned up.
func CloseSessionWithFallback(sessionID string, gracePeriod time.Duration) error {
	session, ok := GetSession(sessionID)
	if !ok {
		return fmt.Errorf("shell session %s not found", sessionID)
	}

	// For tmux panes, a simple 'exit' command is sufficient.
	if session.PaneRef != nil {
		return CloseSession(sessionID)
	}

	// For standalone processes, attempt a graceful shutdown first.
	if session.Cmd == nil || session.Cmd.Process == nil {
		return fmt.Errorf("session %s has no active process to close", sessionID)
	}

	// Close stdin to signal the process to exit.
	if session.Stdin != nil {
		_ = session.Stdin.Close()
	}

	// Wait for the process to exit, with a timeout.
	done := make(chan error, 1)
	go func() {
		done <- session.Cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Printf("⚠️ Session %s exited with error: %v\n", sessionID, err)
		} else {
			fmt.Printf("✅ Shell %s terminated gracefully.\n", sessionID)
		}
		return err
	case <-time.After(gracePeriod):
		fmt.Printf("⚠️ Shell %s did not exit within %s. Forcing termination...\n", sessionID, gracePeriod)
		return ForceTerminateSession(sessionID)
	}
}

// GetSession safely retrieves a session from the global map by its ID. It returns
// the session and a boolean indicating whether the session was found. This
// function is thread-safe.
func GetSession(sessionID string) (*ShellSession, bool) {
	mu.Lock()
	session, ok := sessions[sessionID] // Use package-level sessions
	mu.Unlock()
	return session, ok
}

// IsRunning checks if the underlying process for a standalone session is still
// active. It returns false if the session is not found or if the process has exited.
func IsRunning(sessionID string) bool {
	session, ok := GetSession(sessionID)
	if !ok {
		return false
	}

	// Check if the underlying process is still running.
	// If Cmd.ProcessState is nil, it usually means the command is still running.
	return session.Cmd.ProcessState == nil || !session.Cmd.ProcessState.Exited()
}

// WaitForString polls a session's output buffer until a `target` string is
// found or a `timeout` is reached. This is useful for synchronizing with a
// process by waiting for a specific signal in its output, such as "Ready" or
// "Server started".
func WaitForString(sessionID string, target string, timeout time.Duration) error {
	startTime := time.Now()
	for {
		if time.Since(startTime) > timeout {
			mu.Lock()
			session, ok := sessions[sessionID]
			var output, stderrOutput string
			if ok {
				output = session.OutputBuf.String()
				stderrOutput = session.StderrBuf.String()
			}
			mu.Unlock()
			return fmt.Errorf("timed out waiting for string '%s'.\nLast stdout: %s\nLast stderr: %s", target, output, stderrOutput)
		}

		time.Sleep(200 * time.Millisecond) // Poll every 200ms

		mu.Lock()
		session, ok := sessions[sessionID]
		if !ok {
			mu.Unlock()
			return fmt.Errorf("session %s not found while waiting for string", sessionID)
		}
		output := session.OutputBuf.String()
		mu.Unlock()

		if strings.Contains(output, target) {
			return nil
		}
	}
}

// checkBufferForDelimiter is a helper function that inspects a session's output
// buffer for a given delimiter. If found, it returns the content preceding the
// delimiter and a boolean `true`. Otherwise, it returns an empty string and `false`.
func checkBufferForDelimiter(session *ShellSession, delimiter string) (string, bool) {
	session.mu.Lock()
	defer session.mu.Unlock()
	output := session.OutputBuf.String()
	if strings.Contains(output, delimiter) {
		parts := strings.Split(output, delimiter)
		// If delimiter is found, there will be at least two parts.
		// The content we want is everything *before* the last occurrence of the delimiter.
		if len(parts) > 1 {
			// Join all but the last part, in case the output itself contains the delimiter
			content := strings.Join(parts[:len(parts)-1], delimiter)
			return content, true
		}
	}
	return "", false
}