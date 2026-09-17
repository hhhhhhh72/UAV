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
		if id := imageLedgerKey(firstImage(items[i].Images)); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return
	}
	sizes := s.fileSvc.ImageSizes(ctx, ids)
	for i := range items {
		u := firstImage(items[i].Images)
		id := imageLedgerKey(u)
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

// imageLedgerKey 把图片地址映射成 uploads 台账的主键。
//
// 台账主键就是**磁盘上的文件名**（自家上传是 file-<32hex>，种子图沿用自己的名字
// sl-hero.jpg / demand-lift.jpg 之类），所以取路径最后一段即可。
//
// 只接受**站点相对路径**（单个前导 /）：外链（https://…、协议相对的 //…）一律返回空串。
// 这条限制是必要的——否则 https://evil.example/file-abc 会被拿去查我们的台账。
// 查询只用于取宽高，不涉及任何授权，但仍不该让外部地址进到自家索引里。
// imagePaths 是本站图片的两个根：自家上传在 /uploads/，随包种子图在 /static/。
// 只有落在这两个路径下的地址才查台账——外部地址即使文件名撞上也不查。
func imageLedgerKey(raw string) string {
	u := strings.TrimSpace(raw)
	if u == "" {
		return ""
	}
	// 兼容绝对地址（https://host/uploads/x → /uploads/x）；无路径部分则放弃
	if i := strings.Index(u, "://"); i >= 0 {
		rest := u[i+3:]
		j := strings.IndexByte(rest, '/')
		if j < 0 {
			return ""
		}
		u = rest[j:]
	}
	// 去掉查询串/锚点
	if i := strings.IndexAny(u, "?#"); i >= 0 {
		u = u[:i]
	}
	if !strings.HasPrefix(u, "/uploads/") && !strings.HasPrefix(u, "/static/") {
		return ""
	}
	// 以 / 结尾是目录形态（如 "/uploads/"），不是文件
	if strings.HasSuffix(u, "/") {
		return ""
	}
	id := u
	if i := strings.LastIndex(u, "/"); i >= 0 {
		id = u[i+1:]
	}
	return id
}
