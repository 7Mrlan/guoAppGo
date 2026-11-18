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

// TestM2Nav checks that the navigation lists only user pages, while the
// component page stays reachable by its own hash.
func TestM2Nav(t *testing.T) {
	w := appWindow(t)

	evalString(t, w, `location.hash = '#/', 'ok'`)
	if !waitFor(t, w, `document.querySelector('[data-route=home]') !== null`) {
		t.Fatal("the navigation did not render")
	}
	if got := evalString(t, w, `String(document.querySelectorAll('[data-route=components]').length)`); got != "0" {
		t.Fatalf("navigation entries for the component page = %s, want 0", got)
	}

	evalString(t, w, `location.hash = '#/dev/components', 'ok'`)
	if !waitFor(t, w, `document.querySelectorAll('[data-ac]').length === 12`) {
		t.Fatalf("tagged controls reached by a direct hash = %s, want 12",
			evalString(t, w, `String(document.querySelectorAll('[data-ac]').length)`))
	}
}

// TestM1Components checks the component page: twelve tagged controls, exactly
// one disabled, every control carries a name, and the layer closes on Esc with
// the focus returned to the control that opened it.
func TestM1Components(t *testing.T) {
	w := appWindow(t)

	evalString(t, w, `location.hash = '#/dev/components', 'ok'`)
	if !waitFor(t, w, `document.querySelectorAll('[data-ac]').length === 12`) {
		t.Fatalf("tagged controls = %s, want 12",
			evalString(t, w, `String(document.querySelectorAll('[data-ac]').length)`))
	}

	unnamed := `[...document.querySelectorAll('[data-ac]')].filter(e => !e.getAttribute('aria-label') && !e.textContent.trim()).length`
	if got := evalString(t, w, `String(`+unnamed+`)`); got != "0" {
		t.Fatalf("controls without an accessible name = %s, want 0", got)
	}
	if got := evalString(t, w, `String(document.querySelectorAll('[data-ac][disabled]').length)`); got != "1" {
		t.Fatalf("disabled controls = %s, want 1", got)
	}
	if got := evalString(t, w, `String(document.querySelectorAll('[data-ac]:not([disabled])').length)`); got != "11" {
		t.Fatalf("enabled controls = %s, want 11", got)
	}

	evalString(t, w, `(() => { const el = document.querySelector('[data-ac="layer-trigger"]'); el.focus(); el.click(); return 'ok'; })()`)
	if !waitFor(t, w, `(() => { const l = document.querySelector('[data-role=layer]'); return !!l && l.contains(document.activeElement); })()`) {
		t.Fatalf("the layer did not open with the focus inside it")
	}

	// Tab 在弹层内回绕，否则 aria-modal 就是对辅助技术的空承诺。
	const focusables = `[...document.querySelector('[data-role=layer]').querySelectorAll('a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])')]`
	forward := `(() => { const items = ` + focusables + `;
		items[items.length - 1].focus();
		document.querySelector('[data-role=layer]').dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }));
		return items[0] === document.activeElement ? 'wrapped' : 'stuck'; })()`
	if got := evalString(t, w, forward); got != "wrapped" {
		t.Fatalf("Tab from the last control in the layer = %s, want it wrapped to the first", got)
	}
	backward := `(() => { const items = ` + focusables + `;
		items[0].focus();
		document.querySelector('[data-role=layer]').dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true }));
		return items[items.length - 1] === document.activeElement ? 'wrapped' : 'stuck'; })()`
	if got := evalString(t, w, backward); got != "wrapped" {
		t.Fatalf("Shift+Tab from the first control in the layer = %s, want it wrapped to the last", got)
	}

	evalString(t, w, `document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true })), 'ok'`)
	if !waitFor(t, w, `document.querySelector('[data-role=layer]') === null`) {
		t.Fatalf("the layer is still open after Esc")
	}
	if got := evalString(t, w, `document.activeElement && document.activeElement.dataset.ac`); got != "layer-trigger" {
		t.Fatalf("focus after closing the layer = %q, want layer-trigger", got)
	}
}

// TestM1Routes checks the hash router: a nav click moves the hash and the
// current-page marker, a reload keeps both, an unknown path renders the
// not-found page, and an empty hash falls back to the home route.
func TestM1Routes(t *testing.T) {
	w := appWindow(t)

	if got := evalString(t, w, `location.hash = '', 'ok'`); got != "ok" {
		t.Fatalf("clearing the hash: %q", got)
	}
	if !waitFor(t, w, `document.querySelector('[data-route=home]') !== null`) {
		t.Fatal("the navigation did not render")
	}

	evalString(t, w, `document.querySelector('[data-route=explore]').click(), 'ok'`)
	if !waitFor(t, w, `location.hash === '#/explore'`) {
		t.Fatalf("hash after clicking 发现 = %q, want #/explore", evalString(t, w, `location.hash`))
	}
	if got := evalString(t, w, `document.querySelector('[aria-current=page]').dataset.route`); got != "explore" {
		t.Fatalf("current nav item = %q, want explore", got)
	}

	w.Page().Reload()
	if !waitFor(t, w, `document.readyState === 'complete' && document.title === 'guoAppGo'`) {
		t.Fatal("the page did not reload")
	}
	if got := evalString(t, w, `location.hash`); got != "#/explore" {
		t.Fatalf("hash after a reload = %q, want #/explore", got)
	}
	if got := evalString(t, w, `document.querySelector('[aria-current=page]').dataset.route`); got != "explore" {
		t.Fatalf("current nav item after a reload = %q, want explore", got)
	}

	evalString(t, w, `location.hash = '#/nope', 'ok'`)
	if !waitFor(t, w, `!!document.querySelector('[data-role=not-found]')`) {
		t.Fatal("an unknown hash did not render the not-found page")
	}
	if got := evalString(t, w, `String(document.body.children.length)`); got == "0" {
		t.Fatal("the body is empty on an unknown hash")
	}

	evalString(t, w, `location.hash = '', 'ok'`)
	if !waitFor(t, w, `(() => { const h = location.hash; return (h === '' || h === '#/') && !!document.querySelector('[data-route=home]'); })()`) {
		t.Fatalf("an empty hash did not fall back to the home route, hash = %q", evalString(t, w, `location.hash`))
	}
}
