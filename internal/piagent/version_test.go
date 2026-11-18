package piagent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in      string
		want    Version
		wantErr bool
	}{
		{"1.5.2", Version{1, 5, 2}, false},
		{"v1.5.2", Version{1, 5, 2}, false},
		{"pi 1.5.2\n", Version{1, 5, 2}, false},
		{"1.10.0-beta.1", Version{1, 10, 0}, false},
		{"", Version{}, true},
		{"no digits here", Version{}, true},
	}
	for _, c := range cases {
		got, err := ParseVersion(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseVersion(%q) = %v, want error", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseVersion(%q): %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("ParseVersion(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestCompatible(t *testing.T) {
	cases := []struct {
		in   string
		want Compat
	}{
		{"1.4.9", CompatTooOld},
		{"1.5.0", CompatOK},
		{"1.5.2", CompatOK},
		{"1.6.0", CompatUntested},
		{"1.12.3", CompatUntested},
		{"2.0.0", CompatTooNew},
		{"0.9.9", CompatTooOld},
	}
	for _, c := range cases {
		v, err := ParseVersion(c.in)
		if err != nil {
			t.Fatalf("ParseVersion(%q): %v", c.in, err)
		}
		if got := v.Compatible(); got != c.want {
			t.Errorf("%s.Compatible() = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestResolveOrder(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "pi")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	// PI_EXECUTABLE wins over PATH.
	got, err := Lookup{
		Env:      []string{"PI_EXECUTABLE=" + exe},
		LookPath: func(string) (string, error) { return "/nope", nil },
	}.Resolve()
	if err != nil || got != exe {
		t.Fatalf("PI_EXECUTABLE: got %q, %v; want %q", got, err, exe)
	}

	// A non-executable PI_EXECUTABLE is an error, not a silent fallback.
	plain := filepath.Join(dir, "plain")
	if err := os.WriteFile(plain, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := (Lookup{Env: []string{"PI_EXECUTABLE=" + plain}}).Resolve(); err == nil {
		t.Fatal("non-executable PI_EXECUTABLE was accepted")
	}

	// PATH next.
	got, err = Lookup{
		Env:      []string{},
		LookPath: func(string) (string, error) { return exe, nil },
	}.Resolve()
	if err != nil || got != exe {
		t.Fatalf("PATH: got %q, %v; want %q", got, err, exe)
	}

	// Login shell last.
	got, err = Lookup{
		Env:      []string{},
		LookPath: func(string) (string, error) { return "", os.ErrNotExist },
		ShellOut: func() (string, error) { return "noise\n" + exe + "\n", nil },
	}.Resolve()
	if err != nil || got != exe {
		t.Fatalf("login shell: got %q, %v; want %q", got, err, exe)
	}

	// Nothing at all.
	if _, err := (Lookup{
		Env:      []string{},
		LookPath: func(string) (string, error) { return "", os.ErrNotExist },
	}).Resolve(); err != ErrNotFound {
		t.Fatalf("empty lookup: %v, want ErrNotFound", err)
	}
}

func TestReadVersionRejectsHangingExecutable(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "slow")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\nsleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if _, err := ReadVersion(ctx, exe); err == nil {
		t.Fatal("ReadVersion accepted a hanging executable")
	}
}
