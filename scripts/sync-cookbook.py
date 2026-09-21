#!/usr/bin/env python3
"""Keep tutorial source blocks identical to the executable example. No dependencies."""
from pathlib import Path
import re
import sys

root = Path(__file__).resolve().parents[1]
pattern = re.compile(r'<!-- cookbook-source: ([\w/.-]+) -->.*?<!-- /cookbook-source -->', re.S)
stale = []
for page in (root / 'docs/src/content/docs').rglob('*.md'):
    old = page.read_text()
    def replace(match):
        relative = match[1]
        source = (root / 'examples/community-bot' / relative).resolve()
        source.relative_to(root / 'examples/community-bot')
        language = 'go' if source.suffix == '.go' else 'bash'
        return f'<!-- cookbook-source: {relative} -->\n```{language} title="{relative}"\n{source.read_text().rstrip()}\n```\n<!-- /cookbook-source -->'
    new = pattern.sub(replace, old)
    if new != old:
        stale.append(str(page.relative_to(root)))
        if '--check' not in sys.argv:
            page.write_text(new)
if stale and '--check' in sys.argv:
    sys.exit('Outdated cookbook code; run python3 scripts/sync-cookbook.py:\n' + '\n'.join(stale))
print(f'Cookbook source blocks: {len(stale)} updated' if '--check' not in sys.argv else 'Cookbook source blocks are current')
