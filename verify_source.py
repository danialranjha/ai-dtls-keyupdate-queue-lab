"""Offline hash verification of exact copied function bodies, not full transport."""
import hashlib,json,re
from pathlib import Path
root=Path(__file__).resolve().parent
def verify():
    m=json.loads((root/'source.json').read_text())
    count=0
    for mode in ['original','coalesce_only','progress_only','merged']:
        text=(root/mode/'source.go').read_text()
        for n in m['functions']:
            new=(mode in ['progress_only','merged']) if n in ['startQueuedPostHandshake','commandEligible'] else mode in ['coalesce_only','merged']
            v='new' if new else 'old'
            start=re.search(r'^func (?:\([^\n]+\) )?'+n+r'\(',text,re.M).start()
            body=text[start:text.index('\n}',start)+2]+'\n'
            assert hashlib.sha256(body.encode()).hexdigest()==m['versions'][v]['function_sha256'][n],(mode,n)
            count+=1
    return count
if __name__=='__main__':print(f'{verify()} exact function bodies verified')
