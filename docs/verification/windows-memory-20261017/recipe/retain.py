"""沿既有保存工具封存本輪證據；只在 Docker 執行。"""
import argparse
import importlib.util
import json
from pathlib import Path

assert Path('/.dockerenv').exists()
root=Path('/src')
spec=importlib.util.spec_from_file_location('retained',root/'tools/retain_verification.py')
module=importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
module.OUT=root/'docs/verification/local'
name='windows-crash-20261017'
module.NAMES=(*module.NAMES,name)
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('stage',choices=['archive','verify','delete'])
stage=parser.parse_args().stage
if stage=='archive':
    module.archive(name,0)
    index=module.verify(name)
    public=dict(index,local_archive_path='docs/verification/local/'+name+'/evidence.tar.xz')
    (root/'docs/verification/windows-memory-20261017/retention.json').write_text(json.dumps(public,ensure_ascii=False,indent=2)+'\n')
elif stage=='verify':
    module.verify(name)
else:
    module.delete(name)
