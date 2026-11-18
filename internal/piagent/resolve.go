// Package piagent talks to an installed pi coding agent: it finds the
// executable, checks its version, and speaks the RPC protocol over stdio.
//
// It has no knowledge of the UI. Everything here is usable and testable
// without a window.
package piagent

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Lookup describes how the executable is found, so the search order is
// testable without a real installation.
type Lookup struct {
	// Env is the process environment. PI_EXECUTABLE overrides the search.
	Env []string
	// LookPath is exec.LookPath; nil means the real one.
	LookPath func(string) (string, error)
	// ShellOut finds pi through the user's login shell: used when a GUI app,
	// launched from the Dock, inherits a PATH that lacks an nvm/asdf/volta
	// directory.
	ShellOut func() (string, error)
}

// ErrNotFound means no pi executable was found anywhere.
var ErrNotFound = errors.New("pi executable not found")

// Resolve returns the absolute path of the pi executable.
//
// Order: PI_EXECUTABLE, PATH, then the user's login shell. The last step exists
// because a macOS app launched from the Dock inherits a minimal PATH, while pi
// may live in a directory only the shell's startup files add.
func (l Lookup) Resolve() (string, error) {
	if v := envValue(l.Env, "PI_EXECUTABLE"); v != "" {
		abs, err := filepath.Abs(v)
		if err != nil {
			return "", err
		}
		if !isExecutable(abs) {
			return "", fmt.Errorf("PI_EXECUTABLE is set to %s, which is not an executable file", v)
		}
		return abs, nil
	}

	lookPath := l.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if p, err := lookPath("pi"); err == nil && p != "" {
		if abs, err := filepath.Abs(p); err == nil {
			return abs, nil
		}
		return p, nil
	}

	if l.ShellOut != nil {
		out, err := l.ShellOut()
		if err == nil {
			for line := range strings.SplitSeq(out, "\n") {
				line = strings.TrimSpace(line)
				if line != "" && isExecutable(line) {
					return line, nil
				}
			}
		}
	}
	return "", ErrNotFound
}

// Shell resolves with the real login shell.
func Shell() Lookup {
	return Lookup{Env: os.Environ(), ShellOut: LoginShellOut}
}

// LoginShellOut finds pi in a login shell and returns stdout.
//
// The command is fixed: callers never supply shell text, so no user input can
// reach the shell.
func LoginShellOut() (string, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	out, err := exec.Command(shell, "-lc", "command -v pi").Output()
	if err != nil {
		return "", fmt.Errorf("login shell lookup: %w", err)
	}
	return string(out), nil
}

func envValue(env []string, key string) string {
	for _, kv := range env {
		if v, ok := strings.CutPrefix(kv, key+"="); ok {
			return v
		}
	}
	return ""
}

func isExecutable(path string) bool {
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return false
	}
	return st.Mode()&0o111 != 0
}
