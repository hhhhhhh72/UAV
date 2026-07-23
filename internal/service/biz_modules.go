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

func (s *ExpertService) Create(name, title, org, field, bio string, tags []string) (domain.Expert, error) {
	now := time.Now()
	e := domain.Expert{
		ID:        fmt.Sprintf("expert-%d", now.UnixNano()),
		Name:      name,
		Title:     title,
		Org:       org,
		Field:     field,
		Bio:       bio,
		Tags:      tags,
		Status:    "published",
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.repo.Create(e)
}

func (s *ExpertService) Update(id, name, title, org, field, bio string, tags []string) (domain.Expert, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Expert{}, err
	}
	e.Name = name
	e.Title = title
	e.Org = org
	e.Field = field
	e.Bio = bio
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

func (s *CaseService) Update(id, title, category, description string, images []string, clientName, result string) (domain.CaseEntry, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return domain.CaseEntry{}, err
	}
	c.Title = title
	c.Category = category
	c.Description = description
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
func (s *ComplianceService) CreateDoc(title, category, content, summary, source string, tags []string) (domain.ComplianceDoc, error) {
	now := time.Now()
	d := domain.ComplianceDoc{
		ID:        fmt.Sprintf("compdoc-%d", now.UnixNano()),
		Title:     title,
		Category:  category,
		Content:   content,
		Summary:   summary,
		Source:    source,
		Tags:      tags,
		Status:    "published",
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.repo.CreateDoc(d)
}

func (s *ComplianceService) ListDocs(category string, page, pageSize int) ([]domain.ComplianceDoc, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListDocs(category, offset, pageSize)
}

func (s *ComplianceService) UpdateDoc(id, title, category, content, summary, source string, tags []string) (domain.ComplianceDoc, error) {
	d, err := s.repo.FindDocByID(id)
	if err != nil {
		return domain.ComplianceDoc{}, err
	}
	d.Title = title
	d.Category = category
	d.Content = content
	d.Summary = summary
	d.Source = source
	d.Tags = tags
	d.UpdatedAt = time.Now()
	return s.repo.UpdateDoc(d)
}

func (s *ComplianceService) DeleteDoc(id string) error {
	return s.repo.DeleteDoc(id)
}

// Standards
func (s *ComplianceService) CreateStandard(title, stdNumber, category, version, publisher, content, fileURL string, issueDate time.Time) (domain.StandardDoc, error) {
	now := time.Now()
	sd := domain.StandardDoc{
		ID:        fmt.Sprintf("std-%d", now.UnixNano()),
		Title:     title,
		StdNumber: stdNumber,
		Category:  category,
		Version:   version,
		IssueDate: issueDate,
		Publisher: publisher,
		Content:   content,
		FileURL:   fileURL,
		Status:    "published",
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.repo.CreateStandard(sd)
}

func (s *ComplianceService) ListStandards(category string, page, pageSize int) ([]domain.StandardDoc, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListStandards(category, offset, pageSize)
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

func (s *ReportService) Update(id, title, period, category, summary, content, fileURL, author string) (domain.IndustryReport, error) {
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

func (s *PortfolioService) Get(id string) (domain.MemberPortfolio, error) {
	return s.repo.FindByID(id)
}

func (s *PortfolioService) Update(id, name, logoURL, coverURL, description, contactInfo string, products, honors []string) (domain.MemberPortfolio, error) {
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
	p.UpdatedAt = time.Now()
	return s.repo.Update(p)
}

func (s *PortfolioService) ListByEnterprise(enterpriseID string) ([]domain.MemberPortfolio, error) {
	return s.repo.ListByEnterprise(enterpriseID)
}
