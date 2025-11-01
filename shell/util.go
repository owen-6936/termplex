package shell

import "regexp"

var (
	// ansiRegex is a regular expression to find and remove ANSI escape codes.
	ansiRegex = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*.?[a-zA-Z0-9]")
)

// StripANSI removes all ANSI escape codes from a byte slice.
func StripANSI(data []byte) []byte {
	return ansiRegex.ReplaceAll(data, []byte{})
}
