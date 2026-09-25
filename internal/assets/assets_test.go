package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runtime 只解碼 PNG（#26）：manifest 列了別的副檔名、逃出資產目錄，或內容其實
// 不是 PNG，都要拒絕，不能退而讀原版格式。
func TestDecodeRejectsNonPNG(t *testing.T) {
	root := t.TempDir()
	manifest := `{"schema":"kol-ega-production-manifest-test","archives":[{"archive":"X","images":[
 {"unit_index":0,"file":"KOLDATA.0","width":1,"height":1},
 {"unit_index":1,"file":"../escape.png","width":1,"height":1},
 {"unit_index":2,"file":"fake.png","width":1,"height":1}]}]}`
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "fake.png"), []byte("MZ not a png"), 0o644); err != nil {
		t.Fatal(err)
	}
	lib, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	for index, want := range map[int]string{0: "非 PNG", 1: "資產目錄之外", 2: "解碼"} {
		if _, _, err := lib.Decode("X", index); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("圖號 %d 應以「%s」拒絕，實際 %v", index, want, err)
		}
	}
}
