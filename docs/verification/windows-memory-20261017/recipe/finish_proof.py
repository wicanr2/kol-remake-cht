from pathlib import Path
import hashlib,json,re,shutil,struct,wave
import numpy as np

root=Path('/src'); work=Path('/out'); dist=root/'dist-all/v.1.0.4-20261017'
formal=root/'docs/verification/windows-memory-20261017'
def read(p): return json.loads(p.read_text())
def sha(p): return hashlib.file_digest(p.open('rb'),'sha256').hexdigest()
def write(p,v): p.write_text(json.dumps(v,ensure_ascii=False,indent=2)+'\n')
runtime=read(work/'wine-runtime.json'); captures=read(work/'captures-comparison.json')
assert runtime['status']==captures['status']=='PASS'
assert len(runtime['runs'])==4 and all(r['exit_code']==0 and r['f10_state_unchanged'] for r in runtime['runs'])
assert runtime['default_logging'] and runtime['corrupt_hd_returns_to_ega']
events=dict(re.findall(r'^(\w+) (\d+)$',runtime['memory_events'],re.M))
assert events['oom']==events['oom_kill']=='0'
assert 'FAIL' not in (work/'build-delivery.log').read_text()
assert re.search(r'^ok\s+github.com/wicanr2/kol-remake-cht/internal/assets\s+', (work/'assets-recheck.log').read_text(),re.M)
inputs=read(work/'source-inputs.json')
assert len(inputs)==435 and all(sha(root/n)==h for n,h in inputs.items())
assert read(dist/'build-inputs.json')['program_inputs']==inputs
packages=read(dist/'SHA256SUMS.json')['files']
assert len(packages)==4
for p in packages: assert sha(dist/p['path'])==p['sha256'] and (dist/p['path']).stat().st_size==p['bytes']
exe=(work/'kol.exe').read_bytes(); pos=struct.unpack_from('<I',exe,0x3c)[0]
assert exe[:2]==b'MZ' and exe[pos:pos+4]==b'PE\0\0' and struct.unpack_from('<H',exe,pos+4)[0]==0x8664
# Actual packaged flow captures must agree across full and imported engine.
flow=[]
for theme in ['ega','ega-hd']:
    left=work/('full-'+theme+'-flow'); right=work/('release-'+theme+'-flow')
    files=sorted(p.name for p in left.glob('*.png'))
    assert len(files)==6 and files==sorted(p.name for p in right.glob('*.png'))
    from PIL import Image
    for name in files:
        a=np.asarray(Image.open(left/name).convert('RGBA'),dtype=np.int16)
        b=np.asarray(Image.open(right/name).convert('RGBA'),dtype=np.int16)
        limit=0 if theme=='ega' else 1
        delta=int(np.abs(a-b).max()); assert delta<=limit,(theme,name,delta)
        flow.append(dict(theme=theme,name=name,max_channel_delta=delta,limit=limit))
    assert read(work/('full-'+theme+'-trace.json'))==read(work/('release-'+theme+'-trace.json'))
    assert (work/('full-'+theme+'-text.jsonl')).read_bytes()==(work/('release-'+theme+'-text.jsonl')).read_bytes()
write(work/'package-flow-comparison.json',dict(status='PASS',comparisons=flow,models_and_chinese_equal=True))
# Independent audio device, captured from actual Wine/Oto playback.
audio=work/'smoke/windows-audio.wav'
peak=nonzero=count=0
with wave.open(str(audio),'rb') as stream:
    assert (stream.getnchannels(),stream.getframerate(),stream.getsampwidth())==(2,44100,2)
    seconds=stream.getnframes()/44100
    tail=b''
    while chunk:=stream.readframes(65536):
        samples=np.frombuffer(chunk,dtype='<i2').astype(np.int32)
        peak=max(peak,int(np.abs(samples).max())); nonzero+=int(np.count_nonzero(samples)); count+=len(samples)
        tail=(tail+chunk)[-44100*4:]
