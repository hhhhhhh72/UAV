package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- ExpertService (智库专家) ----

type ExpertService struct {
	repo repository.ExpertRepository
}

func NewExpertService(repo repository.ExpertRepository) *ExpertService {
	return &ExpertService{repo: repo}
}

func (s *ExpertService) Create(ctx context.Context, name, title, org, field, bio, avatarURL, status string, tags []string) (domain.Expert, error) {
	now := time.Now()
	e := domain.Expert{
		ID:        fmt.Sprintf("expert-%d", now.UnixNano()),
		Name:      name,
		Title:     title,
		Org:       org,
		Field:     field,
		Bio:       bio,
		AvatarURL: avatarURL,
		Tags:      tags,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if e.Status == "" {
		e.Status = "published"
	}
	return s.repo.Create(ctx, e)
}

func (s *ExpertService) Update(ctx context.Context, id, name, title, org, field, bio, avatarURL, status string, tags []string) (domain.Expert, error) {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Expert{}, err
	}
	e.Name = name
	e.Title = title
	e.Org = org
	e.Field = field
	e.Bio = bio
	e.AvatarURL = avatarURL
	if status != "" {
		e.Status = status
	}
	e.Tags = tags
	e.UpdatedAt = time.Now()
	return s.repo.Update(ctx, e)
}

func (s *ExpertService) List(ctx context.Context, field string) ([]domain.Expert, error) {
	return s.repo.List(ctx, field)
}

func (s *ExpertService) Get(ctx context.Context, id string) (domain.Expert, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ExpertService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ---- CaseService (项目案例) ----

type CaseService struct {
	repo repository.CaseRepository
}

func NewCaseService(repo repository.CaseRepository) *CaseService {
	return &CaseService{repo: repo}
}

func (s *CaseService) Create(ctx context.Context, title, category, description string, images []string, clientName, result string) (domain.CaseEntry, error) {
	now := time.Now()
	c := domain.CaseEntry{
		ID:          fmt.Sprintf("case-%d", now.UnixNano()),
		Title:       title,
		Category:    category,
		Description: description,
		Images:      images,
		ClientName:  clientName,
		Result:      result,
		Status:      "published",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Create(ctx, c)
}

func (s *CaseService) List(ctx context.Context, category string, page, pageSize int) ([]domain.CaseEntry, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, category, offset, pageSize)
}

func (s *CaseService) Get(ctx context.Context, id string) (domain.CaseEntry, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *CaseService) Update(ctx context.Context, id, title, category, description, status string, images []string, clientName, result string) (domain.CaseEntry, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.CaseEntry{}, err
	}
	c.Title = title
	c.Category = category
	c.Description = description
	c.Status = status
	c.Images = images
	c.ClientName = clientName
	c.Result = result
	c.UpdatedAt = time.Now()
	return s.repo.Update(ctx, c)
}

