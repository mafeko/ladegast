package webui

import (
	"io/fs"
	"testing"
)

func TestAssets_ContainsIndexHTML(t *testing.T) {
	assets, err := Assets()
	if err != nil {
		t.Fatalf("Assets: %v", err)
	}
	data, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("index.html is empty")
	}
}
