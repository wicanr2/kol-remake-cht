"""只在 Docker 中核對正常原版快取觀察；完整平面資料仍留本機。"""
from pathlib import Path
import argparse, hashlib, importlib.util, json, shutil, sys

ROOT=Path('/src')
OUT=Path('/out')
sys.path.insert(0,str(ROOT/'tools'))
from png_io import read_png

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def load(p): return json.loads(p.read_text())
def column(call): return bytes.fromhex(call['packed_x15_four_planes_hex'])
def transform(before, stride):
    result=bytearray(before)
    for plane in range(4):
        start=plane*stride
        for y in range(156):
            offset=start+y*40+1
            actual_plane, local=divmod(offset,8000)
            row, byte=divmod(local,40)
            assert byte==1 and actual_plane<4 and row<200
            n=actual_plane*200+row
            result[n]=(result[n]|1) if plane<2 else (result[n]&254)
    return bytes(result)

def main():
    ap=argparse.ArgumentParser();ap.add_argument('--export',action='store_true');args=ap.parse_args()
    proof=Path('/proof') if args.export else ROOT/'docs/verification/item-detail-20261001'
    for name,digest in LOCKS.items(): assert sha(ROOT/name)==digest,(name,'鎖定輸入改變')
    previous=load(proof/'mark-receipt.json')
    for f in previous['current_inputs']: assert sha(ROOT/f['path'])==f['sha256']
    historical=[]
    for name in ['receipt.json','all-slots-receipt.json','actions-receipt.json','mirror-receipt.json','cjk-figure-receipt.json','mark-receipt.json']:
        receipt=load(proof/name)
        for f in receipt['attachments']: assert sha(proof/f['path'])==f['sha256']
        historical.append(dict(path=name,sha256=sha(proof/name),attachments=len(receipt['attachments'])))
    annotated=load(OUT/'item-edge-annotated-private.json')
    assert annotated['tool']=='IDA Pro 9.4' and annotated['address_space']=='IDA linear'
    assert annotated['sha256']==LOCKS['workplace/analysis/koltown/KOLTOWN-unpacked.EXE']
    assert annotated['database_sha256']==LOCKS['workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64']
    edge_annotations=[]
    for function in annotated['records']:
        assert function['original_name'].startswith('sub_')
        for row in function['instructions']:
            assert row['original_ea'] and row['operand_text'] and row['bytes']
            for a in row['annotations']:
                assert a['confidence']=='已證實' and a['evidence'] and sha(Path(a['index']))==a['index_sha256']
                if a['index'].endswith('/item-edge-address-index.json'): edge_annotations.append(row['original_ea'])
    assert len(edge_annotations)==9
    source=load(OUT/'gui-edge2/item-calls-private.json')
    old=load(OUT/'gui-actions1/item-calls-private.json')
    assert source['keys']==old['keys'] and len(source['keys'])==104
    assert source['states']==old['states']
    assert source['seed']==1 and source['normal_inputs']
    assert not source['state_injection'] and not source['direct_entry']
    for i in range(104):
        a=OUT/'gui-edge2'/f'item-{i:02d}.png'
        b=OUT/'gui-actions1'/f'item-{i:02d}.png'
        assert sha(a)==sha(b),(i,'原版畫格改變')
    calls=source['calls']
    first=next(c for c in calls if c['ida_ea']=='0x19E9D' and c['input_frame']==3
               and c['caller_ida']=='0x16A2A')
    marked=next(c for c in calls if c['ida_ea']=='0x16D0B' and c['input_frame']==3)
    assert first['args'][:5]==[first['buffer_segment'],0,0,40,200]
    assert first['target_segment']==marked['target_segment']==0xA000
    before,after=column(first),column(marked)
    assert len(before)==len(after)==800
    assert transform(before,8160)==after,'實際快取修改與原始指令不符'
    assert transform(before,8000)!=after,'修正過的平面步長反例未拒絕'
    rebuild_checks=[]
    for frame in [3,15,28,42,56,72,88]:
        copy=next(c for c in calls if c['input_frame']==frame and c['ida_ea']=='0x19E9D'
                  and c['caller_ida']=='0x16A2A')
        mark=next(c for c in calls if c['input_frame']==frame and c['ida_ea']=='0x16D0B')
        assert transform(column(copy),8160)==column(mark)
        assert transform(column(copy),8000)!=column(mark)
        rebuild_checks.append(dict(frame=frame,actual_stride=8160,plane_span=8000,
                                   cache_differences=0,corrected_stride_counterexample_rejected=True))
    changes=[dict(plane=n//200,row=n%200,before=before[n],after=after[n])
             for n in range(800) if before[n]!=after[n]]
    print('CALLS',len(calls),'CACHE_CHANGES',len(changes))
    palette=[(0,0,0),(0,0,170),(0,170,0),(0,170,170),
             (170,0,0),(170,0,170),(170,85,0),(170,170,170),
             (85,85,85),(85,85,255),(85,255,85),(85,255,255),
             (255,85,85),(255,85,255),(255,255,85),(255,255,255)]
    spec=importlib.util.spec_from_file_location('existing_figure',ROOT/'docs/verification/item-detail-20261001/verify_cjk_figure.py')
    figure=importlib.util.module_from_spec(spec);spec.loader.exec_module(figure)
    layout=load(ROOT/'assets/original-derived-data/portrait-layout.json')
    def colours(raw):
        return [palette[sum(((raw[p*200+y]&1)>0)<<p for p in range(4))] for y in range(200)]
    edge=None;model_frames=[]
    for frame in range(104):
        remake=read_png(OUT/'normal-mark-en'/f'step-{frame+6:03d}.png')
        rr=remake.rgb()
        for c in calls:
            if c['input_frame']!=frame: continue
            if c['ida_ea']=='0x19E9D' and c['caller_ida']=='0x16A2A':
                edge=colours(column(c))
                for y in range(2,157): edge[y]=palette[3]
            elif c['ida_ea']=='0x19E9D' and c['args'][1:5]==[1,164,38,196]:
                # 原版圖示列底板的實際複製矩形；它覆蓋快取的最下方四列。
                if edge is not None:
                    for y in range(164,196):
                        n=(y*320+15)*3;edge[y]=tuple(rr[n:n+3])
            elif c['ida_ea']=='0x16D0B':
                assert edge is not None
                old_row,new_row=c['args'][:2]
                cache=colours(column(c))
                if old_row!=10:
                    y0=layout['slot_marks'][3][old_row]
                    edge[y0:y0+26]=cache[y0:y0+26]
                if new_row!=10:
                    y0=layout['slot_marks'][3][new_row]
                    for y in range(y0,y0+26):
                        n=(y*320+15)*3;edge[y]=tuple(rr[n:n+3])
        if frame in figure.LIST:
            expected=bytearray(rr)
            for y in range(200):
                n=(y*320+15)*3;expected[n:n+3]=bytes(edge[y])
            original=read_png(OUT/'gui-edge2'/f'item-{frame:02d}.png').rgb()
            points=[]
            for n in range(320*200):
                x,y=n%320,n//320
                if 304<=x<318 and 185<=y<199: continue
                assert expected[n*3:n*3+3]==original[n*3:n*3+3],(frame,x,y,'續畫模型不符')
                if expected[n*3:n*3+3]!=rr[n*3:n*3+3]: points.append([x,y])
            model_frames.append(dict(frame=frame,model_pixel_differences=0,
                                     product_pending_pixels=points))
    assert len(model_frames)==78
    assert sum(bool(c['product_pending_pixels']) for c in model_frames)==70
    pixels=[]
    for frame, rows in [(4,[7]),(100,[0,1,157,158,159,161,162,163]),
                        (101,[0,1,157,158,159,161,162,163])]:
        cache=column(next(c for c in reversed(calls) if c['input_frame']<=frame and c['ida_ea']=='0x16D0B'))
        original=read_png(OUT/'gui-edge2'/f'item-{frame:02d}.png').rgb()
        remake=read_png(OUT/'normal-mark-en'/f'step-{frame+6:03d}.png').rgb()
        for y in rows:
            index=sum(((cache[p*200+y]&1)>0)<<p for p in range(4))
            n=(y*320+15)*3
            assert tuple(original[n:n+3])==palette[index],(frame,y,index)
            assert original[n:n+3]!=remake[n:n+3]
            pixels.append(dict(frame=frame,x=15,y=y,cache_palette_index=index,
                               original_rgb=list(original[n:n+3]),remake_rgb=list(remake[n:n+3])))
    report=dict(schema='kol-item-edge-probe-v1',status='正常原版快取窄證據通過，產品尚未修正',
                original_frames_unchanged=104,states_unchanged=104,
                original_revision='968ca774705853276ae55c8d1163c369481ac71f',
                original_seed=1,remake_seed=1,normal_startup=True,state_injection=False,direct_entry=False,
                cache_change_count=len(changes),observed_pixels=pixels,model_frames=model_frames,
                rebuild_checks=rebuild_checks,
                private_source_sha256=sha(OUT/'gui-edge2/item-calls-private.json'))
    report.update(date='2026-10-01',product_changed=False,product_pending_frames=70,
                  product_binary_sha256=LOCKS['workplace/item-detail-20261001/kol-items-mark-linux'],
                  current_inputs_unchanged=len(previous['current_inputs']),historical=historical,
                  source_inputs=[dict(path=k,sha256=v) for k,v in LOCKS.items()],
                  verifier_sha256=sha(Path(__file__)),confirmed_edge_addresses=edge_annotations,
                  negative_controls=['8000平面步長與正常快取不符，七次重建均拒絕'],
                  limits='只證實Theodoric正常104鍵的快取與曝光原因，產品尚未修正；原版無繁中畫面。')
    attachments=[]
    for frame in [4,15,100,101]:
        for role in ['original','remake-en']:
            source=(OUT/'gui-edge2'/f'item-{frame:02d}.png') if role=='original' else (OUT/'normal-mark-en'/f'step-{frame+6:03d}.png')
            filename=f'edge-source-{role}-{frame:03d}.png'
            if args.export:
                assert not (proof/filename).exists()
                shutil.copyfile(source,proof/filename)
            assert sha(proof/filename)==sha(source)
            attachments.append(dict(path=filename,sha256=sha(source),role=role,frame=frame,
                source_entry='KOLTOWN.EXE正常書本／旅店／物品頁' if role=='original' else 'kol-items-mark-linux正常輸入',
                source_sha256='c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221' if role=='original' else report['product_binary_sha256'],
                tool='dosgolem 968ca774705853276ae55c8d1163c369481ac71f' if role=='original' else 'Go 1.26.7／Ebitengine 2.9.9',
                evidence_status='已驗證PNG；來源原因已證實，產品差異仍未通過',
                rights='private GitHub可保存；不得公開散布'))
    report['attachments']=attachments
    destination=proof/'edge-source-receipt.json'
    if args.export:
        assert not destination.exists()
        destination.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    else: assert load(destination)==report,'來源收據重驗不一致'

    print('MODEL_FRAMES',len(model_frames),'PRODUCT_PENDING',70)
    print('正常原版畫格、角色狀態、快取步長及曝光像素均通過；產品未修正。')

LOCKS={'Knights-of-legend.zip': '2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46', 'workplace/analysis/koltown/KOLTOWN-unpacked.EXE': '01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2', 'workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64': '5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8', 'workplace/item-detail-20261001/kol-items-mark-linux': '821ee273ce3d78f658d4efa83b2716ec89a12e4002fa0365b7c8bf17fc86b78a', 'workplace/item-detail-20261001/gui-edge2/item-calls-private.json': '6291eb96a03be297fc6797e2bfd263e8dfe17b61a7474f51f3acc82572e5809a', 'workplace/item-detail-20261001/gui-edge2/inputs.json': 'b2ae9bd508710131f30c2e686d20dd9284d5aaa95ddfc9cc9fe5c8bd1c017e41', 'workplace/item-detail-20261001/gui-edge2/prepare.py': '6858d52a4bce5eb68a162e6ae769aed8195cc411a319e0855eb15ebbe1433b30', 'workplace/item-detail-20261001/gui-edge2/original.sh': '76c1ba9250e34ee4b0d809449edc80ed85979f4c631eafbe86b22ba20d2a65dd', 'workplace/item-detail-20261001/gui-edge2/instrumented-book-startup-private.go': 'f454dd9d3d3b4c636120b921ed6895e5e081987569dab643920c470c96fc8703', 'workplace/item-detail-20261001/item-edge-annotated-private.json': '819531e117275aca2071b66e6799d1a9fa0ba55e9c1a52ffbeca7ee95f7da304', 'assets/original-derived-data/portrait-layout.json': '23ff63ee3bf17be432cb17fefeb308f31c838f3e8fb796f164ea0186fc30d8df', 'docs/re/item-edge-address-index.json': 'daf589b21d26b3d22086f1256b4d1c3a705096c70394deb0c9771bb16ac2ad0d', 'docs/verification/item-detail-20261001/mark-receipt.json': '75616fa5be90046faff4e213412b57e4bee0a1a3514fed7e50fcda0f997b78e8', 'docs/verification/item-detail-20261001/verify_cjk_figure.py': '430e412ba61e08ebc2f03c72acc79d5943f530caef4666e80ab96b3acfbd2b06', 'tools/png_io.py': '0295c62847b9e66e9a997940224c1d4c04bf5e0acbeab2d820bfecd0fd020c9c'}

if __name__=='__main__': main()
