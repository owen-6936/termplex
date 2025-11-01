package manifest

// Manifest defines the structure for a .termplex.json file, representing a complete
// orchestration session.
type Manifest struct {
	SessionName string            `json:"sessionName"`
	SessionTags map[string]string `json:"sessionTags"`
	Windows     []WindowManifest  `json:"windows"`
}

// WindowManifest describes a single window to be created within a session.
type WindowManifest struct {
	WindowName string            `json:"windowName"`
	WindowTags map[string]string `json:"windowTags"`
	Panes      []PaneManifest    `json:"panes"`
}

// PaneManifest describes a single pane to be created within a window.
type PaneManifest struct {
	PaneTags        map[string]string `json:"paneTags"`
	StartupShell    ShellManifest     `json:"startupShell"`
	StartupCommands []string          `json:"startupCommands"`
}

// ShellManifest describes the shell process to be spawned in a pane.
type ShellManifest struct {
	Interactive bool     `json:"interactive"`
	Command     []string `json:"command"`
}
