package piagent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// RecordKind distinguishes what arrived on stdout.
type RecordKind int

const (
	// KindResponse answers a command.
	KindResponse RecordKind = iota
	// KindEvent is a session event such as message_update or agent_settled.
	KindEvent
	// KindExtensionUI is an extension's request for the host's UI.
	KindExtensionUI
)

// Record is one line of pi's stdout, classified.
type Record struct {
	Kind    RecordKind
	Type    string          // the wire "type"
	ID      string          // response id, or extension UI request id
	Command string          // for responses: the command answered
	Success bool            // for responses
	Error   string          // for responses
	Data    json.RawMessage // for responses
	Raw     json.RawMessage // the whole record, for events and extension UI
}

// Options configures a pi RPC process.
type Options struct {
	// Executable is the resolved pi path.
	Executable string
	// Args are extra arguments after --mode rpc, such as --session.
	Args []string
	// Dir is the working directory: the project.
	Dir string
	// Env adds to the inherited environment.
	Env []string
	// OnRecord receives every record in arrival order. It must not block for
	// long: pi honors stdout backpressure, so a slow handler stalls the agent.
	OnRecord func(Record)
	// OnStderr receives diagnostics lines.
	OnStderr func(line string)
	// OnExit is called once when the process ends, with its error.
	OnExit func(err error)
}

// Client is one pi RPC process.
type Client struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stderr *ring

	writeMu sync.Mutex

	pendingMu sync.Mutex
	pending   map[string]chan Record
	nextID    int64

	done     chan struct{}
	exitErr  error
	exitOnce sync.Once

	opts Options
}

// ring keeps the last n stderr lines for diagnostics.
type ring struct {
	mu    sync.Mutex
	lines []string
	max   int
}

func newRing(max int) *ring { return &ring{max: max} }

func (r *ring) add(line string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lines = append(r.lines, line)
	if len(r.lines) > r.max {
		r.lines = r.lines[len(r.lines)-r.max:]
	}
}

// Lines returns the kept lines, oldest first.
func (r *ring) Lines() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.lines...)
}

// Start launches pi in RPC mode and begins reading.
func Start(ctx context.Context, opts Options) (*Client, error) {
	if opts.Executable == "" {
		return nil, errors.New("piagent: no executable")
	}
	args := append([]string{"--mode", "rpc"}, opts.Args...)
	// The executable comes from Resolve, which stats it and requires the
	// executable bit; no user text is ever concatenated into this command.
	cmd := exec.Command(opts.Executable, args...)
	cmd.Dir = opts.Dir
	if len(opts.Env) > 0 {
		cmd.Env = append(cmd.Environ(), opts.Env...)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("piagent: stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("piagent: stdout: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("piagent: stderr: %w", err)
	}

	c := &Client{
		cmd:     cmd,
		stdin:   stdin,
		stderr:  newRing(40),
		pending: make(map[string]chan Record),
		done:    make(chan struct{}),
		opts:    opts,
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("piagent: start %s: %w", opts.Executable, err)
	}

	go c.readStdout(stdout)
	go c.readStderr(stderrPipe)
	go c.wait()

	// A cancelled context means the caller went away: close the process.
	go func() {
		<-ctx.Done()
		c.terminate()
	}()

	return c, nil
}

func (c *Client) readStdout(r io.Reader) {
	br := bufio.NewReaderSize(r, 256<<10)
	for {
		line, err := readLine(br)
		if len(line) > 0 {
			if rec, ok := c.parse(line); ok {
				c.dispatch(rec)
			}
		}
		if err != nil {
			return
		}
	}
}

// readLine reads one LF-terminated record. It never splits on U+2028 or
// U+2029, which are legal inside JSON strings, and it has no size limit.
func readLine(br *bufio.Reader) ([]byte, error) {
	var buf []byte
	for {
		chunk, err := br.ReadSlice('\n')
		buf = append(buf, chunk...)
		if err == nil {
			buf = buf[:len(buf)-1]
			if n := len(buf); n > 0 && buf[n-1] == '\r' {
				buf = buf[:n-1]
			}
			return buf, nil
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		return buf, err
	}
}

func (c *Client) parse(line []byte) (Record, bool) {
	var head struct {
		Type    string `json:"type"`
		ID      string `json:"id"`
		Command string `json:"command"`
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(line, &head); err != nil {
		// pi never writes non-JSON to stdout; a line that does not parse is a
		// protocol violation worth keeping for diagnostics.
		c.stderr.add("unparseable stdout: " + truncate(string(line), 300))
		return Record{}, false
	}
	raw := append(json.RawMessage(nil), line...)
	rec := Record{Type: head.Type, ID: head.ID, Command: head.Command, Success: head.Success, Error: head.Error, Raw: raw}
	switch head.Type {
	case "response":
		var withData struct {
			Data json.RawMessage `json:"data"`
		}
		_ = json.Unmarshal(line, &withData)
		rec.Data = withData.Data
		rec.Kind = KindResponse
	case "extension_ui_request":
		rec.Kind = KindExtensionUI
	default:
		rec.Kind = KindEvent
	}
	return rec, true
}

func (c *Client) dispatch(rec Record) {
	if rec.Kind == KindResponse {
		c.pendingMu.Lock()
		ch, ok := c.pending[rec.ID]
		if ok {
			delete(c.pending, rec.ID)
		}
		c.pendingMu.Unlock()
		if ok {
			ch <- rec // buffered, size 1
			return
		}
		// A response nobody waits for: pi rejects malformed records this way.
		if rec.Command == "parse" && !rec.Success {
			c.stderr.add("protocol: " + rec.Error)
			return
		}
	}
	if c.opts.OnRecord != nil {
		c.opts.OnRecord(rec)
	}
}

func (c *Client) readStderr(r io.Reader) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		c.stderr.add(line)
		if c.opts.OnStderr != nil {
			c.opts.OnStderr(line)
		}
	}
}

func (c *Client) wait() {
	err := c.cmd.Wait()
	c.exitOnce.Do(func() {
		c.exitErr = err
		close(c.done)
		c.failPending(err)
		if c.opts.OnExit != nil {
			c.opts.OnExit(err)
		}
	})
}

func (c *Client) failPending(err error) {
	c.pendingMu.Lock()
	pending := c.pending
	c.pending = make(map[string]chan Record)
	c.pendingMu.Unlock()
	for id, ch := range pending {
		ch <- Record{Kind: KindResponse, ID: id, Error: "process exited: " + errText(err)}
	}
}

// Send writes one command. The call does not wait for a response.
func (c *Client) Send(cmd any) (string, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.pendingMu.Lock()
	c.nextID++
	id := fmt.Sprintf("req-%d", c.nextID)
	c.pendingMu.Unlock()

	m, ok := cmd.(map[string]any)
	if !ok {
		b, err := json.Marshal(cmd)
		if err != nil {
			return "", fmt.Errorf("piagent: marshal command: %w", err)
		}
		if err := json.Unmarshal(b, &m); err != nil {
			return "", fmt.Errorf("piagent: command is not an object: %w", err)
		}
	}
	m["id"] = id
	line, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("piagent: marshal command: %w", err)
	}
	line = append(line, '\n')
	if _, err := c.stdin.Write(line); err != nil {
		return id, fmt.Errorf("piagent: write command: %w", err)
	}
	return id, nil
}

