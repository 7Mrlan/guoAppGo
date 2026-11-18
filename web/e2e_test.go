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

// TestM2Catalog checks the catalogue cache: a category is fetched once, a
// repeat visit is served from the cache, and a module version change refetches.
func TestM2Catalog(t *testing.T) {
	w := appWindow(t)

	evalString(t, w, `(() => {
		window.__fixtureModuleVersion = '1.0.4';
		window.__hostCalls = { catalog: 0, search: 0 };
		location.hash = '#/explore';
		return 'ok';
	})()`)
	if !waitFor(t, w, `document.querySelector('[data-role=catalog-item]') !== null`) {
		t.Fatal("the catalogue did not render")
	}
	if got := evalString(t, w, `String(window.__hostCalls.catalog)`); got != "1" {
		t.Fatalf("catalogue fetches after the first visit = %s, want 1", got)
	}

	evalString(t, w, `document.querySelector('[data-category=series]').click(), 'ok'`)
	if !waitFor(t, w, `String(window.__hostCalls.catalog) === '2'`) {
		t.Fatalf("catalogue fetches after switching category = %s, want 2",
			evalString(t, w, `String(window.__hostCalls.catalog)`))
	}

	evalString(t, w, `document.querySelector('[data-category=series]').click(), 'ok'`)
	time.Sleep(200 * time.Millisecond)
	if got := evalString(t, w, `String(window.__hostCalls.catalog)`); got != "2" {
		t.Fatalf("a repeated category visit refetched: fetches = %s, want the cached 2", got)
	}

	evalString(t, w, `(() => { window.__fixtureModuleVersion = '1.0.5'; document.querySelector('[data-category=series]').click(); return 'ok'; })()`)
	if !waitFor(t, w, `String(window.__hostCalls.catalog) === '3'`) {
		t.Fatalf("a module version change did not refetch: fetches = %s, want 3",
			evalString(t, w, `String(window.__hostCalls.catalog)`))
	}
}

// TestM2Search checks the debounce: two quick inputs fire one search, an empty
// result renders the empty state, and clearing the input returns to idle.
func TestM2Search(t *testing.T) {
	w := appWindow(t)

	evalString(t, w, `(() => { window.__hostCalls = { catalog: 0, search: 0 }; location.hash = '#/search'; return 'ok'; })()`)
	if !waitFor(t, w, `document.querySelector('[data-role=search-input]') !== null`) {
		t.Fatal("the search page did not render")
	}

	evalString(t, w, `(() => {
		const input = document.querySelector('[data-role=search-input]');
		input.value = '深';
		input.dispatchEvent(new Event('input'));
		window.setTimeout(() => { input.value = '深海'; input.dispatchEvent(new Event('input')); }, 50);
		return 'ok';
	})()`)
	// 等过防抖窗口再读计数：只等结果节点出现的话，第一次计时器触发时就会读到，
	// 无法证明第二次输入被取消。
	time.Sleep(800 * time.Millisecond)
	if got := evalString(t, w, `String(window.__hostCalls.search)`); got != "1" {
		t.Fatalf("searches after two quick inputs = %s, want 1", got)
	}
	if !waitFor(t, w, `document.querySelector('[data-role=search-results]') !== null`) {
		t.Fatal("the search results did not render")
	}

	evalString(t, w, `(() => {
		const input = document.querySelector('[data-role=search-input]');
		input.value = '不存在的片名';
		input.dispatchEvent(new Event('input'));
		return 'ok';
	})()`)
	if !waitFor(t, w, `document.querySelector('[data-role=search-empty]') !== null`) {
		t.Fatal("an empty result did not render the empty state")
	}

	evalString(t, w, `(() => {
		const input = document.querySelector('[data-role=search-input]');
		input.value = '';
		input.dispatchEvent(new Event('input'));
		return 'ok';
	})()`)
	if !waitFor(t, w, `document.querySelector('[data-role=search-idle]') !== null`) {
		t.Fatal("clearing the input did not return to the idle state")
	}
}

