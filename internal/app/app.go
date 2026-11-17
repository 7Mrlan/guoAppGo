// Package app wires the window, the embedded frontend and the services the
// page calls.
package app

import (
	"fmt"

	"github.com/egoist/mygo"

	"github.com/7Mrlan/guoAppGo/web"
)

// Version is the application version, reported to the page by System.
const Version = "0.0.1"

// System reports build information to the page.
type System struct {
	version string
}

// Version returns the application version.
func (s System) Version() string {
	return s.version
}

// Window returns the options of the app's main window.
func Window() mygo.WindowOptions {
	return mygo.WindowOptions{
		Title:           "guoAppGo",
		URL:             "/",
		Width:           1080,
		Height:          720,
		MinWidth:        900,
		MinHeight:       600,
		StateKey:        "main",
		BackgroundColor: "#101014",
	}
}

// Run starts the application and blocks until it quits.
func Run() error {
	fsys, err := web.FS()
	if err != nil {
		return fmt.Errorf("load the frontend: %w", err)
	}
	mygo.SetFrontend(fsys)
	mygo.Bind(System{version: Version})

	mygo.App.WhenReady(func() {
		mygo.NewWindow(Window())
	})
	return mygo.App.Run()
}
