package builder

import (
	"path/filepath"
	"regexp"
)

const (
	// DotDotSanitizeErrorMessage is the error message used in errors that occur
	// because a provided Bobfile path contains ".."
	DotDotSanitizeErrorMessage = "bobfile path must not contain .."

	// InvalidPathSanitizeErrorMessage is the error message used in errors that
	// occur because a provided Bobfile path is invalid
	InvalidPathSanitizeErrorMessage = "bobfile path is invalid"

	// SymlinkSanitizeErrorMessage is the error message used in errors that
	// occur because a provided Bobfile path contains symlinks
	SymlinkSanitizeErrorMessage = "bobfile path must not contain symlinks"
)

var dotDotRegex = regexp.MustCompile("\\.\\.")

// SanitizeBuilderfilePath checks for disallowed entries in the provided
// Bobfile path and returns either a sanitized version of the path or an error
func SanitizeBuilderfilePath(file string) (string, Error) {
	if dotDotRegex.MatchString(file) {
		return "", &SanitizeError{message: DotDotSanitizeErrorMessage}
	}

	abs, err := filepath.Abs("./" + file)
	if err != nil {
		return "", &SanitizeError{message: InvalidPathSanitizeErrorMessage}
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", &SanitizeError{message: InvalidPathSanitizeErrorMessage}
	}

	if abs != resolved {
		return "", &SanitizeError{message: SymlinkSanitizeErrorMessage}
	}

	clean := filepath.Clean(abs)

	return clean, nil
}