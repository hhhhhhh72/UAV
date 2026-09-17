package httpapi

import (
	"context"
	"strings"

	"drone-platform/internal/domain"
)

// fillCoverDimensions 给商品的封面图补上像素尺寸。
//
// 为什么要（2026-09-17）：供给大厅的卡片必须在**渲染前**知道每张图的比例才能预留位置——
// 只知道 URL 无法预留（图片加载完再撑开会造成列表跳动 CLS），写死一个比例又会裁图。
// 尺寸在上传时就记进 uploads 台账（迁移 000115），这里按 ID 批量取，整页一次查询，
// 不按商品逐条查（避免 N+1）。
//
// 取不到尺寸不是错误：封面可能是外链或 /static/ 种子图，此时保持 0，
// 前端退化为 1:1——宁可比例不准，也不能让整个列表接口失败。
func (s *Server) fillCoverDimensions(ctx context.Context, items []domain.DroneProduct) {
	if s.fileSvc == nil || len(items) == 0 {
		return
	}
	ids := make([]string, 0, len(items))
	for i := range items {
		if id := uploadIDFromURL(firstImage(items[i].Images)); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return
	}
	sizes := s.fileSvc.ImageSizes(ctx, ids)
	for i := range items {
		u := firstImage(items[i].Images)
		id := uploadIDFromURL(u)
		if id == "" {
			continue
		}
		wh, ok := sizes[id]
		if !ok {
			continue
		}
		items[i].CoverWidth, items[i].CoverHeight = wh[0], wh[1]
		items[i].CoverURL = u
	}
}

func firstImage(images []string) string {
	for _, u := range images {
		if strings.TrimSpace(u) != "" {
			return strings.TrimSpace(u)
		}
	}
	return ""
}

// uploadIDFromURL 从 /uploads/<id> 形式的图片地址取出上传 ID（文件名就是台账主键）。
// 非本站上传的地址（外链、/static/ 种子图）返回空串，调用方按「尺寸未知」处理。
// 收得比「以 file- 开头」更紧：必须是 /uploads/ 路径下、文件名以 file- 开头，
// 否则一个外链 https://evil/file-xxx 也会被当成自家台账去查。
func uploadIDFromURL(raw string) string {
	u := strings.TrimSpace(raw)
	if u == "" || !strings.Contains(u, "/uploads/") {
		return ""
	}
	// 去掉查询串/锚点，再取路径最后一段
	if i := strings.IndexAny(u, "?#"); i >= 0 {
		u = u[:i]
	}
	u = strings.TrimSuffix(u, "/")
	id := u
	if i := strings.LastIndex(u, "/"); i >= 0 {
		id = u[i+1:]
	}
	if !strings.HasPrefix(id, "file-") || len(id) <= len("file-") {
		return ""
	}
	return id
}
