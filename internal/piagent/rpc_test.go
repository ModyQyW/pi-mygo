package piagent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// lineSep and paraSep are U+2028 and U+2029: legal inside JSON strings, and the
// reason a client must not use a reader that treats them as record boundaries.
var (
	lineSep = string(rune(0x2028))
	paraSep = string(rune(0x2029))
)

// The test binary doubles as a fake pi: TestMain intercepts the run when the
// harness sets FAKE_PI, so RPC tests need no shell script and work on every
// platform.
func TestMain(m *testing.M) {
	if os.Getenv("FAKE_PI") != "" {
		fakePi()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func fakePi() {
	extra := os.Getenv("FAKE_PI_EXTRA")
	enc := json.NewEncoder(os.Stdout)

	switch extra {
	case "unicode":
		// Written raw: an encoder would escape the separators and the test
		// would prove nothing.
		_, _ = fmt.Fprintf(os.Stdout,
			"{\"type\":\"message_update\",\"assistantMessageEvent\":{\"type\":\"text_delta\",\"contentIndex\":0,\"delta\":\"line%sseparator%shere\"}}\n",
			lineSep, paraSep)
		_, _ = fmt.Fprint(os.Stdout, "{\"type\":\"agent_start\"}\r\n")
	case "badline":
		_, _ = fmt.Fprintln(os.Stdout, "not json at all")
	case "huge":
		_ = enc.Encode(map[string]any{
			"type": "tool_execution_update", "toolCallId": "t1",
			"partialResult": map[string]any{"text": strings.Repeat("x", 300<<10)},
		})
	}

	in := newLineReader(os.Stdin)
	for {
		line, ok := in.next()
		if !ok {
			// Orderly shutdown: pi exits when stdin closes, after draining.
			_, _ = fmt.Fprintln(os.Stderr, "eof")
			return
		}
		var cmd struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal(line, &cmd); err != nil {
			continue
		}
		// A real pi paints extension UI before it answers anything.
		_ = enc.Encode(map[string]any{
			"type": "extension_ui_request", "id": "ui-1", "method": "setStatus",
			"statusKey": "demo", "statusText": "\u001b[31mred\u001b[0m",
		})
		_ = enc.Encode(map[string]any{"type": "agent_start"})

		switch cmd.Type {
		case "get_state":
			_ = enc.Encode(map[string]any{
				"id": cmd.ID, "type": "response", "command": "get_state", "success": true,
				"data": map[string]any{"sessionId": "s1", "isStreaming": false},
			})
		case "set_model":
			_ = enc.Encode(map[string]any{
				"id": cmd.ID, "type": "response", "command": "set_model", "success": false,
				"error": "Model not found: invalid/model",
			})
		case "get_messages":
			// Never answer: the caller's context or the process exit must end
			// the wait.
		default:
			_ = enc.Encode(map[string]any{
				"id": cmd.ID, "type": "response", "command": cmd.Type, "success": true,
			})
		}
	}
}

// lineReader splits only on LF, the same rule the client must follow.
type lineReader struct {
	r   *os.File
	buf []byte
}

func newLineReader(f *os.File) *lineReader { return &lineReader{r: f} }

func (l *lineReader) next() ([]byte, bool) {
	one := make([]byte, 4096)
	for {
		if i := strings.IndexByte(string(l.buf), '\n'); i >= 0 {
			line := l.buf[:i]
			l.buf = l.buf[i+1:]
			return trimCR(line), true
		}
		n, err := l.r.Read(one)
		l.buf = append(l.buf, one[:n]...)
		if err != nil {
			if len(l.buf) > 0 {
				line := l.buf
				l.buf = nil
				return trimCR(line), true
			}
			return nil, false
		}
	}
}

func trimCR(b []byte) []byte {
	if n := len(b); n > 0 && b[n-1] == '\r' {
		return b[:n-1]
	}
	return b
}

func startFake(t *testing.T, extra string) (*Client, *records) {
	t.Helper()
	t.Setenv("FAKE_PI", "1")
	if extra != "" {
		t.Setenv("FAKE_PI_EXTRA", extra)
	}
	rec := &records{}
	c, err := Start(context.Background(), Options{
		Executable: os.Args[0],
		Dir:        t.TempDir(),
		OnRecord:   rec.add,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = c.Close(2 * time.Second) })
	return c, rec
}

type records struct {
	all []Record
}

func (r *records) add(rec Record) { r.all = append(r.all, rec) }

func (r *records) count(kind RecordKind) int {
	n := 0
	for _, rec := range r.all {
		if rec.Kind == kind {
			n++
		}
	}
	return n
}

func TestCallCorrelatesByIDWhileEventsArrive(t *testing.T) {
	c, rec := startFake(t, "")

	var state struct {
		SessionID string `json:"sessionId"`
	}
	if err := c.Call(testContext(t), map[string]any{"type": "get_state"}, &state); err != nil {
		t.Fatalf("Call: %v", err)
	}
	if state.SessionID != "s1" {
		t.Fatalf("sessionId = %q, want s1", state.SessionID)
	}

	// Extension UI and events arrived interleaved and were not mistaken for the
	// response.
	if n := rec.count(KindExtensionUI); n == 0 {
		t.Fatal("extension_ui_request was not classified")
	}
	if n := rec.count(KindEvent); n == 0 {
		t.Fatal("session events were not delivered")
	}
}

func TestFailedResponseBecomesCommandError(t *testing.T) {
	c, _ := startFake(t, "")
	err := c.Call(testContext(t), map[string]any{"type": "set_model"}, nil)
	var cmdErr *CommandError
	if !errors.As(err, &cmdErr) {
		t.Fatalf("Call = %v, want *CommandError", err)
	}
	if !strings.Contains(cmdErr.Message, "Model not found") {
		t.Fatalf("message = %q", cmdErr.Message)
	}
}

func TestUnicodeSeparatorsDoNotSplitRecords(t *testing.T) {
	c, rec := startFake(t, "unicode")
	if err := c.Call(testContext(t), map[string]any{"type": "get_state"}, nil); err != nil {
		t.Fatalf("Call: %v", err)
	}
	var delta string
	var starts int
	for _, r := range rec.all {
		switch r.Type {
		case "message_update":
			var ev struct {
				AssistantMessageEvent struct {
					Delta string `json:"delta"`
				} `json:"assistantMessageEvent"`
			}
			if err := json.Unmarshal(r.Raw, &ev); err != nil {
				t.Fatalf("event did not parse: %v", err)
			}
			delta = ev.AssistantMessageEvent.Delta
		case "agent_start":
			starts++
		}
	}
	want := "line" + lineSep + "separator" + paraSep + "here"
	if delta != want {
		t.Fatalf("delta = %q, want the raw separators intact", delta)
	}
	// One CRLF record written at startup plus the one the echo loop emits for
	// the command: two records. A reader that split on U+2028 would report more
	// and lose the delta above.
	if starts != 2 {
		t.Fatalf("agent_start records = %d, want 2", starts)
	}
}

func TestHugeRecordIsRead(t *testing.T) {
	c, rec := startFake(t, "huge")
	if err := c.Call(testContext(t), map[string]any{"type": "get_state"}, nil); err != nil {
		t.Fatalf("Call: %v", err)
	}
	var biggest int
	for _, r := range rec.all {
		if len(r.Raw) > biggest {
			biggest = len(r.Raw)
		}
	}
	if biggest < 300<<10 {
		t.Fatalf("biggest record = %d bytes, want > 300 KiB", biggest)
	}
}

func TestBadLineDoesNotBreakTheStream(t *testing.T) {
	c, rec := startFake(t, "badline")
	if err := c.Call(testContext(t), map[string]any{"type": "get_state"}, nil); err != nil {
		t.Fatalf("Call after a bad line: %v", err)
	}
	if len(rec.all) == 0 {
		t.Fatal("no records after a bad line")
	}
	if !strings.Contains(strings.Join(c.StderrLines(), "\n"), "unparseable") {
		t.Fatal("the bad line was not recorded for diagnostics")
	}
}

func TestCallEndsWhenProcessExits(t *testing.T) {
	c, _ := startFake(t, "")
	// get_messages is never answered; kill the process and expect a prompt
	// error instead of a hang.
	go func() {
		time.Sleep(200 * time.Millisecond)
		_ = c.Close(50 * time.Millisecond)
	}()
	done := make(chan error, 1)
	go func() { done <- c.Call(testContext(t), map[string]any{"type": "get_messages"}, nil) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Call succeeded against a dead process")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Call did not return after the process exited")
	}
}

func TestExitIsReportedAndStderrKept(t *testing.T) {
	c, _ := startFake(t, "")
	if err := c.Close(2 * time.Second); err != nil {
		t.Fatalf("Close: %v", err)
	}
	select {
	case <-c.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Done was not closed")
	}
	lines := strings.Join(c.StderrLines(), "\n")
	if !strings.Contains(lines, "eof") {
		t.Fatalf("stderr = %q, want the fake's eof line", lines)
	}
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}
