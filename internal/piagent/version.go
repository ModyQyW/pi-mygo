package piagent

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Version is a pi version, as "1.5.2".
type Version struct {
	Major, Minor, Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Series is the major version as pi's own releases describe it, "1.x".
func (v Version) Series() string {
	return fmt.Sprintf("%d.x", v.Major)
}

// Min is the oldest pi this app supports. A pi with another major version is
// refused until an app release says otherwise.
var Min = Version{1, 5, 0}

// Tested is the newest pi minor the app was tested against. Newer minors of the
// same major run, with a "not tested yet" note.
var Tested = Version{1, 5, 2}

// Compat is how well the installed pi fits this app.
type Compat string

const (
	// CompatOK: tested range.
	CompatOK Compat = "ok"
	// CompatUntested: same major, newer than Tested.
	CompatUntested Compat = "untested"
	// CompatTooOld: older than Min.
	CompatTooOld Compat = "too_old"
	// CompatTooNew: a different major than Min.
	CompatTooNew Compat = "too_new"
)

var versionPattern = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

// ParseVersion reads the first semantic version in s, tolerating the "v"
// prefix and surrounding text that shells and wrappers add.
func ParseVersion(s string) (Version, error) {
	m := versionPattern.FindStringSubmatch(s)
	if m == nil {
		return Version{}, fmt.Errorf("no version in %q", strings.TrimSpace(s))
	}
	nums := make([]int, 3)
	for i := range nums {
		n, err := strconv.Atoi(m[i+1])
		if err != nil {
			return Version{}, fmt.Errorf("version in %q: %w", s, err)
		}
		nums[i] = n
	}
	return Version{nums[0], nums[1], nums[2]}, nil
}

// Compatible reports how this version fits the supported range.
func (v Version) Compatible() Compat {
	switch {
	case v.Major < Min.Major || (v.Major == Min.Major && v.less(Min)):
		return CompatTooOld
	case v.Major != Min.Major:
		return CompatTooNew
	case v.greater(Tested):
		return CompatUntested
	default:
		return CompatOK
	}
}

func (v Version) less(o Version) bool {
	if v.Major != o.Major {
		return v.Major < o.Major
	}
	if v.Minor != o.Minor {
		return v.Minor < o.Minor
	}
	return v.Patch < o.Patch
}

func (v Version) greater(o Version) bool { return o.less(v) }

// ReadVersion runs `pi --version` and parses the result. A slow or hanging
// executable fails instead of blocking startup.
func ReadVersion(ctx context.Context, path string) (Version, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "--version")
	// A wrapper that spawns children keeps the pipe open after the child is
	// killed; without WaitDelay, Output would wait for the grandchild.
	cmd.WaitDelay = 500 * time.Millisecond
	out, err := cmd.Output()
	if err != nil {
		return Version{}, fmt.Errorf("%s --version: %w", path, err)
	}
	return ParseVersion(string(out))
}
