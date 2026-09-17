#!/usr/bin/env python3
"""回填存量上传图片的像素宽高（迁移 000115 新增列 width/height）。

上传链路从本次起会自己记录尺寸，这里只处理迁移之前的历史数据。
只读文件头（PNG 的 IHDR / JPEG 的 SOF），不解码像素——对 5MB 的图也只是一次很小的读。
解析不出的（非图片、损坏、文件已不在卷里）保持 0，前端会退化为 1:1，不影响功能。
"""
import os
import struct
import subprocess
import sys

VOL = "/var/lib/docker/volumes/uav_uploads/_data"
PSQL = ["docker", "exec", "uav-db-1", "psql", "-U", "drone", "-d", "drone_platform", "-t", "-A", "-F", "|", "-c"]


def dims(path):
    """从文件头读宽高，失败返回 None。"""
    try:
        with open(path, "rb") as f:
            head = f.read(32)
            if head[:2] == b"\xff\xd8":  # JPEG
                f.seek(2)
                while True:
                    b = f.read(1)
                    while b and b != b"\xff":
                        b = f.read(1)
                    marker = f.read(1)
                    while marker == b"\xff":
                        marker = f.read(1)
                    if not marker:
                        return None
                    m = marker[0]
                    if m in (0xD8, 0xD9) or 0xD0 <= m <= 0xD7:
                        continue
                    ln = struct.unpack(">H", f.read(2))[0]
                    if m in (0xC0, 0xC1, 0xC2, 0xC3, 0xC5, 0xC6, 0xC7, 0xC9, 0xCA, 0xCB, 0xCD, 0xCE, 0xCF):
                        d = f.read(ln - 2)
                        h, w = struct.unpack(">HH", d[1:5])
                        return (w, h)
                    f.seek(ln - 2, 1)
            elif head[:8] == b"\x89PNG\r\n\x1a\n":
                w, h = struct.unpack(">II", head[16:24])
                return (w, h)
            elif head[:3] == b"GIF":
                w, h = struct.unpack("<HH", head[6:10])
                return (w, h)
    except Exception:
        return None
    return None


def q(sql):
    r = subprocess.run(PSQL + [sql], capture_output=True, text=True)
    return r.stdout.strip()


def locate(upload_id, storage_key):
    """storage_key 形如 uploads/file-xxx 或 uploads/private/file-xxx；
    容器里 cwd=/，卷挂在 /uploads，所以去掉开头的 uploads/ 就是卷内路径。"""
    key = (storage_key or "").strip()
    if key.startswith("uploads/"):
        key = key[len("uploads/"):]
    candidates = []
    if key:
        candidates.append(os.path.join(VOL, key))
    candidates.append(os.path.join(VOL, upload_id))
    candidates.append(os.path.join(VOL, "private", upload_id))
    for c in candidates:
        if os.path.isfile(c):
            return c
    return None


rows = q("SELECT id, COALESCE(storage_key,'') FROM uploads WHERE width = 0 OR height = 0")
targets = [line.split("|") for line in rows.splitlines() if line.strip()]
print(f"待回填: {len(targets)} 条")

filled = missing = failed = 0
for upload_id, storage_key in targets:
    path = locate(upload_id, storage_key)
    if not path:
        missing += 1
        continue
    d = dims(path)
    if not d:
        failed += 1
        continue
    w, h = d
    q(f"UPDATE uploads SET width = {w}, height = {h} WHERE id = '{upload_id}'")
    filled += 1
    print(f"  {upload_id}  {w}x{h}  ({os.path.basename(path)})")

print()
print(f"回填成功 {filled}，文件不在卷里 {missing}，解析失败 {failed}")
print("剩余未回填:", q("SELECT count(*) FROM uploads WHERE width = 0 OR height = 0"))
print("已记录尺寸:", q("SELECT count(*) FROM uploads WHERE width > 0 AND height > 0"))
