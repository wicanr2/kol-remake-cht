package main

import (
 "encoding/hex"
 "encoding/json"
 "image"
 "image/color"
 "image/draw"
 "image/png"
 "os"
 "path/filepath"
)

type glyph struct { Character string `json:"character"`; Advance int `json:"advance"`; Rows string `json:"rows_hex"` }
type manifest struct { Fonts []struct { Glyphs []glyph `json:"glyphs"`; Advance int `json:"advance"` } `json:"fonts"` }
func must(e error) { if e!=nil {panic(e)} }
func load(p string) image.Image {f,e:=os.Open(p);must(e);defer f.Close();im,e:=png.Decode(f);must(e);return im}
func main() {
 root:="workplace/cjk-longest-input-20261001/attempt2/nameplate-prototype"
 must(os.Mkdir(root,0755))
 bg:=load("workplace/cjk-longest-input-20261001/attempt2/letters/zh-Hant/step-029.png")
 paper:=load("assets/original-derived-png/ega/BOOKPICS/bookpics-002.png")
 raw,e:=os.ReadFile("assets/ui-font/manifest.json");must(e);var m manifest;must(json.Unmarshal(raw,&m))
 gs:=map[rune]glyph{};for _,g:=range m.Fonts[0].Glyphs {gs[[]rune(g.Character)[0]]=g}
 white:=color.RGBA{255,255,255,255};black:=color.RGBA{0,0,0,255};cyan:=color.RGBA{0,170,170,255}
 for _,choice:=range []struct{key,title string}{{"A","方案 A：加寬名牌，完整姓名"},{"B","方案 B：維持名牌，縮略姓名"}} {
  canvas:=image.NewRGBA(image.Rect(0,0,960,640));draw.Draw(canvas,canvas.Bounds(),&image.Uniform{black},image.Point{},draw.Src)
  draw.Draw(canvas,image.Rect(0,40,960,640),bg,bg.Bounds().Min,draw.Src)
  text:=func(s string,x,y int,c color.RGBA) {
   for _,r:=range s {
    if r==' ' {x+=m.Fonts[0].Advance;continue};g,ok:=gs[r];if !ok {panic("缺少原型字形："+string(r))}
    b,e:=hex.DecodeString(g.Rows);must(e)
    for yy:=0;yy<len(b)/2;yy++ {mask:=uint16(b[yy*2])<<8|uint16(b[yy*2+1]);for xx:=0;xx<16;xx++ {if mask&(1<<uint(15-xx))!=0 {canvas.SetRGBA(x+xx,y+yy,c)}}};x+=g.Advance
   }
  }
  text(choice.title,18,10,white)
  // 以來源書頁重畫名牌所在列；這是可丟棄的原生介面原型，不改正式 renderer。
  for y:=432;y<459;y++ {for x:=550;x<890;x++ {canvas.Set(x,y+40,paper.At(x/3,y/3))}}
  left,right:=618,747;name:="WWWWWWW…";nameX:=624
  if choice.key=="A" {left,right=558,806;name="WWWWWWWWWWWWWWWW";nameX=562}
  draw.Draw(canvas,image.Rect(left,472,right,499),&image.Uniform{cyan},image.Point{},draw.Src)
  draw.Draw(canvas,image.Rect(left+2,474,right-2,497),&image.Uniform{black},image.Point{},draw.Src)
  text(name,nameX,478,white)
  f,e:=os.OpenFile(filepath.Join(root,"option-"+choice.key+".png"),os.O_CREATE|os.O_EXCL|os.O_WRONLY,0644);must(e);must(png.Encode(f,canvas));must(f.Close())
 }
}
