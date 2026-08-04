package service

import (
	"fmt"
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

func (s *ExpertService) Create(name, title, org, field, bio, avatarURL, status string, tags []string) (domain.Expert, error) {
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
	return s.repo.Create(e)
}

func (s *ExpertService) Update(id, name, title, org, field, bio, avatarURL, status string, tags []string) (domain.Expert, error) {
	e, err := s.repo.FindByID(id)
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
	return s.repo.Update(e)
}

func (s *ExpertService) List(field string) ([]domain.Expert, error) {
	return s.repo.List(field)
}

func (s *ExpertService) Get(id string) (domain.Expert, error) {
	return s.repo.FindByID(id)
}

func (s *ExpertService) Delete(id string) error {
	return s.repo.Delete(id)
}

// ---- CaseService (项目案例) ----

type CaseService struct {
	repo repository.CaseRepository
}

func NewCaseService(repo repository.CaseRepository) *CaseService {
	return &CaseService{repo: repo}
}

func (s *CaseService) Create(title, category, description string, images []string, clientName, result string) (domain.CaseEntry, error) {
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
	return s.repo.Create(c)
}

func (s *CaseService) List(category string, page, pageSize int) ([]domain.CaseEntry, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(category, offset, pageSize)
}

func (s *CaseService) Get(id string) (domain.CaseEntry, error) {
	return s.repo.FindByID(id)
}

func (s *CaseService) Update(id, title, category, description, status string, images []string, clientName, result string) (domain.CaseEntry, error) {
	c, err := s.repo.FindByID(id)
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
	return s.repo.Update(c)
}

func (s *CaseService) Delete(id string) error {
	return s.repo.Delete(id)
}

// ---- ComplianceService (合规管理) ----

type ComplianceService struct {
	repo repository.ComplianceRepository
}

func NewComplianceService(repo repository.ComplianceRepository) *ComplianceService {
	return &ComplianceService{repo: repo}
}

// Docs
func (s *ComplianceService) CreateDoc(title, category, publisher, publishDate, status, summary, fileURL string, tags []string) (domain.ComplianceDoc, error) {
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
		Category:    category,
		Publisher:   publisher,
		PublishDate: pd,
		Status:      status,
		Summary:     summary,
		FileURL:     fileURL,
		Tags:        tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.CreateDoc(d)
}

func (s *ComplianceService) ListDocs(category string, page, pageSize int) ([]domain.ComplianceDoc, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListDocs(category, offset, pageSize)
}

func (s *ComplianceService) UpdateDoc(id, title, category, publisher, publishDate, status, summary, fileURL string, tags []string) (domain.ComplianceDoc, error) {
	d, err := s.repo.FindDocByID(id)
	if err != nil {
		return domain.ComplianceDoc{}, err
	}
	d.Title = title
	d.Category = category
	d.Publisher = publisher
	pd, _ := time.Parse("2006-01-02", publishDate)
	d.PublishDate = pd
	d.Status = status
	d.Summary = summary
	d.FileURL = fileURL
	d.Tags = tags
	d.UpdatedAt = time.Now()
	return s.repo.UpdateDoc(d)
}

func (s *ComplianceService) DeleteDoc(id string) error {
	return s.repo.DeleteDoc(id)
}

// Standards
func (s *ComplianceService) CreateStandard(title, stdNumber, publisher, effectiveDate, status, scope, fileURL string) (domain.StandardDoc, error) {
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
		StandardNo:    stdNumber,
		Publisher:     publisher,
		EffectiveDate: pd,
		Status:        status,
		Scope:         scope,
		Summary:       "",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return s.repo.CreateStandard(sd)
}

func (s *ComplianceService) ListStandards(category string, page, pageSize int) ([]domain.StandardDoc, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListStandards(category, offset, pageSize)
}

func (s *ComplianceService) DeleteStandard(id string) error { return s.repo.DeleteStandard(id) }

func (s *ComplianceService) FindDocByID(id string) (domain.ComplianceDoc, error) {
	return s.repo.FindDocByID(id)
}
func (s *ComplianceService) FindStandardByID(id string) (domain.StandardDoc, error) {
	return s.repo.FindStandardByID(id)
}

func (s *ComplianceService) UpdateStandard(id, title, status, fileURL string) (domain.StandardDoc, error) {
	sd, err := s.repo.FindStandardByID(id)
	if err != nil {
		return domain.StandardDoc{}, err
	}
	sd.Title = title
	sd.Status = status
	sd.UpdatedAt = time.Now()
	_ = fileURL
	return s.repo.UpdateStandard(sd)
}

// ---- ReportService (行业报告) ----

type ReportService struct {
	repo repository.IndustryReportRepository
}

func NewReportService(repo repository.IndustryReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) Create(title, period, category, summary, content, fileURL, author string) (domain.IndustryReport, error) {
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
	return s.repo.Create(r)
}

func (s *ReportService) List(page, pageSize int) ([]domain.IndustryReport, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(offset, pageSize)
}

func (s *ReportService) Get(id string) (domain.IndustryReport, error) {
	return s.repo.FindByID(id)
}

func (s *ReportService) Update(id, title, period, category, summary, content, fileURL, author, status string) (domain.IndustryReport, error) {
	r, err := s.repo.FindByID(id)
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
	return s.repo.Update(r)
}

func (s *ReportService) Delete(id string) error {
	return s.repo.Delete(id)
}

// ---- PortfolioService (企业品牌展示) ----

type PortfolioService struct {
	repo repository.PortfolioRepository
}

func NewPortfolioService(repo repository.PortfolioRepository) *PortfolioService {
	return &PortfolioService{repo: repo}
}

func (s *PortfolioService) Create(enterpriseID, name, logoURL, coverURL, description, contactInfo string, products, honors []string) (domain.MemberPortfolio, error) {
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
	return s.repo.Create(p)
}

func (s *PortfolioService) ListPublished(page, pageSize int) ([]domain.MemberPortfolio, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPublished(offset, pageSize)
}

// List 管理端全量列表（含草稿/待审），供 admin 列表页使用。
func (s *PortfolioService) List(page, pageSize int) ([]domain.MemberPortfolio, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(offset, pageSize)
}

func (s *PortfolioService) Get(id string) (domain.MemberPortfolio, error) {
	return s.repo.FindByID(id)
}

func (s *PortfolioService) Update(id, name, logoURL, coverURL, description, contactInfo, status string, products, honors []string) (domain.MemberPortfolio, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return domain.MemberPortfolio{}, err
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
	return s.repo.Update(p)
}

func (s *PortfolioService) Delete(id string) error { return s.repo.Delete(id) }

func (s *PortfolioService) ListByEnterprise(enterpriseID string) ([]domain.MemberPortfolio, error) {
	return s.repo.ListByEnterprise(enterpriseID)
}

// ── ShopService ──────────────────────────────────────────

type ShopService struct {
	repo repository.ShopRepository
}

func NewShopService(repo repository.ShopRepository) *ShopService {
	return &ShopService{repo: repo}
}

func (s *ShopService) Create(name, licenseURL, accountName, contactPhone, address, status string, isMember bool) (domain.Shop, error) {
	now := time.Now()
	shop := domain.Shop{
		ID:           fmt.Sprintf("shop-%d", now.UnixNano()),
		Name:         name,
		LicenseURL:   licenseURL,
		AccountName:  accountName,
		ContactPhone: contactPhone,
		Address:      address,
		Status:       status,
		IsMember:     isMember,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if shop.Status == "" {
		shop.Status = "pending"
	}
	return s.repo.Create(shop)
}

func (s *ShopService) Update(id, name, licenseURL, accountName, contactPhone, address, status string, isMember bool) (domain.Shop, error) {
	shop, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Shop{}, err
	}
	shop.Name = name
	shop.LicenseURL = licenseURL
	shop.AccountName = accountName
	shop.ContactPhone = contactPhone
	shop.Address = address
	if status != "" {
		shop.Status = status
	}
	shop.IsMember = isMember
	shop.UpdatedAt = time.Now()
	return s.repo.Update(shop)
}

func (s *ShopService) List(offset, limit int) ([]domain.Shop, int, error) {
	return s.repo.List(offset, limit)
}

func (s *ShopService) Delete(id string) error {
	return s.repo.Delete(id)
}
