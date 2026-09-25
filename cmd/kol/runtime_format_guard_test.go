package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// runtime 只讀衍生格式（CLAUDE.md「圖像」一節）：圖像是 PNG、音訊是 OGG，型別資料是
// JSON／TSV，都放在 assets/ 或 docs/ 底下。原版封裝、EXE、archive 與說明書 PDF
// 只給抽取器用，runtime 一個都不能讀（#26）。
//
// Go 程式直接發系統呼叫，容器裡也沒有 strace，所以這裡用靜態守門：
//  1. 讀檔路徑全部來自 main.go 的旗標預設值，逐一檢查位置與副檔名。
//  2. 非測試程式碼裡不得出現像原版檔名的字面值（不含空白的字串才算路徑）。

var runtimeAllowedExt = map[string]bool{"": true, ".json": true, ".png": true, ".tsv": true, ".ogg": true}

var originalLiteral = regexp.MustCompile(`(?i)\.(zip|exe|pdf|ovl|drv)$|knights-of-legend|^[a-z0-9_]+\.[0-9]$`)

// originalLiterals 回傳一段 Go 原始碼裡像原版檔名的字串字面值。
func originalLiterals(t *testing.T, name string, src any) []string {
	t.Helper()
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, name, src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var hits []string
	ast.Inspect(f, func(n ast.Node) bool {
		if _, ok := n.(*ast.ImportSpec); ok {
			return false
		}
		b, ok := n.(*ast.BasicLit)
		if !ok || b.Kind != token.STRING {
			return true
		}
		s, err := strconv.Unquote(b.Value)
		if err == nil && !strings.ContainsAny(s, " \t\n") && originalLiteral.MatchString(s) {
			hits = append(hits, fs.Position(b.Pos()).String()+" "+s)
		}
		return true
	})
	return hits
}

func TestRuntimeGuardDetectsOriginalLiterals(t *testing.T) {
	src := `package x
var a = "Knights-of-legend.zip"
var b = "KOLDATA.0"
var c = "說明文字提到 TITLE.EXE 不算路徑"
var d = "assets/ui-font/manifest.json"`
	hits := originalLiterals(t, "fixture.go", src)
	if len(hits) != 2 {
		t.Fatalf("守門應只抓到兩個原版檔名，實際 %q", hits)
	}
}

func TestRuntimeReadsNoOriginalFiles(t *testing.T) {
	var hits []string
	for _, root := range []string{"../../cmd", "../../internal"} {
		err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return nil
			}
			hits = append(hits, originalLiterals(t, p, nil)...)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(hits) > 0 {
		t.Fatalf("runtime 程式碼出現原版檔名：\n%s", strings.Join(hits, "\n"))
	}
}

func TestRuntimeFlagDefaultsAreDerivedFormats(t *testing.T) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "main.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "String" {
			return true
		}
		if pkg, ok := sel.X.(*ast.Ident); !ok || pkg.Name != "flag" {
			return true
		}
		lit, ok := call.Args[1].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		def, _ := strconv.Unquote(lit.Value)
		if !strings.Contains(def, "/") {
			return true
		}
		checked++
		if !strings.HasPrefix(def, "assets/") && !strings.HasPrefix(def, "docs/") {
			t.Errorf("%s：預設路徑 %q 不在 assets/ 或 docs/", fs.Position(lit.Pos()), def)
		}
		if ext := strings.ToLower(filepath.Ext(def)); !runtimeAllowedExt[ext] {
			t.Errorf("%s：預設路徑 %q 的格式 %q 不是 runtime 允許的衍生格式", fs.Position(lit.Pos()), def, ext)
		}
		return true
	})
	if checked < 20 {
		t.Fatalf("只檢查到 %d 個路徑旗標，解析可能漏了", checked)
	}
}
