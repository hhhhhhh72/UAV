#!/usr/bin/env perl
use strict;
use warnings;
use utf8;
# 不使用 File::Slurp，用原生 open

binmode(STDOUT, ':utf8');

sub migrate_wxss {
    my ($content) = @_;
    my $modified = 0;
    my @rules = (
        # ===== 主色 =====
        [qr/#2979FF/,             'var(--color-primary)'],
        [qr/#2B6FF5/,             'var(--color-primary)'],
        [qr/#165DFF/,             'var(--color-primary)'],
        # ===== 主色浅色 =====
        [qr/#5B93FF/,             'var(--color-primary-light)'],
        # ===== 主色背景 =====
        [qr/#ECF2FE/,             'var(--color-primary-bg)'],
        # ===== 背景色 =====
        [qr/#F5F5F5/,             'var(--color-bg)'],
        [qr/#F5F6FA/,             'var(--color-bg)'],
        # ===== 白色 =====
        [qr/#FFFFFF/,             'var(--color-white)'],
        # #fff - 只在 CSS 值位置，避免匹配 #ffffff 等
        [qr/(?<![#\w-])#fff\b/,   'var(--color-white)'],
        # ===== 文字色 =====
        [qr/#1A1A1A/,             'var(--color-text)'],
        [qr/#969799/,             'var(--color-text-muted)'],
        [qr/#6b7280/,             'var(--color-text-muted)'],
        # ===== 占位色 =====
        [qr/#C8C9CC/,             'var(--color-text-placeholder)'],
        [qr/#9ca3af/,             'var(--color-text-placeholder)'],
        # ===== 危险色 =====
        [qr/#FA3E3E/,             'var(--color-danger)'],
        [qr/#F53F3F/,             'var(--color-danger)'],
        # ===== 成功色 =====
        [qr/#09B852/,             'var(--color-success)'],
        # ===== 警告色 =====
        [qr/#FCB42A/,             'var(--color-warning)'],
        [qr/#FC8F3E/,             'var(--color-warning)'],
        # ===== 分割线 =====
        [qr/#F2F3F5/,             'var(--color-divider)'],
        [qr/#F0F0F0/,             'var(--color-divider)'],
        # ===== 旧变量声明删除 =====
        [qr/--primary:\s*[^;]+;\s*/,  ''],
        [qr/--gradient-[a-z]+:\s*[^;]+;\s*/, ''],
        # ===== 阴影 =====
        [qr/rgba\(\s*0\s*,\s*0\s*,\s*0\s*,\s*0\.04\s*\)/, 'var(--shadow-card)'],
        # ===== 圆角 =====
        [qr/border-radius:\s*16rpx/, 'border-radius: var(--radius)'],
    );

    for my $rule (@rules) {
        my ($pattern, $replacement) = @$rule;
        my $count = $content =~ s/$pattern/$replacement/g;
        $modified += $count;
    }

    return ($content, $modified);
}

sub process_file {
    my ($filepath) = @_;

    # 读取文件
    open(my $fh, '<:utf8', $filepath) or die "无法打开 $filepath: $!";
    my $original = do { local $/; <$fh> };
    close($fh);

    my ($modified, $count) = migrate_wxss($original);

    if ($count > 0) {
        open(my $fh, '>:utf8', $filepath) or die "无法写入 $filepath: $!";
        print $fh $modified;
        close($fh);
        return ($filepath, $count, 1);
    }
    return ($filepath, 0, 0);
}

# 主程序
my @files = @ARGV;
if (!@files) {
    print "用法: perl scripts/migrate_wxss.pl <文件...>\n";
    exit 1;
}

my @changed;
my $unchanged = 0;

for my $fp (@files) {
    if (!-f $fp) {
        print "WARN: 文件不存在 $fp\n";
        next;
    }
    my ($path, $count, $is_changed) = process_file($fp);
    if ($is_changed) {
        push @changed, { path => $path, count => $count };
        print "  + $path ($count 处替换)\n";
    } else {
        $unchanged++;
    }
}

print "\n=== 结果 ===\n";
print "修改: " . scalar(@changed) . " 个文件\n";
print "未变: $unchanged 个文件\n";
