#!/usr/bin/env python3
"""以Docker唯讀/src、/out重验；run_verify.sh會重播Go規則。--seal僅首封/seal。"""
from pathlib import Path
import argparse, contextlib, hashlib, io, json, os, runpy
ROOT = Path(__file__).resolve().parents[3]
PRIVATE = ROOT / 'workplace/negative-area-20261003'
HERE = Path(__file__).resolve().parent
read = lambda p: json.loads(p.read_text())
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()

def validate():
    saved = read(PRIVATE / 'evidence-review.json')
    assert read(Path(os.environ['KOL_NEGATIVE_RULE_REPLAY'])) == saved['rules']
    captured = {}; original = Path.write_text
    def capture(path, text, *args, **kwargs):
        assert str(path) == '/out/evidence-review.json'
        captured[path.name] = json.loads(text)
        return len(text)
    try:
        Path.write_text = capture
        with contextlib.redirect_stdout(io.StringIO()):
            runpy.run_path(str(PRIVATE / 'review.py'))
    finally:
        Path.write_text = original
    assert captured == {'evidence-review.json': saved}
    parent = read(PRIVATE / 'parent-recheck.json')
    assert parent == dict(parent_receipt_sha256=saved['parent_receipt_sha256'], parent_attachments=1715, recalculated=True)
    driver = read(PRIVATE / 'negative1/driver.json')
    for key, name in [('startup_sha256', 'instrumented-book-private.go'), ('driver_sha256', 'native-driver.go'), ('keys_sha256', 'keys.txt')]:
        assert sha(PRIVATE / 'negative1' / name) == driver[key]
    assert (PRIVATE / 'negative1/native-driver.go').read_text().count('.WriteU16(') == 2
    return saved

def main():
    parser = argparse.ArgumentParser(); parser.add_argument('--seal', action='store_true'); args = parser.parse_args()
    result = validate()
    if args.seal:
        path = Path('/seal/receipt.json'); assert not path.exists()
        names = ['CONTEXT.md', 'WORKLOG.md', 'WORKLIST.md', 'docs/worklist.json', 'docs/re/RESEARCH-LOG.md',
                 'docs/re/20261003-negative-area.md', 'docs/spec/012-combat-spells.md',
                 'docs/verification/negative-area-20261003/verify.py']
        attachments = {name: sha(ROOT / name) for name in names}
        attachments.update(result['production_inputs'])
        for p in sorted(PRIVATE.rglob('*')):
            if p.is_file():
                attachments[str(p.relative_to(ROOT))] = sha(p)
        result = dict(result, rights='local-only', attachments=attachments)
        path.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    else:
        stored = read(HERE / 'receipt.json')
        assert result == {k:v for k,v in stored.items() if k not in ['rights', 'attachments']}
        for name, h in stored['attachments'].items():
            assert sha(ROOT / name) == h, name
    print(json.dumps(dict(normal_prefix_states=175, negative=True, target_chest=5,
                          answer_seed=2128043714, local_rule_passed=True,
                          prototype_direction_key_gap=True, production_changed=False,
                          parent_attachments=1715, sealed=args.seal), ensure_ascii=False))

if __name__ == '__main__':
    main()
