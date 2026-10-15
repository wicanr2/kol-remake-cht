// DRAFT expansion of the approved copper Canvas icon family.
// Each source ID keeps its original object, pose or direction. No action rules.
const G='#d6b77a', S='#d9dde0', R='#9c3540', B='#486982', K='#30272b', SK='#c3a27a';
function disk(c,x,y,r,fill,stroke=G,w=2){c.beginPath();c.arc(x,y,r,0,Math.PI*2);if(fill){c.fillStyle=fill;c.fill()}if(stroke){c.strokeStyle=stroke;c.lineWidth=w;c.stroke()}}
function line(c,p,color=G,w=5){c.beginPath();c.moveTo(...p[0]);for(let i=1;i<p.length;i++)c.lineTo(...p[i]);c.strokeStyle=color;c.lineWidth=w;c.stroke()}
function arr(c,x,y,angle=0,len=46,color=G,width=6){c.save();c.translate(x,y);c.rotate(angle*Math.PI/180);line(c,[[0,len/2],[0,-len/2]],color,width);line(c,[[-12,-len/2+13],[0,-len/2],[12,-len/2+13]],color,width);c.restore()}
function blade(c,x=50,y=50,angle=0,kind='sword',scale=1){
 c.save();c.translate(x,y);c.rotate(angle*Math.PI/180);c.scale(scale,scale);
 if(kind==='axe'){path(c,'M-3 34L0 -35L5 -35L3 34Z','#6d4332',G,2);path(c,'M0 -30Q-25 -43 -25 -10Q-12 -5 0 -15Q12 -5 25 -10Q25 -43 0 -30Z',S,G,2)}
 else{path(c,'M-4 25L-5 -24L0 -38L5 -24L4 25Z',S,'#f4e6cc',2);line(c,[[-12,18],[12,18]],G,5);line(c,[[0,23],[0,37]],'#75402e',6);disk(c,0,38,4,G,G,1)}c.restore()
}
function shield(c,x=50,y=50,scale=1){c.save();c.translate(x,y);c.scale(scale,scale);path(c,'M0 -36L27 -27V-4Q24 23 0 36Q-24 23 -27 -4V-27Z',B,G,3);line(c,[[0,-29],[0,26]],G,2);c.restore()}
function wings(c,x,y,spread=1){c.save();c.translate(x,y);c.scale(spread,1);path(c,'M-4 0Q-34 -38 -44 -27L-39 5L-29 -2L-23 19L-15 5L-4 17M4 0Q34 -38 44 -27L39 5L29 -2L23 19L15 5L4 17',R,G,2);c.restore()}
function person(c,points,cloth=B,opts={}){
 const [head,neck,hip,el,handl,er,handr,kneel,footl,kneer,footr]=points;
 if(opts.wings)wings(c,neck[0],neck[1]+8,opts.wings);
 line(c,[hip,kneel,footl],cloth,11);line(c,[hip,kneer,footr],cloth,11);
 line(c,[neck,el,handl],SK,8);line(c,[neck,er,handr],SK,8);
 line(c,[neck,hip],cloth,18);line(c,[[hip[0]-8,hip[1]-1],[hip[0]+8,hip[1]-1]],G,3);
 disk(c,...head,8,SK,G,1.5);path(c,`M${head[0]-8} ${head[1]-2}Q${head[0]} ${head[1]-14} ${head[0]+8} ${head[1]-2}`,opts.hair||R,null);
 disk(c,...handl,3.2,SK,G,1);disk(c,...handr,3.2,SK,G,1);
 if(opts.shield)shield(c,...handl,.4);
 if(opts.blade)blade(c,handr[0],handr[1]-13,opts.blade,'sword',.56);
}
const standing=[[50,17],[50,31],[50,59],[34,44],[26,61],[67,44],[74,61],[40,75],[30,91],[61,75],[70,91]];
const walk=[[55,17],[52,32],[48,59],[36,43],[31,58],[69,43],[74,29],[37,73],[25,90],[62,72],[72,85]];
const run=[[61,16],[55,30],[45,55],[35,39],[29,25],[72,39],[79,54],[27,69],[14,89],[62,69],[82,74]];
const sprint=[[68,25],[61,36],[43,54],[43,36],[29,28],[69,49],[81,48],[27,67],[11,81],[58,73],[78,87]];
const crouch=[[53,39],[51,53],[47,72],[32,62],[20,57],[68,57],[79,45],[28,77],[16,90],[65,81],[80,90]];
const dodge=[[67,22],[59,36],[45,61],[42,37],[29,26],[75,43],[85,60],[30,73],[20,92],[62,71],[80,80]];
const kick=[[53,15],[52,30],[49,54],[36,40],[23,35],[70,39],[82,29],[65,59],[88,51],[38,73],[28,91]];
const punch=[[51,16],[50,30],[47,58],[35,45],[27,29],[69,35],[87,34],[33,74],[23,91],[61,75],[78,91]];
const flip=[[50,80],[50,66],[48,42],[33,56],[24,38],[67,55],[79,36],[28,24],[13,18],[67,23],[82,9]];
const flying=[[78,40],[65,46],[42,56],[57,31],[72,23],[72,58],[91,56],[26,49],[10,38],[28,70],[12,83]];
function motion(c,kind){if(kind===0)path(c,'M22 78Q8 24 57 14',null,G,3);if(kind===1)path(c,'M18 24Q74 8 87 64',null,G,3);if(kind===2)path(c,'M19 80Q64 96 87 53',null,G,3)}
function bag(c,x=50,y=55){path(c,`M${x-21} ${y-22}Q${x} ${y-33} ${x+21} ${y-22}L${x+25} ${y+31}Q${x} ${y+41} ${x-25} ${y+31}Z`,'#704b32',G,3);line(c,[[x-18,y-16],[x+18,y-16]],G,4);path(c,`M${x-13} ${y+4}H${x+13}V${y+22}H${x-13}Z`,null,G,2)}
function spark(c,x=52,y=42){path(c,`M${x} ${y-27}L${x+6} ${y-7}L${x+28} ${y}L${x+6} ${y+6}L${x} ${y+27}L${x-6} ${y+6}L${x-28} ${y}L${x-6} ${y-6}Z`,G,'#f6dfaa',2)}
function swapSword(c,variant){blade(c,64,49,0);arr(c,23,35,variant===0?0:180,36);if(variant===2)line(c,[[14,81],[37,81]],S,4);if(variant===3){disk(c,82,83,9,SK,G,2);line(c,[[77,76],[77,67]],SK,4)}}
function drawSourceAction(c,id){
 c.save();c.lineCap='round';c.lineJoin='round';
 switch(id){
 case 3:path(c,'M12 14H88V38H12Z',S,G,3);path(c,'M39 57Q39 44 52 44Q69 45 62 59L50 70V76',null,G,5);disk(c,50,88,3,G,null);break;
 case 4:path(c,'M10 21H90V69H10Z',K,G,4);line(c,[[18,83],[82,83]],G,4);c.fillStyle=G;c.font='26px "Noto Serif CJK TC"';c.textAlign='center';c.fillText('空白',50,55);break;
 case 5:icon(c,'door',0,0,100);break;
 case 6:path(c,'M10 48Q30 10 50 31Q70 10 90 48Q73 83 50 72Q27 83 10 48Z',R,G,3);path(c,'M17 48Q50 36 83 48Q50 63 17 48Z',K,'#e5ba9d',2);break;
 case 7:path(c,'M69 83Q33 91 22 51Q11 9 50 9Q87 10 78 42Q76 53 62 55L51 74Q41 84 35 69',SK,G,3);path(c,'M57 67Q35 62 35 38Q40 17 59 26Q76 32 62 44L49 46',null,'#69402f',4);break;
 case 8:icon(c,'rest',0,0,100);break;
 case 9:person(c,standing,B);break;
 case 10:icon(c,'yes',0,0,100);break;
 case 11:c.translate(100,100);c.rotate(Math.PI);icon(c,'yes',0,0,100);break;
 case 12:path(c,'M12 65L50 78L88 65L80 86H20Z',S,G,3);path(c,'M28 43L38 22H66L77 43L51 68Z','#466981',G,3);line(c,[[28,43],[77,43],[51,68],[38,22],[28,43],[66,22],[51,68]],'#c9e0e7',2);break;
 case 13:path(c,'M9 42H86L73 57L59 61V80H76V89H25V80H40V59L19 55Z',S,G,3);path(c,'M23 18L28 8L57 20L52 31Z','#796146',G,3);line(c,[[48,25],[66,43]],G,5);break;
 case 14:path(c,'M41 10H62V26L72 40V87H29V40L41 26Z','#536a69',G,3);path(c,'M36 49H65V79H36Z',R,null);line(c,[[51,53],[51,75]],S,5);line(c,[[41,64],[61,64]],S,5);disk(c,83,73,8,G,S,1);disk(c,83,88,8,G,S,1);break;
 case 15:blade(c,42,48,-42);blade(c,59,48,42);break;
 case 16:icon(c,'fist',4,23,72);spark(c,59,29);break;
 case 17:blade(c,57,46,0);icon(c,'fist',10,29,77);break;
 case 18:icon(c,'back',0,0,100);break;
 case 19:path(c,'M16 17L34 11L43 28L39 72H15L10 29Z',B,G,3);path(c,'M68 14L87 20L90 73H64V26Z',R,G,3);arr(c,54,80,270,47);break;
 case 20:path(c,'M17 9H76V78H17Z',K,G,5);person(c,[[47,27],[47,39],[47,60],[36,44],[30,57],[62,47],[66,60],[39,70],[34,78],[55,70],[62,78]],B);line(c,[[81,30],[81,88]],'#75422c',7);break;
 case 21:path(c,'M26 13H83Q91 22 81 29H75V77Q92 73 88 89H20Q8 80 20 74V29Q6 35 10 14Z','#dfd3b5',G,3);line(c,[[28,33],[66,33]],'#6d5141',3);line(c,[[28,46],[60,46]],'#6d5141',3);line(c,[[28,59],[65,59]],'#6d5141',3);break;
 case 22:path(c,'M31 87V60Q14 42 26 23Q35 7 57 10Q82 16 78 40L90 49L77 55V66H62V87Z',SK,G,3);path(c,'M38 21Q64 12 67 36Q65 49 48 42Q40 35 54 29',null,G,4);break;
 case 23:bag(c);arr(c,50,20,0,28);break;
 case 24:icon(c,'drop',0,0,100);break;
 case 25:swapSword(c,0);break;
 case 26:path(c,'M11 17L42 10L64 20L90 12V83L65 91L43 80L11 88Z','#475a43',G,3);path(c,'M15 73L38 39L55 58L72 36L88 75Z','#9c7f52',G,2);disk(c,71,25,8,'#e6c272',null);break;
 case 27:line(c,[[49,13],[49,92]],'#785034',8);path(c,'M15 20H78L88 31L78 42H15Z',S,G,2);path(c,'M22 49H88V71H22L11 60Z','#c9c0aa',G,2);break;
 case 31:swapSword(c,1);break;
 case 36:icon(c,'sword',0,0,100);break;
 case 37:path(c,'M56 39H83L78 91H61Z','#79523a',G,3);line(c,[[59,42],[83,42]],G,5);line(c,[[68,43],[63,13]],S,3);line(c,[[76,43],[78,14]],S,3);arr(c,29,35,180,42);break;
 case 38:path(c,'M25 10Q84 50 25 90L43 50Z','#785438',G,4);line(c,[[14,50],[89,50]],S,4);line(c,[[77,42],[89,50],[77,58]],G,4);break;
 case 39:swapSword(c,2);break;
 case 40:swapSword(c,3);break;
 case 41:path(c,'M20 83L48 70L79 83L51 95Z','#6c4031',G,3);arr(c,50,38,0,46);break;
 case 42:icon(c,'move',0,0,100);break;
 case 43:person(c,standing,B,{wings:1});arr(c,13,70,180,37);arr(c,87,70,180,37);break;
 case 44:icon(c,'fist',0,0,100);break;
 case 45:person(c,[[51,16],[50,31],[50,58],[31,42],[16,32],[70,40],[81,26],[38,74],[29,92],[62,75],[73,92]],R);blade(c,83,28,15,'sword',.75);motion(c,1);break;
 case 46:person(c,dodge,R);blade(c,38,26,-50,'sword',.75);motion(c,0);break;
 case 47:path(c,'M22 89Q35 44 84 18Q85 50 46 80Z',S,G,3);path(c,'M15 28Q60 2 89 48',null,B,4);line(c,[[21,73],[43,89]],G,6);break;
 case 48:blade(c,51,55,90);path(c,'M13 24Q53 8 90 38',null,B,4);break;
 case 49:blade(c,50,53,40);path(c,'M15 19Q26 49 9 74',null,B,4);break;
 case 50:person(c,standing,R,{shield:false});arr(c,15,45,90,27);break;
 case 51:person(c,crouch,'#91733f',{shield:true});break;
 case 52:person(c,standing,B,{shield:true});break;
 case 53:person(c,dodge,R,{shield:true});break;
 case 54:person(c,crouch,B);line(c,[[16,94],[84,94]],G,3);break;
 case 55:person(c,[[51,14],[50,28],[50,55],[28,36],[12,27],[73,34],[90,26],[35,72],[17,91],[65,71],[83,91]],S,{hair:R});motion(c,2);break;
 case 56:person(c,flying,R,{wings:.9});line(c,[[12,20],[37,20]],G,3);line(c,[[6,31],[25,31]],G,3);break;
 case 57:person(c,[[51,23],[50,37],[50,61],[28,35],[18,16],[72,34],[83,15],[37,76],[27,93],[64,76],[74,93]],R,{shield:false});shield(c,17,26,.35);break;
 case 58:person(c,kick,'#496f4e',{hair:G});break;
 case 59:person(c,punch,B,{hair:S});break;
 case 60:person(c,[[75,42],[62,43],[42,60],[65,63],[81,74],[60,25],[72,14],[28,47],[10,33],[29,78],[13,93]],R);break;
 case 61:person(c,flip,R);motion(c,0);break;
 case 62:person(c,dodge,B,{hair:R});motion(c,2);break;
 case 63:person(c,walk,B);break;
 case 64:person(c,run,B);break;
 case 65:person(c,sprint,B);line(c,[[9,17],[33,17]],G,3);line(c,[[6,28],[22,28]],G,3);break;
 case 66:person(c,flying,R,{wings:1});break;
 case 67:person(c,[[82,40],[67,43],[44,53],[58,29],[76,20],[72,57],[93,59],[27,43],[8,30],[27,67],[8,79]],R,{wings:.85});line(c,[[7,14],[40,14]],G,4);break;
 case 68:person(c,standing,R);arr(c,16,50,90,34);break;
 case 69:person(c,standing,R);arr(c,18,80,45,38);break;
 case 74:case 75:case 76:case 77:case 78:case 79:case 80:case 81:arr(c,50,50,(id-74)*45,68,G,9);break;
 default:throw Error('Unmapped source action '+id);
 }
 c.restore();
}
function renderActionAtlas(ids){
 const canvas=document.createElement('canvas');canvas.width=144*9;canvas.height=120*Math.ceil(ids.length/9);const ctx=canvas.getContext('2d');
 ids.forEach((id,i)=>{
  const cell=document.createElement('canvas');cell.width=192;cell.height=168;const c=cell.getContext('2d');
  c.translate(36,29);c.scale(1.1,1.1);drawSourceAction(c,id);
  const raw=c.getImageData(0,0,192,168).data;let x0=192,y0=168,x1=0,y1=0;
  for(let y=0;y<168;y++)for(let x=0;x<192;x++)if(raw[(y*192+x)*4+3]>=8){x0=Math.min(x0,x);y0=Math.min(y0,y);x1=Math.max(x1,x+1);y1=Math.max(y1,y+1)}
  if(x0<=24||y0<=24||x1>=168||y1>=144)throw Error('Action '+id+' geometry leaves its cell: '+[x0,y0,x1,y1]);
  ctx.drawImage(cell,24,24,144,120,(i%9)*144,Math.floor(i/9)*120,144,120);
 });return canvas;
}
