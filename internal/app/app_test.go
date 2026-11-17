package app

import "testing"

func TestWindowShowsTheFrontendPage(t *testing.T) {
	opts := Window()

	if opts.URL != "/" {
		t.Fatalf("URL = %q, want %q", opts.URL, "/")
	}
	if opts.Width <= 0 || opts.Height <= 0 {
		t.Fatalf("size = %dx%d, want positive", opts.Width, opts.Height)
	}
	if opts.MinWidth > opts.Width || opts.MinHeight > opts.Height {
		t.Fatalf("minimum %dx%d exceeds %dx%d", opts.MinWidth, opts.MinHeight, opts.Width, opts.Height)
	}
	if opts.StateKey == "" {
		t.Fatal("StateKey is empty, the window would not restore its place")
	}
	if opts.Title != "guoAppGo" {
		t.Fatalf("Title = %q", opts.Title)
	}
}

func TestSystemReportsVersion(t *testing.T) {
	if got := (System{version: Version}).Version(); got != Version {
		t.Fatalf("Version() = %q, want %q", got, Version)
	}
	if Version == "" {
		t.Fatal("Version is empty")
	}
}
