package web

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/egoist/mygo"
)

// TestMain runs the app loop the page tests need. Without MYGO_E2E the E2E
// tests skip themselves and the plain Go tests run as usual.
func TestMain(m *testing.M) {
	if os.Getenv("MYGO_E2E") == "" {
		os.Exit(m.Run())
	}
	fsys, err := FS()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	mygo.SetFrontend(fsys)
	code := 1
	mygo.App.WhenReady(func() {
		go func() {
			code = m.Run()
			if sharedWin != nil {
				sharedWin.Destroy()
			}
			mygo.App.Quit()
		}()
	})
	if err := mygo.App.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(code)
}

// sharedWin is opened once: this WebView2 runtime refuses a second controller
// in the same process, and both page tests can share one window.
var (
	sharedOnce sync.Once
	sharedWin  *mygo.Window
	sharedErr  error
)

// blockValues parses the declarations of one selector block of tokens.css.
func blockValues(t *testing.T, css, selector string) map[string]string {
	t.Helper()
	start := strings.Index(css, selector+" {")
	if start == -1 {
		t.Fatalf("tokens.css has no %s block", selector)
	}
	end := strings.Index(css[start:], "}")
	if end == -1 {
		t.Fatalf("tokens.css has an unterminated %s block", selector)
	}
	values := map[string]string{}
	for _, line := range strings.Split(css[start:start+end], "\n") {
		match := regexp.MustCompile(`^\s*(--[a-z0-9-]+)\s*:\s*([^;]+);`).FindStringSubmatch(line)
		if match != nil {
			values[match[1]] = strings.TrimSpace(match[2])
		}
	}
	return values
}

// tokenFor returns a token as the given theme defines it.
func tokenFor(t *testing.T, theme, name string) string {
	t.Helper()
	css, err := os.ReadFile("src/tokens.css")
	if err != nil {
		t.Fatalf("read tokens.css: %v", err)
	}
	base := blockValues(t, string(css), ":root")
	if theme == "light" {
		if overridden, ok := blockValues(t, string(css), `[data-theme="light"]`)[name]; ok {
			return overridden
		}
	}
	value, ok := base[name]
	if !ok {
		t.Fatalf("token %s not found in tokens.css", name)
	}
	return value
}

// appWindow returns the shared window, opening the app page on first use.
func appWindow(t *testing.T) *mygo.Window {
	t.Helper()
	if os.Getenv("MYGO_E2E") == "" {
		t.Skip("set MYGO_E2E=1 to run page tests in a desktop session")
	}
	sharedOnce.Do(func() {
		w := mygo.NewWindow(mygo.WindowOptions{Title: "e2e", URL: "/", Width: 800, Height: 600})
		deadline := time.Now().Add(25 * time.Second)
		for time.Now().Before(deadline) {
			v, err := w.Page().Eval(`document.title === 'guoAppGo' && document.readyState === 'complete'`)
			if err != nil {
				sharedErr = err
			} else if v == true {
				sharedWin = w
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		w.Destroy()
		if sharedErr == nil {
			sharedErr = fmt.Errorf("the page did not load within 25s")
		}
	})
	if sharedWin == nil {
		t.Skipf("no usable webview session: %v", sharedErr)
	}
	return sharedWin
}

// evalString evaluates an expression that must return a string.
func evalString(t *testing.T, w *mygo.Window, expr string) string {
	t.Helper()
	v, err := w.Page().Eval(expr)
	if err != nil {
		t.Fatalf("Eval(%s): %v", expr, err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("Eval(%s) = %v (%T), want a string", expr, v, v)
	}
	return s
}

// rgbString renders a hex colour the way getComputedStyle reports it.
func rgbString(t *testing.T, hex string) string {
	t.Helper()
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		t.Fatalf("unsupported colour %q", hex)
	}
	var r, g, b int
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		t.Fatalf("parse %q: %v", hex, err)
	}
	return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
}

// waitFor polls expr until it is true.
func waitFor(t *testing.T, w *mygo.Window, expr string) bool {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		v, err := w.Page().Eval(expr)
		if err == nil && v == true {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// TestM1Tokens checks that the page resolves --color-bg to the value tokens.css
// declares for the active theme, and that the active theme is one of the two.
func TestM1Tokens(t *testing.T) {
	w := appWindow(t)

	theme := evalString(t, w, `document.documentElement.dataset.theme`)
	if theme != "dark" && theme != "light" {
		t.Fatalf("dataset.theme = %q, want dark or light", theme)
	}

	wantBg := tokenFor(t, theme, "--color-bg")
	gotBg := strings.TrimSpace(evalString(t, w, `getComputedStyle(document.documentElement).getPropertyValue('--color-bg')`))
	if !strings.EqualFold(gotBg, wantBg) {
		t.Fatalf("--color-bg = %q, want %q for the %s theme (tokens.css)", gotBg, wantBg, theme)
	}
}

// TestM1Theme checks that with no override the page follows the system
// preference, that a valid override wins and survives a reload, and that an
// invalid one falls back to the system preference.
func TestM1Theme(t *testing.T) {
	w := appWindow(t)

	systemTheme := evalString(t, w, `matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'`)
	evalString(t, w, `localStorage.removeItem('theme'), 'ok'`)
	w.Page().Reload()
	if !waitFor(t, w, `document.readyState === 'complete' && document.title === 'guoAppGo'`) {
		t.Fatal("the page did not reload")
	}
	if got := evalString(t, w, `document.documentElement.dataset.theme`); got != systemTheme {
		t.Fatalf("theme without an override = %q, want the system preference %q", got, systemTheme)
	}

	evalString(t, w, `localStorage.setItem('theme','dark'), 'ok'`)
	w.Page().Reload()
	if !waitFor(t, w, `document.readyState === 'complete' && document.title === 'guoAppGo'`) {
		t.Fatal("the page did not reload")
	}
	if got := evalString(t, w, `document.documentElement.dataset.theme`); got != "dark" {
		t.Fatalf("theme after an override = %q, want dark", got)
	}
	wantBg := rgbString(t, tokenFor(t, "dark", "--color-bg"))
	if got := evalString(t, w, `getComputedStyle(document.body).backgroundColor`); got != wantBg {
		t.Fatalf("body background = %q, want the dark token %q", got, wantBg)
	}

	evalString(t, w, `localStorage.setItem('theme','blue'), 'ok'`)
	w.Page().Reload()
	if !waitFor(t, w, `document.readyState === 'complete' && document.title === 'guoAppGo'`) {
		t.Fatal("the page did not reload")
	}
	theme := evalString(t, w, `document.documentElement.dataset.theme`)
	if theme != systemTheme {
		t.Fatalf("theme with an invalid override = %q, want the system preference %q", theme, systemTheme)
	}
	if got := evalString(t, w, `getComputedStyle(document.body).backgroundColor`); got == "transparent" || got == "rgba(0, 0, 0, 0)" {
		t.Fatalf("body background = %q, want a themed colour", got)
	}
}