assert 0<peak<32767 and nonzero>0 and not np.any(np.frombuffer(tail,dtype='<i2'))
write(work/'audio-runtime.json',dict(status='PASS',seconds=seconds,peak=peak,nonzero_samples=nonzero,samples=count,
    tail_one_second_silent=True,sha256=sha(audio),human_listening_repeated=False))
tests=(work/'tests.log').read_text()
unchanged=[p for p in re.findall(r'^ok\s+(github.com/wicanr2/kol-remake-cht/\S+)\s+',tests,re.M)
           if not p.endswith(('/cmd/kol','/internal/ui','/internal/layers','/internal/assets'))]
for name in ['source-inputs.json','post-scene-notice-fix.json','wine-runtime.json','captures-comparison.json',
             'package-flow-comparison.json','audio-runtime.json','build-delivery.log','assets-recheck.log']:
    shutil.copyfile(work/name,formal/name)
for name in ['build.sh','build-scenes.sh','verify_windows.py','check_captures.py','finish_proof.py']:
    shutil.copyfile(work/name,formal/'recipe'/name)
attachments={n:sha(work/n) for n in ['source-inputs.json','post-scene-notice-fix.json','wine-runtime.json',
    'captures-comparison.json','package-flow-comparison.json','audio-runtime.json','build-delivery.log','assets-recheck.log',
    'tests.log','scenes-linux/motion-captures.json','scenes-windows/motion-captures.json','scenes-linux/source-inputs.json']}
receipt=dict(schema='kol-windows-memory/1',status='PASS',verified_on='2026-10-17',version='v.1.0.4-20261017',
    tools=dict(go='1.26.7',ebitengine='2.9.9',wine='9.0',images={
        'kol-package:1.26.7-ebiten2.9.9-20261011-r1':'sha256:0b933af1238ab5839ad7a64b6ae05f069f57cb85da5917dfbe59730f6d7ba53f',
        'eob-remake-wine-verify:ubuntu-noble-20261008-r2':'sha256:c047a6a259a1fc9a90babf7c937310fa9df4b3465bf75f3fcaad1bc640099c2f',
        'eob-audio-capture:20261009-r3':'sha256:d831720f8dfa388af8ad175fa8adeae8f2814dd4637be1e3db6a43274b26f36c'}),
    remake_seed=1,original_seed=None,
    program_inputs=inputs,program_files=435,windows_binary_sha256=sha(work/'kol.exe'),packages=packages,
    runtime=runtime,controlled_scenes=13,frames_per_platform=52,
    cross_platform_max=max(x['max_channel_delta'] for x in captures['cross_platform']),
    previous_release_max=max(x['max_channel_delta'] for x in captures['previous_release']),
    ega_roundtrip_max=max(x['max_channel_delta'] for x in captures['roundtrips'] if x['theme']=='ega'),
    hd_roundtrip_max=max(x['max_channel_delta'] for x in captures['roundtrips'] if x['theme']=='hd'),
    missing_glyphs=[],normal_intro_verified=True,opening_crash_originally_reproduced=False,
    controlled_scenes_original_oracle=False,natural_all_quests_completion=False,real_windows_machine=False,
    no_validation_only_heap_override=True,texture_budget_bytes=192<<20,sample_budget_bytes=96<<20,
    tests=dict(final_changed_packages=['cmd/kol','internal/ui','internal/layers','internal/assets/TestDeferredHD'],
        final_log='build-delivery.log',unchanged_core_packages=unchanged,
        offline_asset_suite='assets-recheck.log; 796s after first 600s timeout; same offline validators before provider forwarding method',
        failed_initial_run_retained='tests.log',staged_testing=True),
    post_scene_delta=read(work/'post-scene-notice-fix.json'),
    publication_scope='Two asset-free Windows engine ZIPs and SHA256SUMS.txt; full builds local only',
    other_platforms_and_promo='v.1.0.3-20261017 unchanged',attachments=attachments)
write(work/'receipt.json',receipt); write(formal/'receipt.json',receipt)
print('FINAL PROOF PASS: four Windows packages, actual controls, 104 controlled frames, 24 flow frames, audio, corrupt-HD fallback')
