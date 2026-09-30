// Package layers 只放一支守門測試：檢查套件之間的依賴方向，以及規則層有沒有偷讀時鐘與亂數。
// 契約全文見 docs/spec/038-package-boundaries.md（#29）。
package layers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const module = "github.com/wicanr2/kol-remake-cht/"

// allowed 是每個套件**允許**依賴的專案內套件。表裡沒有的依賴一律失敗：新增依賴要改這張表，
// 而改表就是一次看得見的決定。`cmd/kol` 是最外層，可以依賴全部。
var allowed = map[string][]string{
	// 資料與規則：彼此獨立，沒有 UI、沒有 Ebitengine。
	"internal/assets":     {},
	"internal/battle":     {},
	"internal/bitmapedit": {},
	"internal/book":       {},
	"internal/chardata":   {},
	"internal/ending":     {},
	"internal/input":      {},
	"internal/magic":      {},
	"internal/portrait":   {},
	"internal/rng":        {},
	"internal/roster":     {},
	"internal/save":       {},
	"internal/title":      {},
	"internal/world":      {},
	// 依賴一兩個資料套件的規則。
	"internal/charcreate": {"internal/chardata"},
	"internal/i18n":       {"internal/magic"},
	"internal/town":       {"internal/magic", "internal/portrait"},
	// 場景串接：只看規則，不看畫面。
	"internal/flow": {"internal/bitmapedit", "internal/charcreate", "internal/roster", "internal/town"},
	// 對拍探針的輔助套件，只讀戰鬥規則。
	"internal/oracle": {"internal/battle"},
	// 畫面層：唯一可以依賴 Ebitengine 的內部套件。
	"internal/font":   {},
	"internal/screen": {},
	"internal/ui": {
		"internal/battle", "internal/bitmapedit", "internal/book", "internal/charcreate",
		"internal/ending", "internal/font", "internal/i18n", "internal/portrait",
		"internal/roster", "internal/town", "internal/world",
	},
}

// presentation 是可以依賴 Ebitengine 的套件（含最外層）。
var presentation = map[string]bool{
	"cmd/kol": true, "internal/font": true, "internal/screen": true, "internal/ui": true,
}

// clockAndRandom 記下規則層以外**目前**直接讀時鐘或亂數的檔案。規則層一律不准；
// 這兩處在最外層：存檔時間戳記與新角色的隨機 ID。
var clockAndRandom = map[string][]string{
	"cmd/kol/main.go":            {"time.Now"},
	"cmd/kol/creation_commit.go": {"crypto/rand"},
}

var bannedTime = map[string]bool{
	"Now": true, "Since": true, "Until": true, "After": true, "AfterFunc": true,
	"Tick": true, "NewTicker": true, "NewTimer": true, "Sleep": true,
}

var bannedImports = map[string]bool{"math/rand": true, "math/rand/v2": true, "crypto/rand": true}

func root(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("找不到 go.mod")
		}
		dir = parent
	}
}

// parsed 是一個套件（一個目錄）裡的非測試檔。
type parsed struct {
	dir   string
	files map[string]*ast.File // 相對專案根的路徑 → 語法樹
}

// load 讀 internal/* 與 cmd/kol 的非測試檔。
func load(t *testing.T) map[string]*parsed {
	t.Helper()
	base := root(t)
	var dirs []string
	entries, err := os.ReadDir(filepath.Join(base, "internal"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, "internal/"+e.Name())
		}
	}
	dirs = append(dirs, "cmd/kol")
	out := map[string]*parsed{}
	fset := token.NewFileSet()
	for _, d := range dirs {
		files, err := os.ReadDir(filepath.Join(base, d))
		if err != nil {
			t.Fatal(err)
		}
		p := &parsed{dir: d, files: map[string]*ast.File{}}
		for _, f := range files {
			name := f.Name()
			if f.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			rel := d + "/" + name
			file, err := parser.ParseFile(fset, filepath.Join(base, rel), nil, 0)
			if err != nil {
				t.Fatalf("%s：%v", rel, err)
			}
			p.files[rel] = file
		}
		if len(p.files) > 0 {
			out[d] = p
		}
	}
	return out
}

func imports(f *ast.File) []string {
	var out []string
	for _, spec := range f.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err == nil {
			out = append(out, path)
		}
	}
	return out
}

// 每個內部套件都要在表裡有一列，表裡也不能有不存在的套件。
func TestEveryPackageHasALayer(t *testing.T) {
	pkgs := load(t)
	for d := range pkgs {
		if d == "cmd/kol" {
			continue
		}
		if _, ok := allowed[d]; !ok {
			t.Errorf("%s 沒有登記在 allowed：新增套件要先決定它在哪一層", d)
		}
	}
	for d := range allowed {
		if _, ok := pkgs[d]; !ok {
			t.Errorf("allowed 有 %s，但這個套件沒有非測試檔", d)
		}
	}
}

// 依賴方向：套件只能依賴 allowed 列出的內部套件。
func TestDependencyDirection(t *testing.T) {
	pkgs := load(t)
	for d, p := range pkgs {
		if d == "cmd/kol" {
			continue
		}
		ok := map[string]bool{}
		for _, a := range allowed[d] {
			ok[a] = true
		}
		for rel, f := range p.files {
			for _, imp := range imports(f) {
				dep, internal := strings.CutPrefix(imp, module)
				if !internal || dep == d {
					continue
				}
				if !ok[dep] {
					t.Errorf("%s 依賴 %s：不在 %s 允許的依賴裡", rel, dep, d)
				}
			}
		}
	}
}

// 只有畫面層與最外層可以依賴 Ebitengine；規則與資料套件不行。
func TestOnlyPresentationImportsEbiten(t *testing.T) {
	for d, p := range load(t) {
		if presentation[d] {
			continue
		}
		for rel, f := range p.files {
			for _, imp := range imports(f) {
				if strings.HasPrefix(imp, "github.com/hajimehoshi/ebiten") {
					t.Errorf("%s 依賴 %s：只有 cmd/kol、ui、font、screen 可以碰 Ebitengine", rel, imp)
				}
			}
		}
	}
}

// 規則層不讀時鐘、不用全域亂數：時間與亂數都要由外層注入（Rand 介面、回合數、按鍵 token）。
func TestNoAmbientClockOrRandom(t *testing.T) {
	seen := map[string]map[string]bool{}
	for d, p := range load(t) {
		for rel, f := range p.files {
			found := map[string]bool{}
			for _, imp := range imports(f) {
				if bannedImports[imp] {
					found[imp] = true
				}
			}
			ast.Inspect(f, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if id, ok := sel.X.(*ast.Ident); ok && id.Name == "time" && bannedTime[sel.Sel.Name] {
					found["time."+sel.Sel.Name] = true
				}
				return true
			})
			seen[rel] = found
			_ = d
			ok := map[string]bool{}
			for _, a := range clockAndRandom[rel] {
				ok[a] = true
			}
			for name := range found {
				if !ok[name] {
					t.Errorf("%s 直接使用 %s：規則層不准讀時鐘或全域亂數，要由外層注入", rel, name)
				}
			}
		}
	}
	// 例外清單不能過期：列了就要真的用到，否則刪掉。
	var stale []string
	for rel, names := range clockAndRandom {
		for _, name := range names {
			if !seen[rel][name] {
				stale = append(stale, rel+" "+name)
			}
		}
	}
	sort.Strings(stale)
	for _, s := range stale {
		t.Errorf("clockAndRandom 列了 %s，但檔案已經沒有這個用法，請從例外清單移除", s)
	}
}
