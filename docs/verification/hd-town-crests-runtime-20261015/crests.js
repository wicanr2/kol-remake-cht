/* 原生 Canvas 紋章續作。框、尺寸與色盤沿用已核准 crest()；圖案依 TWN1…6PICS0。 */
const palette={paper:'#d9d1bb',gold:'#dbbd70',edge:'#c4a976',red:'#722631',blue:'#283d54',ink:'#1e2433',silver:'#c6c7c0',light:'#eee1ca',green:'#426a45',brown:'#805437'};
const shield='M10 10H310V220Q275 270 160 313Q45 270 10 220Z';
const inner='M24 24H296V214Q260 253 160 294Q60 253 24 214Z';
function path(c,d,fill,stroke,width=3){const p=new Path2D(d);if(fill){c.fillStyle=fill;c.fill(p)}if(stroke){c.strokeStyle=stroke;c.lineWidth=width;c.lineJoin='miter';c.stroke(p)}}
function rect(c,x,y,w,h,fill){c.fillStyle=fill;c.fillRect(x,y,w,h)}
function circle(c,x,y,r,fill){c.fillStyle=fill;c.beginPath();c.arc(x,y,r,0,Math.PI*2);c.fill()}
function tower(c,x,y,s=1){c.save();c.translate(x,y);c.scale(s,s);
 path(c,'M0 16H64V94H0ZM-7 0H9V21H-7ZM17 0H33V21H17ZM41 0H71V21H41Z',palette.silver,palette.light,3);
 path(c,'M5 25H59M5 48H59M5 71H59M20 25V48M43 48V71M20 71V91',null,'#999d9d',2);
 path(c,'M25 94V64Q32 54 39 64V94Z',palette.ink);c.restore()}
function sun(c,x,y,r){circle(c,x,y,r,palette.gold);for(let i=0;i<12;i++){c.save();c.translate(x,y);c.rotate(i*Math.PI/6);path(c,`M0 ${-r-8}V${-r-19}`,null,palette.gold,4);c.restore()}}
function stripes(c,box,diagonal){c.save();c.beginPath();c.rect(...box);c.clip();rect(c,...box,diagonal?palette.blue:palette.red);
 if(diagonal){for(let x=-300;x<500;x+=40)path(c,`M${x} 320L${x+320} 0L${x+336} 0L${x+16} 320Z`,palette.paper)}
 else{for(let x=box[0]+14;x<box[0]+box[2];x+=40)rect(c,x,box[1],20,box[3],palette.gold)}c.restore()}
function quarter(c,n){rect(c,24,24,272,272,palette.ink);stripes(c,[164,24,132,132],n===3);stripes(c,[24,164,132,132],n===3);
 if(n===1)sun(c,89,88,30);else{path(c,'M83 121V88H97V121L114 128H67Z',palette.brown,palette.gold,2);circle(c,70,78,28,palette.green);circle(c,96,60,34,palette.green);circle(c,119,82,27,palette.green);circle(c,93,87,34,palette.green);path(c,'M55 82Q71 47 101 45M75 101Q102 117 130 88',null,'#6d8e58',3)}
 tower(c,198,177,.89)}
function ship(c){rect(c,24,24,272,272,palette.paper);
 for(let i=0;i<4;i++){const y=187+i*34;path(c,`M18 ${y}Q42 ${y-12} 65 ${y}T111 ${y}T157 ${y}T203 ${y}T249 ${y}T305 ${y}V${y+23}Q280 ${y+11} 257 ${y+23}T211 ${y+23}T165 ${y+23}T119 ${y+23}T73 ${y+23}T27 ${y+23}Z`,palette.blue)}
 path(c,'M62 151H247L229 170Q190 183 106 178L82 169Z',palette.ink);
 path(c,'M152 43V151M105 66V153M214 73V151M81 50H200M92 68H126M200 75H237',null,palette.ink,5);
 path(c,'M119 54H192L184 119H118Q132 86 119 54Z',palette.light,palette.ink,4);
 path(c,'M93 76H122L117 125H86Q100 102 93 76ZM210 82H239Q259 102 241 125H207Z',palette.silver,palette.ink,3);
 path(c,'M73 148L152 44L233 149M107 67L83 147',null,palette.ink,2);
 path(c,'M161 41V29L190 35L161 41Z',palette.red);path(c,'M86 158H226',null,palette.silver,2)}
