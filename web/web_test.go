package web

import (
	"io/fs"
	"strings"
	"testing"
)

func TestFSRootsAtDist(t *testing.T) {
	fsys, err := FS()
	if err != nil {
		t.Fatalf("FS() error = %v", err)
	}

	b, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}

	if !strings.Contains(string(b), `data-app="guoappgo"`) {
		t.Fatalf("index.html is missing the app marker")
	}
}
