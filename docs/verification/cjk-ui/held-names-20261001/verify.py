from pathlib import Path
import subprocess, json, hashlib, os

root = Path('workplace/cjk-held-names-20261001')
def sha(p): return hashlib.sha256(Path(p).read_bytes()).hexdigest()
selection = json.loads((root/'selection.json').read_text())
binary = 'workplace/platform-smoke-20261001/build/kol-linux'
assert sha(binary) == selection['binary_sha256']
assert (root.stat().st_uid, root.stat().st_gid) == (os.getuid(), os.getgid())
oldtable = str(root/'before-display-names.tsv')
records = []
def run(label, script, shots, language, table, save=None, load=None):
    out = root/label
    args = [binary, '-scene', 'play' if load else 'flow', '-seed', '1',
            '-lang', language, '-display-names', table,
            '-play-script' if load else '-flow-script', ' '.join(script),
            '-shot-dir', str(out), '-shot-steps', ','.join(map(str, shots)),
            '-trace', str(out/'trace.json')]
    if save: args += ['-save', str(save)]
    if load: args += ['-load', str(load)]
    if out.exists():
        assert '-maul-' in label, out
        record=json.loads((out/'run.json').read_text())
        assert record['command']==args and record['returncode']==0
        assert record['binary_sha256']==sha(binary) and record['catalog_sha256']==sha(table)
        trace=json.loads((out/'trace.json').read_text())
        assert len(trace['steps'])==len(script)+1
        if save: assert save.is_file()
        records.append(record)
        return trace
    out.mkdir()
    result = subprocess.run(args, capture_output=True, text=True, timeout=70)
    (out/'engine.log').write_text(result.stdout+result.stderr)
    record = {'label':label, 'command':args, 'returncode':result.returncode,
              'binary_sha256':sha(binary), 'catalog_sha256':sha(table),
              'seed':1, 'inputs':script, 'expected_states':len(script)+1}
    (out/'run.json').write_text(json.dumps(record, ensure_ascii=False, indent=2)+'\n')
    assert result.returncode == 0, (label, result.stderr)
    trace = json.loads((out/'trace.json').read_text())
    assert len(trace['steps']) == len(script)+1
    if save: assert save.is_file(), (label, trace['steps'][-1])
    records.append(record)
    print(label, len(script)+1, trace['steps'][-1], flush=True)
    return trace

tail = ['R','L','RET','SPACE','1','>','>','>','>','RET','>','RET','RET']
plans = []
for kind, scriptfile, steps, member, down in [
    ('maul', 'scout-script.txt', [60,61,63,78,79,83], '3', 15),
    ('warder', 'warder-script.txt', [15,16,18,25,26,30], '1', 7)]:
    base = (root/scriptfile).read_text().split()
    # 單人直接進價錢問句；多人先答「全隊嗎」，選 NO 才進價錢問句。
    normal = base+(tail if kind=='maul' else tail[:10]+['RET'])
    reload = [member,'>','RET']+['D']*down+['RET']
    plans.append({'kind':kind, 'normal':normal, 'reload':reload,
                  'shots':steps+[len(normal)], 'reload_shots':[0,3,len(reload)-1,len(reload)]})
assert not (root/'verification-plan-v2.json').exists()
(root/'verification-plan-v2.json').write_text(json.dumps(plans, ensure_ascii=False, indent=2)+'\n')
for plan in plans:
    kind = plan['kind']; normal = plan['normal']; baseline = None; saves = []
    for version, table in [('before',oldtable), ('after','docs/i18n/display-names-zh-Hant.tsv')]:
        for language in ['en','zh-Hant']:
            label = f'{version}-{language}-{kind}-save'+('-v2' if kind=='warder' else '')
            save = root/(label+'.json')
            trace = run(label, normal, plan['shots'], language, table, save=save)
            if baseline is None: baseline = trace
            else: assert trace == baseline, label
            data = json.loads(save.read_text()); data.pop('saved_at')
            saves.append(data)
    assert all(x == saves[0] for x in saves)
    held = next(c for c in saves[0]['characters'] if c['name'] == ('Olear' if kind=='maul' else 'Warder'))
    item = held['backpack'][4] if kind=='maul' else held['worn'][7]
    expected = selection['selected' if kind=='maul' else 'warder']
    assert item['name'] == expected['name'] and item['record'] == expected['record']
    assert expected['name'] in json.dumps(saves[0])
    rebaseline = None
    for language in ['en','zh-Hant']:
        suffix='-v2' if kind=='warder' else ''
        save = root/(f'after-{language}-{kind}-save'+suffix+'.json'); beforehash=sha(save)
        trace = run(f'after-{language}-{kind}-reload'+suffix, plan['reload'], plan['reload_shots'],
                    language,'docs/i18n/display-names-zh-Hant.tsv',load=save)
        assert sha(save)==beforehash
        if rebaseline is None: rebaseline=trace
        else: assert rebaseline==trace
(root/'state-checks.json').write_text(json.dumps({
    'schema':'kol-held-names-state-checks-v1', 'runs':records,
    'normal_states_identical_before_after_en_zh':True,
    'normalized_saves_identical_before_after_en_zh':True,
    'reload_states_identical_en_zh':True, 'reload_did_not_rewrite_save':True,
    'original_names_and_item_records_preserved':True
},ensure_ascii=False,indent=2)+'\n')
