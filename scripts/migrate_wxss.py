#!/usr/bin/env python3
"""
迁移 WXSS 样式到 CSS 变量。
用法: python migrate_wxss.py <文件路径...>
"""

import re
import sys

def migrate_wxss(content):
    """对单个 .wxss 文件内容执行替换规则。"""

    # 替换规则表（按顺序执行，由具体到一般）
    rules = [
        # ===== 主色 =====
        (re.compile(r'#2979FF'), 'var(--color-primary)'),
        (re.compile(r'#2B6FF5'), 'var(--color-primary)'),
        (re.compile(r'#165DFF'), 'var(--color-primary)'),
        # ===== 主色浅色 =====
        (re.compile(r'#5B93FF'), 'var(--color-primary-light)'),
        # ===== 主色背景 =====
        (re.compile(r'#ECF2FE'), 'var(--color-primary-bg)'),
        # ===== 背景色 =====
        (re.compile(r'#F5F5F5'), 'var(--color-bg)'),
        (re.compile(r'#F5F6FA'), 'var(--color-bg)'),
        # ===== 白色（只在 CSS 值位置替换） =====
        # #FFFFFF 全大写形式
        (re.compile(r'#FFFFFF'), 'var(--color-white)'),
        # #fff 短形式 - 使用负向前瞻避免匹配到 #ffffff 等
        # 在 .wxss 中 #fff 只出现在 CSS 值位置，安全替换
        (re.compile(r'(?<![#\w-])#fff\b'), 'var(--color-white)'),
        # ===== 文字色 =====
        (re.compile(r'#1A1A1A'), 'var(--color-text)'),
        (re.compile(r'#969799'), 'var(--color-text-muted)'),
        (re.compile(r'#6b7280'), 'var(--color-text-muted)'),
        # ===== 占位色 =====
        (re.compile(r'#C8C9CC'), 'var(--color-text-placeholder)'),
        (re.compile(r'#9ca3af'), 'var(--color-text-placeholder)'),
        # ===== 危险色 =====
        (re.compile(r'#FA3E3E'), 'var(--color-danger)'),
        (re.compile(r'#F53F3F'), 'var(--color-danger)'),
        # ===== 成功色 =====
        (re.compile(r'#09B852'), 'var(--color-success)'),
        # ===== 警告色 =====
        (re.compile(r'#FCB42A'), 'var(--color-warning)'),
        (re.compile(r'#FC8F3E'), 'var(--color-warning)'),
        # ===== 分割线 =====
        (re.compile(r'#F2F3F5'), 'var(--color-divider)'),
        (re.compile(r'#F0F0F0'), 'var(--color-divider)'),
        # ===== 旧变量声明删除 =====
        # --primary: 开头的自定义属性声明（带分号结尾）
        (re.compile(r'--primary:\s*[^;]+;\s*'), ''),
        # --gradient-header: 等旧变量声明
        (re.compile(r'--gradient-[a-z]+:\s*[^;]+;\s*'), ''),
        # ===== 阴影 =====
        (re.compile(r'rgba\(\s*0\s*,\s*0\s*,\s*0\s*,\s*0\.04\s*\)'), 'var(--shadow-card)'),
        # rgba(0,0,0,0.08) 保留，规则说不存在 var(--shadow-hover)
        # ===== 圆角 =====
        (re.compile(r'border-radius:\s*16rpx'), 'border-radius: var(--radius)'),
    ]

    for pattern, replacement in rules:
        content = pattern.sub(replacement, content)

    return content


def process_file(filepath):
    """处理单个文件。返回是否做了修改。"""
    with open(filepath, 'r', encoding='utf-8') as f:
        original = f.read()

    modified = migrate_wxss(original)

    if original != modified:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(modified)
        return True
    return False


if __name__ == '__main__':
    files = sys.argv[1:]
    if not files:
        print('用法: python migrate_wxss.py <文件路径...>')
        sys.exit(1)

    changed = []
    unchanged = []
    for fp in files:
        try:
            if process_file(fp):
                changed.append(fp)
            else:
                unchanged.append(fp)
        except Exception as e:
            print(f'ERROR: {fp} - {e}')

    print(f'\n=== 结果 ===')
    print(f'修改: {len(changed)} 个文件')
    print(f'未变: {len(unchanged)} 个文件')
    for f in changed:
        print(f'  + {f}')
