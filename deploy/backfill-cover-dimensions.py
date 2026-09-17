#!/usr/bin/env python3
"""把商品引用、但台账里没有记录的图片登记进 uploads 台账并写入尺寸。"""
"""
为什么需要（2026-09-17）：供给大厅的卡片要按每张图自己的比例预留位置，尺寸来自
uploads 台账。自家上传（file-<32hex>）在迁移 000115 后已由第一个回填脚本补齐，
但商品封面里还有两类**在台账里查不到**的图：
  - /uploads/sl-*.jpg   随包种子图，直接躺在 uploads 卷里，但不是用户上传、没有台账行
  - /static/home/*.jpg  前端静态资源，位于 /var/www/admin/static/
结果这 16 件商品的封面拿不到尺寸、退化成 1:1，和写死比例没有区别。

做法：按「文件名即台账主键」的约定补一行台账（owner_id=seed）。这些行不参与任何
用户配额（配额按 owner_id 统计），也不会被账号注销流程清理（没有用户 id 是 seed）。
重复执行安全：已有台账行的图片直接跳过。
"""
import hashlib
import json
import os
import struct
import subprocess

VOL = "/var/lib/docker/volumes/uav_uploads/_data"
STATIC = "/var/www/admin/static"
PSQL = ["docker", "exec", "uav-db-1", "psql", "-U", "drone", "-d", "drone_platform", "-t", "-A", "-c"]


def dims(path):
    try:
        with open(path, "rb") as f:
            head = f.read(32)
            if head[:2] == b"\xff\xd8":
                f.seek(2)
                while True:
                    b = f.read(1)
                    while b and b != b"\xff":
                        b = f.read(1)
                    m = f.read(1)
                    while m == b"\xff":
                        m = f.read(1)
                    if not m:
                        return None
                    v = m[0]
                    if v in (0xD8, 0xD9) or 0xD0 <= v <= 0xD7:
                        continue
                    ln = struct.unpack(">H", f.read(2))[0]
                    if v in (0xC0, 0xC1, 0xC2, 0xC3, 0xC5, 0xC6, 0xC7, 0xC9, 0xCA, 0xCB, 0xCD, 0xCE, 0xCF):
                        d = f.read(ln - 2)
                        h, w = struct.unpack(">HH", d[1:5])
                        return (w, h)
                    f.seek(ln - 2, 1)
            elif head[:8] == b"\x89PNG\r\n\x1a\n":
                w, h = struct.unpack(">II", head[16:24])
                return (w, h)
    except Exception:
        return None
    return None


def q(sql):
    return subprocess.run(PSQL + [sql], capture_output=True, text=True).stdout.strip()


def resolve(url):
    """站点相对地址 → 服务器上的真实文件路径。"""
    if url.startswith("/uploads/"):
        rel = url[len("/uploads/"):]
        p = os.path.join(VOL, rel)
        return p if os.path.isfile(p) else None
    if url.startswith("/static/"):
        rel = url[len("/static/"):]
        p = os.path.join(STATIC, rel)
        return p if os.path.isfile(p) else None
    return None


# 商品引用到的全部图片（封面 + 图集），取站点相对路径的那些
raw = q("SELECT DISTINCT jsonb_array_elements_text(images) FROM drone_products WHERE images IS NOT NULL AND jsonb_array_length(images) > 0")
urls = [u.strip() for u in raw.splitlines() if u.strip().startswith("/")]
print(f"商品引用的站点内图片: {len(urls)} 个不同地址")

added = skipped = missing = failed = 0
for url in urls:
    key = url.rsplit("/", 1)[-1]
    if not key:
        continue
    exists = q(f"SELECT count(*) FROM uploads WHERE id = '{key}'")
    if exists not in ("0", ""):
        skipped += 1
        continue
    path = resolve(url)
    if not path:
        missing += 1
        print(f"  文件不在服务器上: {url}")
        continue
    d = dims(path)
    if not d:
        failed += 1
        print(f"  解析失败: {url}")
        continue
    w, h = d
    size = os.path.getsize(path)
    with open(path, "rb") as f:
        sha = hashlib.sha256(f.read()).hexdigest()
    ext = os.path.splitext(key)[1].lower()
    ctype = {'.png': 'image/png', '.gif': 'image/gif'}.get(ext, 'image/jpeg')
    q(f"INSERT INTO uploads (id, owner_id, storage_key, sha256, content_type, size_bytes, visibility, created_at, width, height) "
      f"VALUES ('{key}', 'seed', '{url}', '{sha}', '{ctype}', {size}, 'public', now(), {w}, {h}) "
      f"ON CONFLICT (id) DO UPDATE SET width = EXCLUDED.width, height = EXCLUDED.height")
    added += 1
    print(f"  + {key:34s} {w}x{h}  比例 {w/h:.2f}  ({url})")

print()
print(f"登记 {added}，已有台账跳过 {skipped}，文件缺失 {missing}，解析失败 {failed}")
print("台账里有尺寸的总数:", q("SELECT count(*) FROM uploads WHERE width > 0 AND height > 0"))
