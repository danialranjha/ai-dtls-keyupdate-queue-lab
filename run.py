"""Offline test and evidence runner. Needs locally installed Go >=1.22 and Python3."""
import json,os,subprocess,tempfile
from pathlib import Path
from verify_source import verify
root=Path(__file__).resolve().parent
with tempfile.TemporaryDirectory(prefix='queue-lab-cache-') as cache:
    env=dict(os.environ,GOCACHE=cache,GOPROXY='off',GOSUMDB='off',GOTOOLCHAIN='local')
    bodies=verify()
    p=subprocess.run(['go','test','-json','./...'],cwd=root,env=env,text=True,capture_output=True)
    events=[json.loads(line) for line in p.stdout.splitlines() if line.startswith('{')]
    passed=[e['Test'] for e in events if e.get('Action')=='pass' and 'Test' in e]
    if p.returncode:print(p.stdout,p.stderr);raise SystemExit(p.returncode)
    assert len(passed)==11,passed
    subprocess.run(['go','run','.'],cwd=root,env=env,check=True)
    (root/'artifacts/checks.json').write_text(json.dumps({'tests_passed':passed,'tests_failed':0,'exact_source_bodies':bodies,'scope':'component harness; no network or model'},indent=2)+'\n')
    print(f'{len(passed)} tests passed; {bodies} exact function bodies verified')