// TestM2Detail checks the detail page: the description unclamps on demand, the
// last-watched episode is marked, and that mark survives a season switch.
func TestM2Detail(t *testing.T) {
	w := appWindow(t)

	evalString(t, w, `location.hash = '#/detail/film-001', 'ok'`)
	if !waitFor(t, w, `document.querySelector('[data-role=detail-title]') !== null`) {
		t.Fatal("the detail page did not render")
	}
	if got := evalString(t, w, `String(document.querySelector('[data-role=detail-description]').classList.contains('detail__clamp'))`); got != "true" {
		t.Fatalf("the description is not clamped by default: %s", got)
	}

	evalString(t, w, `document.querySelector('[data-role=detail-expand]').click(), 'ok'`)
	if !waitFor(t, w, `!document.querySelector('[data-role=detail-description]').classList.contains('detail__clamp')`) {
		t.Fatal("the description did not unclamp after expanding")
	}

	if got := evalString(t, w, `String(document.querySelectorAll('[data-current-episode=true]').length)`); got != "1" {
		t.Fatalf("marked last-watched episodes = %s, want 1", got)
	}

	evalString(t, w, `document.querySelector('[data-season="1"]').click(), 'ok'`)
	time.Sleep(150 * time.Millisecond)
	evalString(t, w, `document.querySelector('[data-season="0"]').click(), 'ok'`)
	if !waitFor(t, w, `document.querySelectorAll('[data-current-episode=true]').length === 1`) {
		t.Fatalf("the last-watched mark was lost after switching seasons")
	}

	// 当前窗口宽 800（<840）：选集走追剧过渡页，而不是就地播放。
	if got := evalString(t, w, `String(matchMedia('(max-width: 839px)').matches)`); got != "true" {
		t.Fatalf("narrow media query = %s, want true at this window width", got)
	}
	evalString(t, w, `document.querySelector('[data-episode]').click(), 'ok'`)
	if !waitFor(t, w, `location.hash === '#/follow/film-001'`) {
		t.Fatalf("hash after picking an episode in the narrow layout = %q, want #/follow/film-001",
			evalString(t, w, `location.hash`))
	}
}

// TestM2Shell checks the shell: four navigation entries plus a search entry,
// the current-page marker, and the collapsed layout below 840px.
func TestM2Shell(t *testing.T) {
	w := appWindow(t)

	evalString(t, w, `location.hash = '#/', 'ok'`)
	if !waitFor(t, w, `document.querySelector('[data-role=home]') !== null`) {
		t.Fatal("the home page did not render")
	}

	if got := evalString(t, w, `String(document.querySelectorAll('[data-role=nav-link]').length)`); got != "4" {
		t.Fatalf("navigation entries = %s, want 4", got)
	}
	labels := `[...document.querySelectorAll('[data-role=nav-link]')].map((e) => e.textContent.trim()).join(',')`
	if got := evalString(t, w, labels); got != "首页,发现,追剧,我的" {
		t.Fatalf("navigation labels = %q, want 首页,发现,追剧,我的", got)
	}
	if got := evalString(t, w, `String(document.querySelectorAll('[data-route=search]').length)`); got != "1" {
		t.Fatalf("search entries = %s, want 1", got)
	}

	// 当前窗口宽 800：窄布局生效，导航应为横向排列。
	if got := evalString(t, w, `String(matchMedia('(max-width: 839px)').matches)`); got != "true" {
		t.Fatalf("narrow media query = %s, want true at this window width", got)
	}
	if got := evalString(t, w, `getComputedStyle(document.querySelector('.nav')).flexDirection`); got != "row" {
		t.Fatalf("navigation direction = %q, want row in the collapsed layout", got)
	}

	evalString(t, w, `location.hash = '#/follow', 'ok'`)
	if !waitFor(t, w, `document.querySelector('[aria-current=page]') !== null`) {
		t.Fatal("the current-page marker did not render")
	}
	if got := evalString(t, w, `document.querySelector('[aria-current=page]').dataset.route`); got != "follow-list" {
		t.Fatalf("current nav entry on the follow route = %q, want follow-list", got)
	}
}

