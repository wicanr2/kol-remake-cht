'use strict';
const params=new URLSearchParams(location.search);
const projectRoot=new URL(location.pathname.includes('/workplace/')?'../../':'../../../',location.href);
const local=(name)=>new URL(name,location.href).href;
const root=(name)=>new URL(name,projectRoot).href;
const sources={doorEga:local('door-ega.png'),innEga:local('inn-ega.png'),battleEga:local('battle-ega.png'),doorHd:root('docs/verification/hd-runtime-20261011/inn-door.png'),innHd:root('docs/verification/hd-runtime-20261011/six.png'),battleHd:root('docs/verification/hd-runtime-20261011/battle.png'),party:root('assets/hd/first-slice/party-painted-v1.png'),facade:root('assets/hd/first-slice/inn-painted-v1.png'),npc:local('innkeeper-painted-v1.png'),font:local('font-atlas.png')};
const images={},names=['Theod','Poont','Olear','Morai','Fight','Kenne'];
const commands={door:[['door','進門'],['back','返回']],inn:[['door','返回']],battle:[['yes','確認'],['sword','武器攻擊'],['fist','徒手'],['swap','切換武器'],['drop','丟棄'],['move','移動'],['rest','休息']]};
const state={scene:params.get('scene')||'inn',theme:params.get('theme')||'hd',member:0,command:0,menu:params.has('menu'),menuCursor:1,baseline:params.get('baseline')||'hd',compare:!params.has('single')};
const baselineCanvas=document.querySelector('#base'),canvas=document.querySelector('#candidate');
const output=document.querySelector('output'),screens=document.querySelector('.screens');
const capture=params.get('capture');if(capture){document.body.classList.add('capture',capture==='pair'?'pair-shot':'single-shot');state.compare=capture==='pair'}
const partyRects=names.map((_,i)=>[84+i*288,6,256,288]);
const commandRects=Array.from({length:7},(_,i)=>[192+i*192,1020,144,136]);
const controls={button:[1704,1152,192,48],panel:[1144,352,744,712],modes:[[1184,496,664,148],[1184,664,664,148]],back:[1416,952,432,72]};
const colors={ink:'#eee1ca',gold:'#d4b070',muted:'#b6a99a'};
const tintCache={},checks=[],fits=[];let fontInfo;let lastAction='';
const sceneLabels={door:'旅店門面',inn:'半途之家 · 戴文·托洛克',battle:'戰鬥規劃 · 第1回合'};
function rectContains(r,x,y){return x>=r[0]&&y>=r[1]&&x<r[0]+r[2]&&y<r[1]+r[3]}
function point(clientX,clientY,b=canvas.getBoundingClientRect()){return [(clientX-b.left)*1920/b.width,(clientY-b.top)*1200/b.height]}
function hit(x,y){
 if(rectContains(controls.button,x,y))return {kind:'toggle'};
 if(state.menu){for(let i=0;i<2;i++)if(rectContains(controls.modes[i],x,y))return {kind:'theme',index:i};if(rectContains(controls.back,x,y))return {kind:'back'};return null}
 if(state.scene!=='battle')for(let i=0;i<partyRects.length;i++)if(rectContains(partyRects[i],x,y))return {kind:'member',index:i};
 for(let i=0;i<commands[state.scene].length;i++)if(rectContains(commandRects[i],x,y))return {kind:'command',index:i};
 return null;
}
function tinted(color){if(!tintCache[color]){const c=document.createElement('canvas');c.width=images.font.width;c.height=images.font.height;const x=c.getContext('2d');x.drawImage(images.font,0,0);x.globalCompositeOperation='source-in';x.fillStyle=color;x.fillRect(0,0,c.width,c.height);tintCache[color]=c}return tintCache[color]}
function measure(s,size){return Array.from(s).reduce((w,ch)=>{if(!fontInfo.glyphs[ch])throw Error('缺少字模：'+ch);return w+fontInfo.glyphs[ch].advance*size/64},0)}
function label(ctx,s,r,size=32,color=colors.ink,align='left'){
 const original=s;let truncated=false;while(measure(s,size)>r[2]&&s.length){s=Array.from(s).slice(0,-1).join('');truncated=true}if(truncated){while(measure(s+'…',size)>r[2]&&s.length)s=Array.from(s).slice(0,-1).join('');s+='…'}
 const width=measure(s,size),scale=size/64;let x=r[0]+(align==='center'?(r[2]-width)/2:0);const baseline=r[1]+(r[3]-size)/2+size*.88;
 const ink=tinted(color);for(const ch of s){const g=fontInfo.glyphs[ch];ctx.drawImage(ink,g.x,g.y,96,96,x-10*scale,baseline-76*scale,96*scale,96*scale);x+=g.advance*scale}
 fits.push({text:original,rect:r,size,width,truncated});return width;
}
function path(ctx,d,fill,stroke,width=2){const p=new Path2D(d);if(fill){ctx.fillStyle=fill;ctx.fill(p)}if(stroke){ctx.strokeStyle=stroke;ctx.lineWidth=width;ctx.stroke(p)}}
function panel(ctx,r,{selected=false,red=false}={}){
 const [x,y,w,h]=r;ctx.save();const g=ctx.createLinearGradient(x,y,x,y+h);g.addColorStop(0,red?'#64232a':'#28201e');g.addColorStop(1,red?'#31151a':'#111316');ctx.fillStyle=g;ctx.beginPath();ctx.roundRect(x,y,w,h,4);ctx.fill();ctx.strokeStyle=selected?'#f2cf82':'#96764b';ctx.lineWidth=selected?4:2;ctx.stroke();ctx.strokeStyle='#473a2c';ctx.lineWidth=1;ctx.strokeRect(x+6,y+6,w-12,h-12);
 for(const [a,b] of [[x+10,y+10],[x+w-10,y+10],[x+10,y+h-10],[x+w-10,y+h-10]]){ctx.fillStyle=selected?'#e4be70':'#9e7b47';ctx.beginPath();ctx.moveTo(a,b-4);ctx.lineTo(a+4,b);ctx.lineTo(a,b+4);ctx.lineTo(a-4,b);ctx.closePath();ctx.fill()}ctx.restore();
}
function icon(ctx,key,x,y,size){
 ctx.save();ctx.translate(x,y);ctx.scale(size/100,size/100);ctx.lineJoin='round';ctx.lineCap='round';const gold='#d6b77a',silver='#d9dde0',red='#9c3540';
 const shapes={door:()=>{path(ctx,'M22 90V12H78V90M30 90V22H64V90',null,gold,5);path(ctx,'M34 26L63 19V85L34 91Z','#542829',gold,2);ctx.fillStyle=gold;ctx.beginPath();ctx.arc(54,57,3,0,Math.PI*2);ctx.fill()},back:()=>{path(ctx,'M28 43H53C78 43 83 76 58 86',null,gold,9);path(ctx,'M34 25L14 43L34 62',null,gold,9)},yes:()=>{path(ctx,'M18 48H30V88H18ZM32 49L47 29L51 12Q66 11 64 33L61 45H82Q92 45 89 57L82 84Q80 90 70 90H34Z',gold,'#f1dcc0',2)},sword:()=>{path(ctx,'M34 73L73 16L88 10L85 28L44 81Z',silver,'#ede4d0',2);path(ctx,'M27 64L51 82M20 90L37 73',null,gold,7);path(ctx,'M51 50L83 15',null,'#899599',2)},fist:()=>{path(ctx,'M25 84L18 60L20 38Q24 29 32 36L34 22Q42 16 48 24Q55 15 62 25Q71 20 77 30L79 49L86 59L78 85Z','#c3a27a',gold,3);path(ctx,'M33 39V56M48 29V54M62 29V55M75 37V57M23 62L47 70L70 59',null,'#5c4030',3)},swap:()=>{path(ctx,'M15 31H79M64 17L81 31L64 45M85 69H21M36 55L19 69L36 83',null,gold,7)},drop:()=>{path(ctx,'M52 12V56M34 41L52 59L69 41',null,gold,7);path(ctx,'M23 76L50 62L81 76L52 92Z','#51322c',gold,3)},move:()=>{path(ctx,'M53 18L63 30L59 47L75 62L69 69L54 56L45 73L34 91L24 87L37 67L44 49L29 56L22 49L41 37L47 20Z',gold,'#f2d299',2);ctx.fillStyle=silver;ctx.beginPath();ctx.arc(59,10,8,0,Math.PI*2);ctx.fill()},rest:()=>{path(ctx,'M16 42V91M17 74H87V91M19 53H38V71H19M39 54H87V75H39',null,gold,5);path(ctx,'M41 53Q52 37 81 49V66H41Z','#637b8c',silver,2)},shield:()=>{path(ctx,'M50 10L83 21V49Q80 76 50 93Q20 76 17 49V21Z','#354b58',gold,4);path(ctx,'M50 17V83M24 47H76',null,gold,3)},arrow:()=>{path(ctx,'M50 87V17M28 38L50 15L72 38',null,gold,8)}};
 (shapes[key]||shapes.shield)();ctx.restore();
}
function background(ctx){const g=ctx.createRadialGradient(1050,450,100,1000,600,1400);g.addColorStop(0,'#632329');g.addColorStop(.6,'#37181e');g.addColorStop(1,'#1b1117');ctx.fillStyle=g;ctx.fillRect(0,0,1920,1200);ctx.strokeStyle='#92683d';ctx.lineWidth=2;ctx.strokeRect(18,18,1884,1164);ctx.strokeStyle='#663c35';ctx.lineWidth=1;for(let y=330;y<980;y+=26){ctx.beginPath();ctx.moveTo(32,y);ctx.bezierCurveTo(160,y-35,180,y+38,300,y);ctx.stroke();ctx.beginPath();ctx.moveTo(1798,y);ctx.bezierCurveTo(1850,y-30,1875,y+30,1890,y);ctx.stroke()}panel(ctx,[48,986,1776,204],{red:true})}
function party(ctx){partyRects.forEach((r,i)=>{panel(ctx,r,{selected:i===state.member});ctx.drawImage(images.party,(i%3)*512,Math.floor(i/3)*512,512,512,r[0]+8,r[1]+8,r[2]-16,230);label(ctx,names[i],[r[0]+12,r[1]+244,r[2]-24,34],28,i===state.member?colors.gold:colors.ink,'center')})}
function toolbar(ctx){commands[state.scene].forEach(([key,title],i)=>{const r=commandRects[i];panel(ctx,r,{selected:i===state.command});icon(ctx,key,r[0]+35,r[1]+12,74);label(ctx,title,[r[0]+10,r[1]+96,r[2]-20,30],26,colors.ink,'center')})}
function crest(ctx){ctx.save();ctx.translate(140,512);path(ctx,'M0 0H300V210Q265 260 150 303Q35 260 0 210Z','#d9d1bb','#c4a976',6);path(ctx,'M14 14H146V151H14ZM154 159H286V211Q252 251 154 286Z','#722631');path(ctx,'M154 14H286V151H154Z','#283d54');path(ctx,'M14 159H146V286Q46 249 14 211Z','#1e2433');path(ctx,'M50 50H105V130H50ZM43 38H59V60H43ZM67 38H83V60H67ZM91 38H110V60H91Z','#c6c7c0','#eee1ca',3);ctx.fillStyle='#dbbd70';ctx.beginPath();ctx.arc(220,87,33,0,Math.PI*2);ctx.fill();for(let i=0;i<8;i++){ctx.save();ctx.translate(220,87);ctx.rotate(i*Math.PI/4);path(ctx,'M0 -44V-56',null,'#dbbd70',4);ctx.restore()}path(ctx,'M100 184Q42 183 54 234Q64 261 100 245Q68 231 80 206Z','#c4ccd3');path(ctx,'M257 177L209 209L244 209L187 257L217 218L185 218Z','#d3d9de');ctx.restore()}
function healthyBody(ctx){ctx.save();ctx.translate(476,516);ctx.scale(4,4);path(ctx,'M26 0Q13 0 13 14Q14 21 20 24L20 28L9 32L1 45L4 52L14 41L17 45L12 59L18 66L19 77L12 95L15 103L25 91L29 76L33 91L43 103L46 95L39 77L40 66L46 59L41 45L44 41L54 52L57 45L49 32L37 28L37 24Q44 21 44 14Q44 0 30 0Z','#c4c9c8','#e4d8b7',.8);path(ctx,'M21 31L29 38L36 31M18 48Q28 53 40 48M21 61L29 67L37 61M29 39V60M19 78L25 71M39 78L33 71',null,'#56626a',1.2);ctx.restore()}
function drawHD(ctx){
 background(ctx);if(state.scene!=='battle')party(ctx);
 if(state.scene==='door'){panel(ctx,[84,342,400,132]);label(ctx,'四月一日',[110,356,344,42],32);label(ctx,'西元 1100',[110,412,344,40],30,colors.muted);crest(ctx);panel(ctx,[618,342,1164,588]);ctx.drawImage(images.facade,624,348,1152,576)}
 if(state.scene==='inn'){panel(ctx,[54,390,514,372]);ctx.drawImage(images.npc,0,90,images.npc.width,images.npc.height-260,62,398,498,356);panel(ctx,[612,350,1166,118]);label(ctx,'半途之家',[642,360,1106,44],34,colors.gold,'center');label(ctx,'戴文·托洛克',[642,412,1106,38],28,colors.muted,'center');panel(ctx,[612,504,1166,406]);const lines=['「日安，」他用很尖的嗓音說，','「這是什麼？幾位戰士嗎？」','「哪一位需要幫忙？」'];lines.forEach((s,i)=>label(ctx,s,[650,526+i*50,1090,44],34))}
 if(state.scene==='battle'){
  panel(ctx,[96,12,736,450]);label(ctx,'Poontang',[144,32,644,50],34);ctx.drawImage(images.party,512,0,512,512,260,80,368,368);
  panel(ctx,[48,492,808,456]);panel(ctx,[906,42,878,882]);ctx.imageSmoothingEnabled=false;ctx.drawImage(images.battleEga,453,21,439,441,906,42,878,882);ctx.imageSmoothingEnabled=true;
  label(ctx,'攻擊',[96,544,150,44],32);icon(ctx,'sword',128,602,90);label(ctx,'方向',[288,544,140,44],32);icon(ctx,'arrow',314,602,90);label(ctx,'部位',[96,744,150,44],32);icon(ctx,'arrow',126,798,90);label(ctx,'防禦',[288,744,140,44],32);icon(ctx,'shield',314,798,90);healthyBody(ctx);
 }
 toolbar(ctx);
}
function settings(ctx){
 panel(ctx,controls.button);label(ctx,'畫面 F10',[1716,1158,168,36],28,colors.ink,'center');if(!state.menu)return;
 ctx.fillStyle='rgba(0,0,0,.55)';ctx.fillRect(0,0,1920,1200);panel(ctx,controls.panel,{red:true});
 label(ctx,'畫面設定',[1184,384,664,52],38);label(ctx,'切換後立即預覽',[1184,440,664,44],30,colors.muted);
 for(let i=0;i<2;i++){const r=controls.modes[i],active=state.theme===(i?'hd':'ega');panel(ctx,r,{selected:active||state.menuCursor===i});label(ctx,i?'HD　手繪':'EGA　經典',[r[0]+28,r[1]+18,420,42],34);label(ctx,i?'介面、人物與旅店':'原作像素美術',[r[0]+28,r[1]+82,596,42],30,colors.muted);if(active)label(ctx,'使用中',[r[0]+510,r[1]+20,126,40],28,colors.gold,'center')}
 label(ctx,'同一場景，保留隊伍與操作位置',[1184,836,664,44],28,colors.muted);panel(ctx,controls.back,{selected:state.menuCursor===2});label(ctx,'回到遊戲　Esc',[1432,966,400,42],32,colors.ink,'center');
}
function render(){
 fits.length=0;const b=baselineCanvas.getContext('2d'),ctx=canvas.getContext('2d');b.imageSmoothingEnabled=false;b.drawImage(images[state.scene+(state.baseline==='ega'?'Ega':'Hd')],0,0,1920,1200);
 ctx.imageSmoothingEnabled=true;if(state.theme==='ega'){ctx.imageSmoothingEnabled=false;ctx.drawImage(images[state.scene+'Ega'],0,0,1920,1200)}else drawHD(ctx);settings(ctx);
 screens.classList.toggle('single',!state.compare);document.querySelector('#before-title').textContent=state.baseline==='ega'?'EGA 經典 · 相同狀態':'目前 HD 首批 · 相同狀態';document.querySelector('#after-title').textContent=state.theme==='ega'?'EGA 經典 · 切換預覽':'HD 操作介面候選 · 相同狀態';
 document.querySelectorAll('[data-scene]').forEach(e=>e.setAttribute('aria-pressed',String(e.dataset.scene===state.scene)));document.querySelector('#compare').setAttribute('aria-pressed',String(state.compare));document.querySelector('#baseline').textContent=state.baseline==='hd'?'對比：目前 HD／候選':'對比：EGA／HD';
 output.textContent=sceneLabels[state.scene]+' · '+(state.theme==='hd'?'HD 手繪':'EGA 經典')+' · '+(state.scene==='battle'?'Poontang':names[state.member])+' · '+commands[state.scene][state.command][1]+(lastAction?' · '+lastAction:'');
}
function scene(s){if(!commands[s])throw Error('未知場景');state.scene=s;state.command=0;state.menu=false;render()}
function gameIdentity(){return JSON.stringify({scene:state.scene,member:state.member,command:state.command})}
function theme(i){const before=gameIdentity();state.theme=i?'hd':'ega';state.menuCursor=i;render();if(gameIdentity()!==before)throw Error('切換改動場景或操作位置')}
function activate(h){if(!h)return;if(h.kind==='toggle'){state.menu=!state.menu;state.menuCursor=state.theme==='hd'?1:0}else if(h.kind==='back')state.menu=false;else if(h.kind==='theme'){theme(h.index);return}else if(h.kind==='member')state.member=h.index;else if(h.kind==='command'){state.command=h.index;lastAction='介面原型確認'}render()}
canvas.addEventListener('click',e=>{canvas.focus();activate(hit(...point(e.clientX,e.clientY)))});
canvas.addEventListener('contextmenu',e=>{e.preventDefault();if(state.menu){state.menu=false;render()}});
function key(e){
 if(e.key==='F10'){e.preventDefault();activate({kind:'toggle'});return}
 if(state.menu){if(e.key==='Escape'){state.menu=false;render()}else if(e.key==='ArrowDown'||e.key==='Tab'){e.preventDefault();state.menuCursor=(state.menuCursor+1)%3;render()}else if(e.key==='ArrowUp'){e.preventDefault();state.menuCursor=(state.menuCursor+2)%3;render()}else if(e.key==='Enter'){e.preventDefault();activate(state.menuCursor===2?{kind:'back'}:{kind:'theme',index:state.menuCursor})}return}
 if(/^[1-6]$/.test(e.key)&&state.scene!=='battle'){state.member=Number(e.key)-1;render()}else if([',','.','ArrowLeft','ArrowRight'].includes(e.key)){e.preventDefault();const n=commands[state.scene].length;state.command=(state.command+(['.','ArrowRight'].includes(e.key)?1:-1)+n)%n;render()}else if(e.key==='Enter'){e.preventDefault();lastAction='介面原型確認';render()}
}
document.addEventListener('keydown',key);document.querySelectorAll('[data-scene]').forEach(b=>b.onclick=()=>scene(b.dataset.scene));document.querySelector('#compare').onclick=()=>{state.compare=!state.compare;render()};document.querySelector('#baseline').onclick=()=>{state.baseline=state.baseline==='hd'?'ega':'hd';render()};
function test(){
 const saved={...state},ok=(name,pass)=>{checks.push({name,pass});if(!pass)throw Error(name)};
 for(const s of ['door','inn','battle']){state.scene=s;state.menu=false;state.command=0;state.theme='hd';render();ok(s+' 中文矩形',fits.every(f=>!f.truncated&&f.width<=f.rect[2]));const start=gameIdentity();state.menu=true;render();key({key:'1',preventDefault(){}});ok(s+' 設定攔截',start===gameIdentity());theme(0);theme(1);ok(s+' 主題往返',start===gameIdentity());key({key:'Escape',preventDefault(){}});ok(s+' 返回',!state.menu)}
 state.scene='inn';state.menu=false;for(const [w,h] of [[1920,1200],[960,600],[1280,800],[640,400]])for(let i=0;i<6;i++){const r=partyRects[i],p=point((r[0]+r[2]/2)*w/1920,(r[1]+r[3]/2)*h/1200,{left:0,top:0,width:w,height:h});const got=hit(...p);ok('隊員命中 '+w+' '+i,got&&got.kind==='member'&&got.index===i)}
 state.scene='battle';state.menu=false;for(let i=0;i<7;i++){const r=commandRects[i],got=hit(r[0]+r[2]/2,r[1]+r[3]/2);ok('指令命中 '+i,got&&got.kind==='command'&&got.index===i)}
 state.scene='inn';state.menu=false;state.member=0;key({key:'6',preventDefault(){}});ok('數字鍵選隊員',state.member===5);key({key:'F10',preventDefault(){}});ok('F10開啟',state.menu);key({key:'ArrowUp',preventDefault(){}});key({key:'Enter',preventDefault(){}});ok('鍵盤切EGA',state.theme==='ega');key({key:'ArrowDown',preventDefault(){}});key({key:'Enter',preventDefault(){}});ok('鍵盤切HD',state.theme==='hd');key({key:'F10',preventDefault(){}});ok('F10返回',!state.menu);
 Object.assign(state,saved);lastAction='';render();window.hdReview={state,scene,theme,hit,point,checks,fits,gameIdentity,render};document.body.dataset.tests='PASS';document.body.dataset.checkCount=String(checks.length);document.body.dataset.ready='true';
}
Promise.all([fetch(local('font-atlas.json')).then(r=>{if(!r.ok)throw Error('字模資料載入失敗');return r.json()}).then(d=>fontInfo=d),...Object.entries(sources).map(([key,url])=>new Promise((resolve,reject)=>{const img=new Image();img.onload=()=>{images[key]=img;resolve()};img.onerror=()=>reject(Error('圖片載入失敗：'+key));img.src=url}))]).then(()=>{render();test()}).catch(e=>{document.body.dataset.tests='FAIL';output.textContent=e.message;console.error(e)});
