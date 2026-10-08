#!/usr/bin/env python3
"""在鎖定音訊 Docker image 重建標題 OGG，不覆寫已驗收資產。"""
import argparse, hashlib, json, pathlib, subprocess, sys

def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--out',required=True)
    args=p.parse_args()
    if not pathlib.Path('/.dockerenv').is_file():p.error('必須在 Docker 中執行')
    root=pathlib.Path(__file__).resolve().parent.parent
    version=subprocess.run(['ffmpeg','-version'],check=True,capture_output=True,text=True,timeout=10).stdout.splitlines()[0]
    if not version.startswith('ffmpeg version 5.1.9-'):p.error('請用鎖定的eob-remake-audio:20260901工具鏈')
    out=pathlib.Path(args.out);out.mkdir(parents=True,exist_ok=False)
    def run(command):
        return subprocess.run(command,check=True,capture_output=True,text=True,timeout=30).stdout
    wav=out/'title-thunder.wav';ogg=out/'title-thunder.ogg'
    generator=[sys.executable,str(root/'tools/synthesize_title_thunder.py'),'--out',str(wav),'--steps-out',str(out/'steps.json'),'--seed','1','--sample-rate','44100','--cue-seconds','0.961','--peak-db','-10.317424']
    run(generator)
    encoder=['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-i',str(wav),'-c:a','libvorbis','-q:a','5','-fflags','+bitexact','-serial_offset','2008966748',str(ogg)]
    run(encoder)
    sha=lambda path:hashlib.sha256(path.read_bytes()).hexdigest()
    accepted=root/'assets/audio/title-thunder.ogg'
    hashes=[]
    for name,source in [('accepted',accepted),('rebuilt',ogg)]:
        target=out/(name+'.f32')
        run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-i',str(source),'-f','f32le','-acodec','pcm_f32le',str(target)])
        hashes.append(sha(target))
    if hashes[0]!=hashes[1]:raise RuntimeError('重建音訊與已驗收候選解碼不同')
    meta=json.loads(run(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(ogg)]))
    stream=meta['streams'][0]
    assert stream['codec_name']=='vorbis' and stream['sample_rate']=='44100' and stream['channels']==1
    receipt=dict(status='PASS',generator=generator,encoder=encoder,ffmpeg=run(['ffmpeg','-version']).splitlines()[0],wav_sha256=sha(wav),ogg_sha256=sha(ogg),accepted_ogg_sha256=sha(accepted),decoded_f32_sha256=hashes[0],same_decoded_audio=True,accepted_asset_unchanged=True,byte_identical_to_accepted=sha(ogg)==sha(accepted),sample_rate=44100,channels=1,duration=meta['format']['duration'],rights='remake自有；可公開',evidence_grade='hardware-spec approximation',note='固定OGG serial使同版本重建可重現；不覆寫已聆聽資產。')
    (out/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(dict(status='PASS',ogg_sha256=sha(ogg),byte_identical_to_accepted=receipt['byte_identical_to_accepted']),ensure_ascii=False))
if __name__=='__main__':main()