// TestM2Home checks the home page: three sections in order, and the error state
// with a retry that recovers once the host stops failing.
func TestM2Home(t *testing.T) {
	w := appWindow(t)

	evalString(t, w, `location.hash = '#/', 'ok'`)
	if !waitFor(t, w, `document.querySelector('[data-role=home]') !== null`) {
		t.Fatal("the home page did not render")
	}
	sections := `[...document.querySelectorAll('.section__title')].map((e) => e.textContent.trim()).join(',')`
	if got := evalString(t, w, sections); got != "继续观看,为你推荐,榜单" {
		t.Fatalf("home sections = %q, want 继续观看,为你推荐,榜单", got)
	}
	for _, role := range []string{"home-watching", "home-feed", "home-ranking"} {
		if got := evalString(t, w, `String(document.querySelectorAll('[data-role=`+role+`]').length)`); got != "1" {
			t.Fatalf("section %s = %s, want 1", role, got)
		}
	}

	evalString(t, w, `location.hash = '#/explore', 'ok'`)
	if !waitFor(t, w, `document.querySelector('[data-role=catalog-item]') !== null`) {
		t.Fatal("the catalogue did not render before the error case")
	}
	evalString(t, w, `(() => { window.__forceHostError = true; location.hash = '#/'; return 'ok'; })()`)
	if !waitFor(t, w, `document.querySelector('[data-role=home-error]') !== null`) {
		t.Fatal("a failing host did not render the error state")
	}

	evalString(t, w, `(() => { window.__forceHostError = false; document.querySelector('[data-role=home-error] button').click(); return 'ok'; })()`)
	if !waitFor(t, w, `document.querySelector('[data-role=home-feed]') !== null`) {
		t.Fatal("retry did not recover the home page")
	}
}

// TestM2Routes checks the parameterised routes: a catalogue item opens its
// detail page, the follow page renders for a known id, and an unknown id falls
// through to the not-found page instead of a blank screen.
func TestM2Routes(t *testing.T) {
	w := appWindow(t)

	// 页面数据必须来自本地 fixture：把 fetch 换成计数器，跑完核对调用次数为 0。
	evalString(t, w, `(() => { window.__netCalls = 0; window.fetch = () => { window.__netCalls++; return Promise.reject(new Error('M2 does not use the network')); }; return 'ok'; })()`)

	evalString(t, w, `location.hash = '#/explore', 'ok'`)
	if !waitFor(t, w, `document.querySelector('[data-role=catalog-item]') !== null`) {
		t.Fatal("the catalogue did not render")
	}

	firstID := evalString(t, w, `document.querySelector('[data-role=catalog-item]').dataset.id`)
	evalString(t, w, `document.querySelector('[data-role=catalog-item]').click(), 'ok'`)
	if !waitFor(t, w, `location.hash === '#/detail/`+firstID+`'`) {
		t.Fatalf("hash after opening the first item = %q, want #/detail/%s", evalString(t, w, `location.hash`), firstID)
	}
	if !waitFor(t, w, `document.querySelector('[data-role=detail-title]') !== null`) {
		t.Fatal("the detail page did not render")
	}
	if got := evalString(t, w, `document.querySelector('[data-role=detail-title]').textContent.trim().length > 0 ? 'ok' : 'empty'`); got != "ok" {
		t.Fatalf("the detail title = %q, want a non-empty title", got)
	}

	evalString(t, w, `location.hash = '#/detail/nope', 'ok'`)
	if !waitFor(t, w, `!!document.querySelector('[data-role=not-found]')`) {
		t.Fatal("an unknown detail id did not render the not-found page")
	}

	evalString(t, w, `location.hash = '#/follow/`+firstID+`', 'ok'`)
	if !waitFor(t, w, `document.querySelector('[data-role=follow]') !== null`) {
		t.Fatal("the follow page did not render for a known id")
	}
	if got := evalString(t, w, `document.querySelector('[data-role=follow-title]').textContent.trim().length > 0 ? 'ok' : 'empty'`); got != "ok" {
		t.Fatalf("the follow title = %q, want a non-empty title", got)
	}

	evalString(t, w, `location.hash = '#/follow/nope', 'ok'`)
	if !waitFor(t, w, `!!document.querySelector('[data-role=not-found]')`) {
		t.Fatal("an unknown follow id did not render the not-found page")
	}

	if got := evalString(t, w, `String(window.__netCalls)`); got != "0" {
		t.Fatalf("the content pages made %s network calls, want 0", got)
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
