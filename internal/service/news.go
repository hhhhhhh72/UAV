package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type NewsService struct {
	repo repository.ArticleRepository
}

func NewNewsService(repo repository.ArticleRepository) *NewsService {
	return &NewsService{repo: repo}
}

// htmlTagRe 摘要剥离用：正文为富文本 HTML，列表摘要必须为纯文本
//（直接截取 HTML 会把 <p> 等残缺标签带进小程序列表/详情兜底展示）。
var htmlTagRe = regexp.MustCompile(`(?s)<[^>]*>`)

// summaryFromContent 富文本 → 纯文本摘要：剥标签 + 还原常见实体 + 截断 100 字。
func summaryFromContent(content string) string {
	s := htmlTagRe.ReplaceAllString(content, "")
	s = strings.NewReplacer(
		"&nbsp;", " ",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&#39;", "'",
	).Replace(s)
	s = strings.Join(strings.Fields(s), " ") // 压缩空白（HTML 排版产生的换行/多空格）
	return truncate(s, 100)
}

func (s *NewsService) Create(ctx context.Context, title, content, category, source, author string, isPinned bool) (domain.Article, error) {
	now := time.Now()
	a := domain.Article{ID: nextID("article"), Title: title, Content: content,
		Summary: summaryFromContent(content), Category: category, Source: source,
		Author: strings.TrimSpace(author), IsPinned: isPinned, Status: "draft",
		Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.repo.Create(ctx, a)
}

// Update 编辑资讯内容（保留 ID/状态/创建时间，摘要重新截取）。
func (s *NewsService) Update(ctx context.Context, id, title, content, category, source, author string, isPinned bool) (domain.Article, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	a.Title = title
	a.Content = content
	a.Summary = summaryFromContent(content)
	a.Category = category
	a.Source = source
	a.Author = strings.TrimSpace(author)
	a.IsPinned = isPinned
	a.UpdatedAt = time.Now()
	return s.repo.Update(ctx, a)
}

func (s *NewsService) Publish(ctx context.Context, id string) (domain.Article, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	a.Status = "published"
	a.UpdatedAt = time.Now()
	return s.repo.Update(ctx, a)
}

// Delete 删除资讯（草稿/已发布均可；列表/详情由查询侧自然消失）。
func (s *NewsService) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *NewsService) ListByCategory(ctx context.Context, category string, page, pageSize int) ([]domain.Article, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByCategory(ctx, category, offset, pageSize)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
