package web

import (
	"io/fs"
	"strings"
	"testing"
)

// TestDistEmbedded guards the embedded output contract: the embedded tree has
// the page, the page carries the app marker and title, and at least one
// non-empty script was built next to it.
func TestDistEmbedded(t *testing.T) {
	fsys, err := FS()
	if err != nil {
		t.Fatalf("FS() error = %v", err)
	}

	html, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	page := string(html)
	if !strings.Contains(page, `data-app="guoappgo"`) {
		t.Fatal(`index.html is missing data-app="guoappgo"`)
	}
	if !strings.Contains(page, "<title>guoAppGo</title>") {
		t.Fatal("index.html is missing the page title")
	}
	if !strings.Contains(page, `type="module"`) {
		t.Fatal("index.html does not load a module script")
	}

	scripts := 0
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".js") {
			return nil
		}
		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		if len(b) == 0 {
			t.Fatalf("%s is empty", path)
		}
		scripts++
		return nil
	})
	if err != nil {
		t.Fatalf("walk web/dist: %v", err)
	}
	if scripts == 0 {
		t.Fatal("no .js file in web/dist, run npm run build")
	}
}
