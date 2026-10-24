package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/save"
)

func TestInteractiveSavePaths(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	p, err := defaultSavePath()
	if err != nil {
		t.Fatal(err)
	}
	custom := filepath.Join(t.TempDir(), "other.dat")
	for _, tc := range []struct {
		name, write, read   string
		explicit, exists    bool
		wantWrite, wantRead string
	}{
		{"first", "", "", false, false, p, ""},
		{"resume", "", "", false, true, p, p},
		{"new", "", "", true, true, p, ""},
		{"custom-load", "", custom, true, true, custom, custom},
		{"custom-save", custom, "", false, true, custom, custom},
		{"separate", p, custom, true, true, p, custom},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Remove(p)
			_ = os.Remove(custom)
			if tc.exists {
				for _, f := range []string{p, custom} {
					if err := os.WriteFile(f, []byte("broken"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			w, r, err := interactiveSavePaths(tc.write, tc.read, tc.explicit)
			if err != nil || w != tc.wantWrite || r != tc.wantRead {
				t.Fatalf("%q %q %v", w, r, err)
			}
			if r != "" { // 路由不能吞掉壞檔，現有 parser 須拒絕它，檔案保持。
				if _, err := save.Read(r); err == nil {
					t.Fatal("壞檔被接受")
				}
				b, _ := os.ReadFile(r)
				if string(b) != "broken" {
					t.Fatal("壞檔被改寫")
				}
			}
		})
	}
	if _, _, err := interactiveSavePaths(t.TempDir(), "", false); err == nil {
		t.Fatal("目錄被當存檔")
	}
}
