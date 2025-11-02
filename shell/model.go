package shell

import (
	"bytes"
	"io"
	"os/exec"
	"sync"
	"time"
)

// PaneOutput represents a piece of output from a shell within a pane,
// providing context about its origin.
type PaneOutput struct {
	ShellID   string
	Timestamp time.Time
	Data      []byte
	IsStderr  bool
}

// ShellSession represents an active, managed shell process.
// It holds references to the process's I/O streams and buffers for capturing output.
type ShellSession struct {
	ID          string         // Unique identifier for the session.
	Cmd         *exec.Cmd      // The underlying command process.
	Stdin       io.WriteCloser // Pipe for writing to the shell's standard input.
	Stdout      io.ReadCloser  // Pipe for reading from the shell's standard output.
	Stderr      io.ReadCloser  // Pipe for reading from the shell's standard error.
	StartedAt   time.Time      // Timestamp of when the session was created.
	Interactive bool           // Tracks if the shell is interactive.
	OutputBuf   bytes.Buffer   // Buffer to capture stdout.
	StderrBuf   bytes.Buffer   // Buffer to capture stderr.
	mu          sync.Mutex     // Mutex to protect concurrent access to session buffers.
}
