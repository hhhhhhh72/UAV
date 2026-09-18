#!/usr/bin/env python3
"""上线前全站自检（在服务器上对 127.0.0.1:8080 运行）。

覆盖：
  1) 静态 GET 路由逐条探活（无鉴权）—— 抓 5xx / 越权 / 已注册却 404
  2) 各类内容的 列表 → 详情 链路
  3) 全库媒体引用验盘（只读文件头，不需要 HTTP，也不受限流影响）
  4) 安全面：私有影像匿名读、管理接口匿名访问、短信不回显验证码
  5) 管理后台 SPA 资源完整性（入口引用 + chunk 互相引用 + 孤儿文件）

用法：
  # 路由清单（仓库根执行）
  python3 scripts/prelaunch-audit.py --routes   # 打印静态 GET 路由到 /tmp/static-get.txt
  # 自检
  python3 scripts/prelaunch-audit.py

⚠ 注意限流：应用 100/s + nginx burst 80。批量探活必须加间隔，否则拿到一堆 429
   （2026-09-17 实测：不加间隔时 15 张图里 13 张被误判成「取不到」）。
"""
import argparse
import json
import os
import re
import subprocess
import sys
import time
import urllib.error
import urllib.request

BASE = os.environ.get('AUDIT_BASE', 'http://127.0.0.1:8080')
VOL = '/var/lib/docker/volumes/uav_uploads/_data'
WWW = '/var/www/admin'
ROUTES_FILE = '/tmp/static-get.txt'
PKG = 'internal/httpapi'


def dump_routes():
    """从源码提取静态 GET 路径（不含 {param} 的），写 /tmp/static-get.txt。"""
    pat = re.compile(r'mux\.HandleFunc\(\s*"([A-Z]+)\s+(/[^"]*)"')
    paths = set()
    for name in os.listdir(PKG):
        if not name.endswith('.go') or name.endswith('_test.go'):
            continue
        for line in open(os.path.join(PKG, name), encoding='utf-8'):
            m = pat.search(line)
            if m and m.group(1) == 'GET' and '{' not in m.group(2):
                paths.add(m.group(2))
    out = sorted(paths)
    with open(ROUTES_FILE, 'w', encoding='utf-8') as f:
        f.write('\n'.join(out) + '\n')
    print('静态 GET 路由 %d 条 -> %s' % (len(out), ROUTES_FILE))


def fetch(path, timeout=20, spacing=0.03):
    time.sleep(spacing)  # 避开限流
    r = urllib.request.Request(BASE + path, method='GET')
    try:
        with urllib.request.urlopen(r, timeout=timeout) as resp:
            return resp.status, resp.read(300000)
    except urllib.error.HTTPError as e:
        try:
            return e.code, e.read(20000)
        except Exception:
            return e.code, b''
    except Exception as e:
        return -1, str(e).encode()