func (s *CaseService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ---- ComplianceService (合规管理) ----

type ComplianceService struct {
	repo repository.ComplianceRepository
}

func NewComplianceService(repo repository.ComplianceRepository) *ComplianceService {
	return &ComplianceService{repo: repo}
}

// complianceDocCategoryAliases: 合规文档英文分类 → 中文规范值。
// 种子数据与小程序/管理端均以中文为准（政策/法规/标准/指南），兼容历史英文传值。
var complianceDocCategoryAliases = map[string]string{
	"policy": "政策", "regulation": "法规", "standard": "标准", "guide": "指南",
}

// complianceStandardCategoryAliases: 团体标准英文分类 → 中文规范值
// （国家标准/行业标准/团体标准/企业标准，与小程序 standards.vue tabs 一致）。
var complianceStandardCategoryAliases = map[string]string{
	"national": "国家标准", "industry": "行业标准", "group": "团体标准", "enterprise": "企业标准",
}

func normalizeComplianceDocCategory(s string) string {
	if v, ok := complianceDocCategoryAliases[strings.TrimSpace(s)]; ok {
		return v
	}
	return strings.TrimSpace(s)
}

func normalizeComplianceStandardCategory(s string) string {
	if v, ok := complianceStandardCategoryAliases[strings.TrimSpace(s)]; ok {
		return v
	}
	return strings.TrimSpace(s)
}

// Docs
func (s *ComplianceService) CreateDoc(ctx context.Context, title, category, publisher, publishDate, status, summary, fileURL string, tags []string) (domain.ComplianceDoc, error) {
	now := time.Now()
	if status == "" {
		status = "published"
	}
	pd, err := time.Parse("2006-01-02", publishDate)
	if err != nil {
		return domain.ComplianceDoc{}, fmt.Errorf("invalid publish date: %w", err)
	}
	d := domain.ComplianceDoc{
		ID:          fmt.Sprintf("compdoc-%d", now.UnixNano()),
		Title:       title,
		Category:    normalizeComplianceDocCategory(category),
		Publisher:   publisher,
		PublishDate: pd,
		Status:      status,
		Summary:     summary,
		FileURL:     fileURL,
		Tags:        tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.CreateDoc(ctx, d)
}

func (s *ComplianceService) ListDocs(ctx context.Context, category string, page, pageSize int) ([]domain.ComplianceDoc, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListDocs(ctx, normalizeComplianceDocCategory(category), offset, pageSize)
}

func (s *ComplianceService) UpdateDoc(ctx context.Context, id, title, category, publisher, publishDate, status, summary, fileURL string, tags []string) (domain.ComplianceDoc, error) {
	d, err := s.repo.FindDocByID(ctx, id)
	if err != nil {
		return domain.ComplianceDoc{}, err
	}
	d.Title = title
	d.Category = normalizeComplianceDocCategory(category)
	d.Publisher = publisher
	pd, err := time.Parse("2006-01-02", publishDate)
	if err != nil {
		slog.Warn("compliance update doc: parse publish date", "publish_date", publishDate, "err", err)
	}
	d.PublishDate = pd
	d.Status = status
	d.Summary = summary
	d.FileURL = fileURL
	d.Tags = tags
	d.UpdatedAt = time.Now()
	return s.repo.UpdateDoc(ctx, d)
}

func (s *ComplianceService) DeleteDoc(ctx context.Context, id string) error {
	return s.repo.DeleteDoc(ctx, id)
}

// Standards
func (s *ComplianceService) CreateStandard(ctx context.Context, title, category, stdNumber, publisher, effectiveDate, status, scope, fileURL string) (domain.StandardDoc, error) {
	now := time.Now()
	if status == "" {
		status = "published"
	}
	pd, err := time.Parse("2006-01-02", effectiveDate)
	if err != nil {
		return domain.StandardDoc{}, fmt.Errorf("invalid effective date: %w", err)
	}
	sd := domain.StandardDoc{
		ID:            fmt.Sprintf("std-%d", now.UnixNano()),
		Title:         title,
		Category:      normalizeComplianceStandardCategory(category),
		StandardNo:    stdNumber,
		Publisher:     publisher,
		EffectiveDate: pd,
		Status:        status,
		Scope:         scope,
		Summary:       "",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return s.repo.CreateStandard(ctx, sd)
}

func (s *ComplianceService) ListStandards(ctx context.Context, category string, page, pageSize int) ([]domain.StandardDoc, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListStandards(ctx, normalizeComplianceStandardCategory(category), offset, pageSize)
}

func (s *ComplianceService) DeleteStandard(ctx context.Context, id string) error {
	return s.repo.DeleteStandard(ctx, id)
}

func (s *ComplianceService) FindDocByID(ctx context.Context, id string) (domain.ComplianceDoc, error) {
	return s.repo.FindDocByID(ctx, id)
}
func (s *ComplianceService) FindStandardByID(ctx context.Context, id string) (domain.StandardDoc, error) {
	return s.repo.FindStandardByID(ctx, id)
}

func (s *ComplianceService) UpdateStandard(ctx context.Context, id, title, category, stdNumber, publisher, effectiveDate, scope, status, fileURL string) (domain.StandardDoc, error) {
	sd, err := s.repo.FindStandardByID(ctx, id)
	if err != nil {
		return domain.StandardDoc{}, err
	}
	sd.Title = title
	sd.Category = normalizeComplianceStandardCategory(category)
	sd.StandardNo = stdNumber
	sd.Publisher = publisher
	sd.Scope = scope
	sd.Status = status
	sd.UpdatedAt = time.Now()
	if effectiveDate != "" {
		if pd, perr := time.Parse("2006-01-02", effectiveDate); perr == nil {
			sd.EffectiveDate = pd
		}
	}
	if fileURL != "" {
		sd.FileURL = fileURL
	}
	return s.repo.UpdateStandard(ctx, sd)
}

// ---- ReportService (行业报告) ----

type ReportService struct {
	repo repository.IndustryReportRepository
}

func NewReportService(repo repository.IndustryReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) Create(ctx context.Context, title, period, category, summary, content, fileURL, author string) (domain.IndustryReport, error) {
	now := time.Now()
	r := domain.IndustryReport{
		ID:        fmt.Sprintf("report-%d", now.UnixNano()),
		Title:     title,
		Period:    period,
		Category:  category,
		Summary:   summary,
		Content:   content,
		FileURL:   fileURL,
		Author:    author,
		Status:    "published",
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.repo.Create(ctx, r)
}

func (s *ReportService) List(ctx context.Context, page, pageSize int) ([]domain.IndustryReport, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

func (s *ReportService) Get(ctx context.Context, id string) (domain.IndustryReport, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ReportService) Update(ctx context.Context, id, title, period, category, summary, content, fileURL, author, status string) (domain.IndustryReport, error) {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.IndustryReport{}, err
	}
	r.Title = title
	r.Period = period
	r.Category = category
	r.Summary = summary
	r.Content = content
	r.FileURL = fileURL
	r.Author = author
	r.Status = status
	r.UpdatedAt = time.Now()
	return s.repo.Update(ctx, r)
}

func (s *ReportService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ---- PortfolioService (企业品牌展示) ----

type PortfolioService struct {
	repo repository.PortfolioRepository
}

func NewPortfolioService(repo repository.PortfolioRepository) *PortfolioService {
	return &PortfolioService{repo: repo}
}

func (s *PortfolioService) Create(ctx context.Context, enterpriseID, name, logoURL, coverURL, description, contactInfo string, products, honors []string) (domain.MemberPortfolio, error) {
	now := time.Now()
	p := domain.MemberPortfolio{
		ID:           fmt.Sprintf("portfolio-%d", now.UnixNano()),
		EnterpriseID: enterpriseID,
		Name:         name,
		LogoURL:      logoURL,
		CoverURL:     coverURL,
		Description:  description,
		Products:     products,
		Honors:       honors,
		ContactInfo:  contactInfo,
		Status:       "draft",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.repo.Create(ctx, p)
}

func (s *PortfolioService) ListPublished(ctx context.Context, page, pageSize int) ([]domain.MemberPortfolio, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPublished(ctx, offset, pageSize)
}

// List 管理端全量列表（含草稿/待审），供 admin 列表页使用。
func (s *PortfolioService) List(ctx context.Context, page, pageSize int) ([]domain.MemberPortfolio, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

func (s *PortfolioService) Get(ctx context.Context, id string) (domain.MemberPortfolio, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *PortfolioService) Update(ctx context.Context, a domain.Actor, id, name, logoURL, coverURL, description, contactInfo, status string, products, honors []string) (domain.MemberPortfolio, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.MemberPortfolio{}, err
	}
	// 越权防护：仅属主（创建者 a.ID 即 EnterpriseID）或管理员可修改
	if !canMutate(a, p.EnterpriseID) {
		return domain.MemberPortfolio{}, ErrNotOwner
	}
	p.Name = name
	p.LogoURL = logoURL
	p.CoverURL = coverURL
	p.Description = description
	p.Products = products
	p.Honors = honors
	p.ContactInfo = contactInfo
	p.Status = status
	p.UpdatedAt = time.Now()
	return s.repo.Update(ctx, p)
}

func (s *PortfolioService) Delete(ctx context.Context, a domain.Actor, id string) error {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	// 越权防护：仅属主或管理员可删除
	if !canMutate(a, p.EnterpriseID) {
		return ErrNotOwner
	}
	return s.repo.Delete(ctx, id)
}

func (s *PortfolioService) ListByEnterprise(ctx context.Context, enterpriseID string) ([]domain.MemberPortfolio, error) {
	return s.repo.ListByEnterprise(ctx, enterpriseID)
}
