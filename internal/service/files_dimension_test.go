package service_test

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"strings"
	"testing"

	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 上传图片时必须记录像素宽高。
//
// 回归背景：供给大厅的卡片要在**渲染前**按每张图自己的比例预留位置——只知道 URL
// 无法预留（图加载完再撑开会造成列表跳动），写死一个比例又会把图裁掉。
// 尺寸在上传时从文件头解出并落进 uploads 台账（迁移 000115），列表接口再按 ID 批量取。
func TestUploadRecordsImageDimensions(t *testing.T) {
	ctx := context.Background()
	// 造一张 40×30 的真 PNG（不是伪装成 png 的文本）
	src := image.NewRGBA(image.Rect(0, 0, 40, 30))
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	repo := memory.NewUploadRepository()
	svc := service.NewFileService(t.TempDir(), service.WithUploadQuota(repo, 1<<20))

	rec, err := svc.Upload(ctx, "owner-1", "a.png", "image/png", bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("upload png: %v", err)
	}
	if rec.Width != 40 || rec.Height != 30 {
		t.Fatalf("PNG 尺寸应为 40×30，实际 %d×%d", rec.Width, rec.Height)
	}

	// 非图片：尺寸保持 0，上传本身不该失败（消费方按「比例未知」处理）
	txt, err := svc.Upload(ctx, "owner-1", "b.txt", "text/plain", strings.NewReader("not an image"))
	if err != nil {
		t.Fatalf("upload txt: %v", err)
	}
	if txt.Width != 0 || txt.Height != 0 {
		t.Fatalf("非图片不该有尺寸，实际 %d×%d", txt.Width, txt.Height)
	}

	// 批量取尺寸：图片在，非图片与不存在的 ID 都不在（不报错）
	sizes := svc.ImageSizes(ctx, []string{rec.ID, txt.ID, "file-does-not-exist"})
	if wh, ok := sizes[rec.ID]; !ok || wh[0] != 40 || wh[1] != 30 {
		t.Fatalf("ImageSizes 应返回图片尺寸，实际 %v", sizes)
	}
	if _, ok := sizes[txt.ID]; ok {
		t.Fatal("非图片不该出现在尺寸表里")
	}
	if _, ok := sizes["file-does-not-exist"]; ok {
		t.Fatal("不存在的 ID 不该出现在尺寸表里")
	}
}
