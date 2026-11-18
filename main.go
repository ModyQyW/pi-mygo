package main

import (
	"context"
	"errors"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/egoist/mygo"

	// The app's catalogs, which make its languages the app's.
	"github.com/ModyQyW/pi-mygo/internal/piagent"
	_ "github.com/ModyQyW/pi-mygo/locales"
)

// Version is replaced by `mygo build` with the version in mygo.config.ts.
const Version = "0.1.0"

// ProbeState is the outcome of looking for a usable pi.
type ProbeState string

const (
	ProbeChecking    ProbeState = "checking"
	ProbeReady       ProbeState = "ready"
	ProbeMissing     ProbeState = "missing"
	ProbeUnsupported ProbeState = "unsupported"
	ProbeError       ProbeState = "error"
)

// Probe reports what the app found: the pi binary, its version, and whether
// this app supports that version.
type Probe struct {
	State     ProbeState `json:"state"`
	Path      string     `json:"path"`
	Version   string     `json:"version"`
	Compat    string     `json:"compat"`
	Detail    string     `json:"detail"`
	CheckedAt time.Time  `json:"checkedAt"`
}

// Info is everything the window needs to render its environment panel.
type Info struct {
	AppVersion string `json:"appVersion"`
	Platform   string `json:"platform"`
	Arch       string `json:"arch"`
	Pi         Probe  `json:"pi"`
	MinPi      string `json:"minPi"`
	TestedPi   string `json:"testedPi"`
}

// App holds the state the frontend may read. Every exported method is callable
// from the page through the generated client.
type App struct {
	mu    sync.Mutex
	probe Probe
	win   *mygo.Window
}

// EnvironmentChanged is emitted when a probe finishes.
var EnvironmentChanged = mygo.NewEvent[Probe]("environmentChanged")

// Info returns the current environment report. The first call starts a probe in
// the background; the page receives EnvironmentChanged when it finishes.
func (a *App) Info() Info {
	a.mu.Lock()
	probe := a.probe
	needProbe := probe.State == ""
	a.mu.Unlock()

	if needProbe {
		a.mu.Lock()
		a.probe = Probe{State: ProbeChecking, CheckedAt: time.Now()}
		a.mu.Unlock()
		go a.probeNow(context.Background())
	}

	return Info{
		AppVersion: Version,
		Platform:   runtime.GOOS,
		Arch:       runtime.GOARCH,
		Pi:         probe,
		MinPi:      piagent.Min.String(),
		TestedPi:   piagent.Tested.String(),
	}
}

// Recheck probes for pi again, and returns the result.
func (a *App) Recheck(ctx context.Context) (Probe, error) {
	a.mu.Lock()
	a.probe = Probe{State: ProbeChecking, CheckedAt: time.Now()}
	a.mu.Unlock()
	return a.probeNow(ctx), nil
}

// probeNow resolves and inspects pi, stores the result, and tells the page.
func (a *App) probeNow(ctx context.Context) Probe {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	probe := inspect(ctx)

	a.mu.Lock()
	a.probe = probe
	win := a.win
	a.mu.Unlock()

	if win != nil {
		// A closed window ends the loop by returning an error; there is
		// nothing left to tell.
		_ = EnvironmentChanged.Emit(win, probe)
	}
	return probe
}

func inspect(ctx context.Context) Probe {
	now := time.Now()
	path, err := piagent.Shell().Resolve()
	if err != nil {
		if errors.Is(err, piagent.ErrNotFound) {
			return Probe{
				State:     ProbeMissing,
				Detail:    "No pi executable found in PI_EXECUTABLE, PATH, or the login shell.",
				CheckedAt: now,
			}
		}
		return Probe{State: ProbeError, Detail: err.Error(), CheckedAt: now}
	}

	version, err := piagent.ReadVersion(ctx, path)
	if err != nil {
		return Probe{State: ProbeError, Path: path, Detail: err.Error(), CheckedAt: now}
	}

	compat := version.Compatible()
	probe := Probe{
		Path:      path,
		Version:   version.String(),
		Compat:    string(compat),
		CheckedAt: now,
	}
	switch compat {
	case piagent.CompatOK:
		probe.State = ProbeReady
	case piagent.CompatUntested:
		probe.State = ProbeReady
		probe.Detail = "Newer than the version this app was tested with."
	case piagent.CompatTooOld:
		probe.State = ProbeUnsupported
		probe.Detail = "This app needs pi " + piagent.Min.String() + " or newer."
	case piagent.CompatTooNew:
		probe.State = ProbeUnsupported
		probe.Detail = "This app supports pi " + piagent.Min.Series() + ". A newer app is needed."
	}
	return probe
}

func main() {
	app := &App{}
	mygo.Bind(app)

	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:     "Pi MyGo",
			Width:     1240,
			Height:    840,
			MinWidth:  960,
			MinHeight: 600,
			// The page draws its own title bar under the traffic lights, which
			// sit 18 pt in from the top-left corner.
			TitleBarStyle:        mygo.TitleBarHidden,
			TrafficLightPosition: &mygo.Point{X: 18, Y: 18},
			BackgroundColor:      "light-dark(#faf7f2, #14120f)",
			StateKey:             "main",
			URL:                  "/",
		})
		app.mu.Lock()
		app.win = win
		app.mu.Unlock()

		mygo.App.SetMenu(buildMenu())
		go app.probeNow(context.Background())
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// buildMenu is the menu bar. Items that only a later phase can act on are not
// here yet: a menu item that does nothing is worse than a missing one.
func buildMenu() *mygo.Menu {
	appMenu := &mygo.MenuItem{Role: mygo.RoleAppMenu, Submenu: []*mygo.MenuItem{
		{Role: mygo.RoleAbout},
		mygo.Separator(),
		{Role: mygo.RoleServices, Submenu: []*mygo.MenuItem{}},
		mygo.Separator(),
		{Role: mygo.RoleHide},
		{Role: mygo.RoleHideOthers},
		{Role: mygo.RoleUnhide},
		mygo.Separator(),
		{Role: mygo.RoleQuit},
	}}

	viewMenu := &mygo.MenuItem{Label: "View", Submenu: []*mygo.MenuItem{
		{Role: mygo.RoleReload},
		{Role: mygo.RoleToggleDevTools},
		mygo.Separator(),
		{Role: mygo.RoleToggleFullScreen},
	}}

	helpMenu := &mygo.MenuItem{Role: mygo.RoleHelp, Submenu: []*mygo.MenuItem{
		{Label: "pi Documentation", Click: func(*mygo.MenuItem, *mygo.Window) {
			_ = mygo.Shell.OpenExternal("https://docs.pi.dev/latest")
		}},
		{Label: "Report an Issue", Click: func(*mygo.MenuItem, *mygo.Window) {
			_ = mygo.Shell.OpenExternal("https://github.com/ModyQyW/pi-mygo/issues")
		}},
	}}

	return mygo.NewMenu([]*mygo.MenuItem{
		appMenu,
		{Role: mygo.RoleEditMenu},
		viewMenu,
		{Role: mygo.RoleWindowMenu},
		helpMenu,
	})
}