function bridge(c){rect(c,24,24,272,272,palette.ink);
 path(c,'M181 45Q130 33 123 76Q122 116 166 118Q134 136 108 110Q82 80 109 50Q141 22 181 45Z',palette.silver);
 rect(c,24,192,272,115,palette.blue);
 path(c,'M27 170L55 153H265L293 170V206H270V193Q257 169 240 192V206H193V190Q161 160 129 190V206H82V193Q65 169 50 193V206H27Z',palette.silver,palette.light,3);
 path(c,'M54 152V143H264V152M40 176H280M105 158V181M215 158V181',null,palette.light,3);
 for(const [x,y,w] of [[43,218,38],[116,214,86],[235,218,42],[57,231,24],[122,230,74],[132,246,55],[144,259,30]])rect(c,x,y,w,5,palette.ink)}
function castle(c){rect(c,24,24,272,272,palette.ink);
 path(c,'M160 181L276 287H44Z',palette.brown);path(c,'M160 186V298',null,'#a77b48',3);
 path(c,'M90 101H230V194H90Z',palette.silver,palette.light,3);
 path(c,'M93 91H109V111H93ZM124 91H140V111H124ZM156 91H172V111H156ZM188 91H204V111H188ZM219 91H231V111H219Z',palette.silver,palette.light,2);
 tower(c,53,48,1.1);tower(c,197,48,1.1);
 path(c,'M130 194V158Q160 122 190 158V194Z',palette.ink,palette.light,4);
 path(c,'M94 122H225M94 142H128M194 142H226M100 163H121M198 163H222',null,'#999d9d',2)}
function dragon(c){rect(c,24,24,272,132,palette.red);rect(c,24,164,272,132,palette.ink);
 path(c,'M51 88L76 80L73 64L90 68L101 52L111 62L130 46L147 48L169 70L188 66L205 82L235 76L252 60L268 60L258 77L239 92L218 102L216 118L233 130L229 141L204 146L210 135L197 127L185 126L181 140L158 143L166 132L159 120L144 111L123 114L111 133L91 137L99 123L94 110L81 107L75 98L54 98L47 92L33 92L33 85Z',palette.ink);
 path(c,'M116 90L112 61L131 41L144 42L170 70L165 99Z',palette.ink);
 path(c,'M203 69L216 58L231 58L227 46L245 45L256 51L244 57L241 64L220 68Z',palette.ink);
 circle(c,74,87,2,palette.gold);
 path(c,'M157 281V194M157 261L111 228V190L91 211V246M157 261L203 228V190L223 211V246M157 237L141 222V208M157 237L173 222V208M157 213L146 198M157 213L170 197',null,palette.gold,8)}
function drawCrest(c,n){c.clearRect(0,0,320,320);c.save();path(c,shield,palette.paper);c.clip(new Path2D(inner));
 if(n===1||n===3)quarter(c,n);else if(n===2)ship(c);else if(n===4)bridge(c);else if(n===5)castle(c);else if(n===6)dragon(c);else throw Error('未知來源');
 c.restore();path(c,shield,null,palette.edge,6)}
const exports={};
for(let n=1;n<=6;n++){const c=document.querySelector(`#crest-${n}`);drawCrest(c.getContext('2d'),n);exports[`twn${n}pics-000.png`]=c.toDataURL('image/png')}
document.querySelector('#exports').textContent=JSON.stringify(exports);