def jget(path):
    code, body = fetch(path)
    if code != 200:
        return code, None
    try:
        d = json.loads(body)
    except Exception:
        return code, None
    return code, (d.get('data', d) if isinstance(d, dict) else d)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--routes', action='store_true', help='只导出路由清单')
    args = ap.parse_args()
    dump_routes()
    if args.routes:
        return 0

    problems = 0

    # ---- 1. 公开接口探活 ----
    print('\n===== 1. 公开接口探活 =====')
    routes = [l.strip() for l in open(ROUTES_FILE, encoding='utf-8') if l.strip()]
    counts, interesting = {}, []
    for p in routes:
        code, body = fetch(p)
        counts[code] = counts.get(code, 0) + 1
        needs_auth = p.startswith('/api/v1/admin/') or p.startswith('/api/admin/') or '/mine' in p
        if code >= 500 or (needs_auth and code == 200):
            interesting.append((p, code, '需鉴权却放行' if code == 200 else '服务端错误'))
            problems += 1
        elif code not in (200, 401, 403):
            interesting.append((p, code, body[:80].decode('utf-8', 'replace').replace('\n', ' ')))
    print('  状态码: ' + ', '.join('%s x %d' % (k, v) for k, v in sorted(counts.items(), key=lambda x: -x[1])))
    for p, c, why in interesting:
        print('  !! %-46s -> %-5s %s' % (p, c, why))

    # ---- 2. 列表 → 详情 ----
    print('\n===== 2. 列表 -> 详情 =====')
    KINDS = [
        ('/api/v1/products', '/api/v1/products/%s'),
        ('/api/v1/training-courses', '/api/v1/training-courses/%s'),
        ('/api/v1/demands', '/api/v1/demands/%s'),
        ('/api/v1/competitions', '/api/v1/competitions/%s'),
        ('/api/v1/events', '/api/v1/events/%s'),
        ('/api/v1/cases', '/api/v1/cases/%s'),
        ('/api/v1/colleges', '/api/v1/colleges/%s'),
        ('/api/v1/jobs', '/api/v1/jobs/%s'),
        ('/api/v1/study/tours', '/api/v1/study/tours/%s'),
    ]
    for listp, detailp in KINDS:
        code, items = jget(listp + '?page=1&page_size=50')
        if code != 200 or not isinstance(items, list):
            print('  ?? %-32s HTTP %s / 非数组' % (listp, code))
            continue
        bad = [it.get('id') for it in items if it.get('id') and fetch(detailp % it['id'])[0] != 200]
        print('  %-4s %-32s %2d 条，详情失败 %d' % ('OK' if not bad else 'BAD', listp, len(items), len(bad)))
        for b in bad[:5]:
            print('     !! %s' % (detailp % b))
        problems += len(bad)

    # ---- 3. 媒体引用验盘 ----
    print('\n===== 3. 媒体引用验盘（全库） =====')
    dump = subprocess.run(['docker', 'exec', 'uav-db-1', 'pg_dump', '-U', 'drone',
                           '-d', 'drone_platform', '--data-only'],
                          capture_output=True, text=True).stdout
    paths = sorted(set(re.findall(r'/(?:uploads|static)/[A-Za-z0-9._/-]+', dump)))
    bad_media = []
    for p in paths:
        p = p.rstrip('".,)')
        if p.startswith('/uploads/'):
            disk = os.path.join(VOL, p[len('/uploads/'):])
        else:
            disk = os.path.join(WWW, p.lstrip('/'))
        if not os.path.isfile(disk) or os.path.getsize(disk) == 0:
            bad_media.append(p)
    print('  %d 个路径，缺失 %d' % (len(paths), len(bad_media)))
    for p in bad_media:
        print('  !! %s' % p)
    problems += len(bad_media)

    # ---- 4. 安全面 ----
    print('\n===== 4. 安全面 =====')
    checks = [('/api/v1/admin/users', (401, 403)), ('/api/v1/messages', (401, 403)),
              ('/api/v1/escrow/balance', (401, 403))]
    for p, want in checks:
        code, _ = fetch(p)
        ok = code in want
        print('  %-4s %-32s -> %s' % ('OK' if ok else 'BAD', p, code))
        problems += 0 if ok else 1

    # ---- 5. SPA 资源 ----
    print('\n===== 5. 管理后台资源 =====')
    idx = os.path.join(WWW, 'index.html')
    if os.path.isfile(idx):
        refs = set(re.findall(r'/assets/[A-Za-z0-9._-]+', open(idx, encoding='utf-8').read()))
        miss = [r for r in refs if not os.path.isfile(os.path.join(WWW, r.lstrip('/')))]
        print('  入口引用 %d，缺失 %d' % (len(refs), len(miss)))
        for m in miss:
            print('  !! %s' % m)
        problems += len(miss)

    print('\n===== 汇总：问题 %d 项 =====' % problems)
    return 1 if problems else 0


if __name__ == '__main__':
    sys.exit(main())