// Call sends a command and waits for its response, decoding data into out when
// out is not nil. A failed response becomes an error carrying pi's message.
func (c *Client) Call(ctx context.Context, cmd any, out any) error {
	ch := make(chan Record, 1)

	c.writeMu.Lock()
	c.pendingMu.Lock()
	c.nextID++
	id := fmt.Sprintf("req-%d", c.nextID)
	c.pending[id] = ch
	c.pendingMu.Unlock()

	m, ok := cmd.(map[string]any)
	if !ok {
		b, err := json.Marshal(cmd)
		if err != nil {
			c.forget(id)
			c.writeMu.Unlock()
			return fmt.Errorf("piagent: marshal command: %w", err)
		}
		if err := json.Unmarshal(b, &m); err != nil {
			c.forget(id)
			c.writeMu.Unlock()
			return fmt.Errorf("piagent: command is not an object: %w", err)
		}
	}
	m["id"] = id
	line, err := json.Marshal(m)
	if err != nil {
		c.forget(id)
		c.writeMu.Unlock()
		return fmt.Errorf("piagent: marshal command: %w", err)
	}
	_, err = c.stdin.Write(append(line, '\n'))
	c.writeMu.Unlock()
	if err != nil {
		c.forget(id)
		return fmt.Errorf("piagent: write command: %w", err)
	}

	select {
	case rec := <-ch:
		if !rec.Success {
			msg := rec.Error
			if msg == "" {
				msg = "command failed"
			}
			return &CommandError{Command: rec.Command, Message: msg}
		}
		if out != nil && len(rec.Data) > 0 {
			if err := json.Unmarshal(rec.Data, out); err != nil {
				return fmt.Errorf("piagent: decode %s response: %w", rec.Command, err)
			}
		}
		return nil
	case <-c.done:
		c.forget(id)
		return fmt.Errorf("piagent: %s: process exited: %w", commandType(m), c.exitErr)
	case <-ctx.Done():
		c.forget(id)
		return ctx.Err()
	}
}

func (c *Client) forget(id string) {
	c.pendingMu.Lock()
	delete(c.pending, id)
	c.pendingMu.Unlock()
}

// CommandError is a command pi answered with success:false.
type CommandError struct {
	Command string
	Message string
}

func (e *CommandError) Error() string { return e.Command + ": " + e.Message }

// StderrLines returns the recent diagnostics lines.
func (c *Client) StderrLines() []string { return c.stderr.Lines() }

// Pid reports the process id, or 0 when not started.
func (c *Client) Pid() int {
	if c.cmd.Process == nil {
		return 0
	}
	return c.cmd.Process.Pid
}

// Done is closed when the process has exited.
func (c *Client) Done() <-chan struct{} { return c.done }

// ExitErr reports why the process ended, nil when it has not or exited cleanly.
func (c *Client) ExitErr() error {
	select {
	case <-c.done:
		return c.exitErr
	default:
		return nil
	}
}

// Close shuts pi down in the order it requires: close stdin, keep draining
// stdout until the process exits, and kill only if it does not comply.
//
// Closing the read end first crashes pi with an unhandled EPIPE, so this order
// is not optional.
func (c *Client) Close(grace time.Duration) error {
	_ = c.stdin.Close()

	select {
	case <-c.done:
		return nil
	case <-time.After(grace):
	}

	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	select {
	case <-c.done:
		return nil
	case <-time.After(2 * time.Second):
		return errors.New("piagent: process did not exit after kill")
	}
}

func (c *Client) terminate() {
	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
}

func commandType(m map[string]any) string {
	if s, ok := m["type"].(string); ok {
		return s
	}
	return "command"
}

func errText(err error) string {
	if err == nil {
		return "exit status 0"
	}
	return err.Error()
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
