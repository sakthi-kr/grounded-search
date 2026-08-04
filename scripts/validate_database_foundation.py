#!/usr/bin/env python3
from __future__ import annotations
from pathlib import Path
import re

def main()->int:
 root=Path(__file__).resolve().parents[1]
 migrations=sorted((root/'db'/'migrations').glob('*.up.sql'))
 downs=sorted((root/'db'/'migrations').glob('*.down.sql'))
 if len(migrations)!=len(downs): raise ValueError('migration counts differ')
 seed=(root/'db'/'seeds'/'development.sql').read_text(encoding='utf-8')
 required=['alice','bob','carol','disabled-user','public-release-notes','restricted-incident','support-runbook','disabled-document','deleted-document']
 missing=[item for item in required if item not in seed]
 if missing: raise ValueError(f'missing seed fixtures: {missing}')
 if not re.search(r"document_acl_users[\s\S]+?'deny'",seed): raise ValueError('explicit user deny fixture missing')
 print(f'database foundation validation passed: {len(migrations)} migrations, {len(required)} fixtures')
 return 0
if __name__=='__main__': raise SystemExit(main())
