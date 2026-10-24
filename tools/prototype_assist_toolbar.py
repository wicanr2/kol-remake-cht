#!/usr/bin/env python3
"""Docker限定：Zeliard風格工具列的可丟棄原型，不讀寫遊戲存檔。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
from PIL import Image, ImageDraw, ImageFont

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--root',type=Path,required=True)
p.add_argument('--reference',type=Path,required=True)
p.add_argument('--out',type=Path,required=True)
a=p.parse_args()
assert Path('/.dockerenv').exists() and not a.out.exists()
assert (a.out.parent.stat().st_uid,a.out.parent.stat().st_gid)==(os.getuid(),os.getgid())
sha=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
fontdir=a.reference/'workplace/full-hd-v1/manual-hd-v1/rounded-font'
font=fontdir/'jf-openhuninn-2.1.ttf'
assert sha(font)=='9d5bf4932d31fe94c18cd8cfddc98bc1b14ce10f4e354c682179db290a99c825'
source=a.root/'docs/verification/player-feedback-20261024/attack-area-zh-Hant.png'
game=Image.open(source).convert('RGBA');assert game.size==(960,600)
a.out.mkdir()
shutil.copy2(font,a.out/'font.ttf')
shutil.copy2(fontdir/'OFL.txt',a.out/'OFL.txt')
shutil.copy2(source,a.out/'game.png')
faces={n:ImageFont.truetype(str(font),n) for n in (20,22,24)}
white=(230,235,245,255);button=(42,52,70,255);active=(40,90,115,255)
top=[((8,8,104,40),'原貌'),((112,8,208,40),'HD畫面'),((216,8,352,40),'輔助 F8'),
     ((360,8,472,40),'配樂 F9'),((480,8,592,40),'畫面 F10'),((600,8,704,40),'暫停'),
     ((712,8,832,40),'快捷鍵'),((840,8,952,40),'收起 F12')]


def label(draw,box,s,size=22,color=white):
    x,y,x2,y2=box;f=faces[size]
    assert draw.textlength(s,font=f)<=x2-x-12,(s,box)
    bounds=draw.textbbox((0,0),s,font=f)
    assert bounds[3]-bounds[1]<=y2-y,(s,box)
    draw.text((x+6,y+(y2-y-(bounds[3]-bounds[1]))/2-bounds[1]),s,font=f,fill=color)


def draw_button(draw,box,s,on=False,size=22):
    draw.rectangle(box,fill=active if on else button)
    label(draw,box,s,size)


for name,expanded,panel in [('toolbar',False,False),('quick-controls',True,False),('character-controls',True,True)]:
    height=96 if expanded else 48
    im=Image.new('RGBA',(960,600+height),(25,32,45,255));im.alpha_composite(game,(0,height));d=ImageDraw.Draw(im)
    for box,text in top:draw_button(d,box,text,text=='原貌' or expanded and text=='輔助 F8')
    if expanded:
        for box,text,on in [((8,56,184,88),'玩家不疲勞：關',False),((192,56,400,88),'敵疲勞倍率：1x',False),
                            ((408,56,552,88),'免怯戰：關',False),((560,56,752,88),'永久角色強化',panel),((760,56,952,88),'戰鬥輔助歸零',False)]:
            draw_button(d,box,text,on,20)
    if panel:
        shade=Image.new('RGBA',im.size);ImageDraw.Draw(shade).rectangle((0,96,960,696),fill=(0,0,0,115));im.alpha_composite(shade);d=ImageDraw.Draw(im)
        d.rectangle((184,136,888,554),fill=(25,32,45,255),outline=(92,125,145,255),width=2)
        label(d,(210,148,862,184),'角色強化（永久）',24)
        label(d,(210,187,862,217),'選擇對象後，調整數值並按套用。',20)
        draw_button(d,(216,224,856,264),'強化對象：目前發言者 / 全隊')
        for y,title,value,verb in [(280,'基本屬性下限','自訂','套用'),(338,'已學武器技能下限','200','套用'),(396,'金幣至少','9000','補足')]:
            label(d,(216,y,518,y+40),title)
            d.rectangle((528,y,650,y+40),fill=(15,22,30,255),outline=(92,125,145,255))
            label(d,(528,y,650,y+40),value)
            draw_button(d,(674,y,850,y+40),verb)
        label(d,(210,456,862,488),'關閉輔助不還原角色變更；住宿時保存。',20)
        draw_button(d,(704,504,856,540),'返回 Esc')
    im.convert('RGB').save(a.out/(name+'.png'))

html='''<!doctype html><html lang="zh-Hant"><meta charset="utf-8"><title>KOL 工具列操作原型</title>
<style>@font-face{font-family:rounded;src:url(font.ttf)}*{box-sizing:border-box}body{background:#10151e;color:#e6ebf5;font-family:rounded,sans-serif;margin:24px}button,input,select{font:20px rounded;color:inherit;background:#2a3446;border:1px solid #526477;padding:5px 9px;height:34px}button{cursor:pointer}button.on{background:#285a73}#app{width:960px;position:relative}.bar{height:48px;display:flex;gap:8px;padding:8px;background:#19202d}.bar button{padding:0 8px;flex:1}#quick{display:none}#quick.open{display:flex}#game{display:block;width:960px;height:600px}#panel{display:none;position:absolute;left:184px;top:136px;width:704px;padding:22px 30px;background:#19202d;border:2px solid #5c7d91;box-shadow:0 0 0 100vmax #0007;clip-path:inset(-40px -184px -142px)}#panel.open{display:block}h2{font-size:24px;margin:0 0 12px}.row{display:flex;align-items:center;gap:16px;margin:16px 0}.row label{width:285px}.row input{width:125px}.row button{width:150px}small{font-size:18px;color:#bcc7d8}.note{margin-top:14px;color:#b7c6d9}#message{min-height:28px}.hidden{display:none!important}</style>
<h1>武士傳說：工具列操作原型</h1><p>參照 Zeliard 工具列；此頁不讀寫角色或存檔。</p>
<div id="app"><div class="bar" id="bar"><button class="on" onclick="notice('原貌畫面')">原貌</button><button onclick="notice('正式遊戲沿既有 HD 切換')">HD畫面</button><button id="assist" onclick="toggleQuick()">輔助 F8</button><button id="music" onclick="toggleMusic()">配樂 F9</button><button onclick="notice('正式遊戲沿既有 F10 畫面設定')">畫面 F10</button><button id="pause" onclick="this.classList.toggle('on')">暫停</button><button onclick="notice('F8 輔助；F9 配樂；F10 畫面；F12 工具列；Esc 返回')">快捷鍵</button><button onclick="toggleBar()">收起 F12</button></div>
<div class="bar" id="quick"><button onclick="flip(this,'玩家不疲勞')">玩家不疲勞：關</button><label>敵疲勞 <select id="enemy"><option>1x</option><option>2x</option><option>3x</option><option>6x</option></select></label><button onclick="flip(this,'免怯戰')">免怯戰：關</button><button onclick="document.getElementById('panel').classList.toggle('open')">永久角色強化</button><button onclick="reset()">戰鬥輔助歸零</button></div>
<img id="game" src="game.png" alt="目前正式遊戲畫面，供工具列位置比較">
<section id="panel"><h2>角色強化（永久）</h2><small>選擇對象後，調整數值並按套用。</small><div class="row"><label>強化對象</label><select><option>目前發言者</option><option>全隊</option></select></div>
<div class="row"><label>基本屬性下限</label><input type="number" min="1" max="255" placeholder="自訂"><button onclick="notice('原型：預覽永久屬性修改，不操作存檔')">套用</button></div>
<div class="row"><label>已學武器技能下限</label><input type="number" min="1" max="200" value="200"><button onclick="notice('原型：預覽永久技能修改，不操作存檔')">套用</button></div>
<div class="row"><label>金幣至少</label><input type="number" min="0" max="16777215" value="9000"><button onclick="notice('原型：預覽金幣補足，不操作存檔')">補足</button></div>
<small>關閉輔助不還原角色變更；住宿時保存。</small><p style="text-align:right"><button onclick="closePanel()">返回 Esc</button></p></section></div>
<p class="note">操作原型，尚未連接正式遊戲。戰鬥輔助每次啟動關閉；角色強化由玩家明確套用。</p><p id="message"></p>
<script>const $=x=>document.getElementById(x);function notice(s){$('message').textContent=s}function toggleQuick(){$('quick').classList.toggle('open');$('assist').classList.toggle('on');closePanel()}function toggleBar(){$('bar').classList.toggle('hidden');$('quick').classList.remove('open');closePanel()}function toggleMusic(){$('music').classList.toggle('on');notice('配樂狀態已切換，原型沒有音訊')}function closePanel(){$('panel').classList.remove('open')}function flip(b,s){b.classList.toggle('on');b.textContent=s+'：'+(b.classList.contains('on')?'開':'關')}function reset(){document.querySelectorAll('#quick button.on').forEach(b=>b.click());$('enemy').selectedIndex=0;notice('已關閉戰鬥輔助；不還原角色數值')}document.addEventListener('keydown',e=>{if(e.target.tagName==='INPUT')return;if(['F8','F9','F10','F12','Escape'].includes(e.key))e.preventDefault();if(e.key==='F8')toggleQuick();if(e.key==='F9')toggleMusic();if(e.key==='F10')notice('正式遊戲沿既有畫面設定');if(e.key==='F12')toggleBar();if(e.key==='Escape')closePanel()});</script></html>'''
(a.out/'review.html').write_text(html)
manifest=dict(status='DRAFT_UI_ONLY',production_code_changed=False,save_data_changed=False,
              reference_project='~/cht/zeliard',reference_files=['cmd/zeliard-preview/main.go','cmd/zeliard-preview/cheats.go','cmd/zeliard-preview/speed.go'],
              game_source=str(source.relative_to(a.root)),game_source_sha256=sha(source),
              font_source=str(font.relative_to(a.reference)),font_sha256=sha(font),font_license='OFL-1.1',
              decisions_confirmed=['可選輔助，預設關閉','參照Zeliard工具列','常用功能直接放工具列並支援快捷鍵','角色屬性技能永久修改，正常保存','README實機截圖及功能說明'],
              proposal=['戰鬥輔助獨立開關','敵疲勞可選1/2/3/6倍','永久強化分別套用，對象可選發言者或全隊','屬性下限自訂；已學技能200、金幣9000為可調起始值'],
              artifacts={p.name:sha(p) for p in a.out.iterdir() if p.is_file()})
(a.out/'plan.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
assert all((p.stat().st_uid,p.stat().st_gid)==(os.getuid(),os.getgid()) for p in a.out.iterdir())
print('DRAFT TOOLBAR PROTOTYPE READY',a.out,flush=True)
