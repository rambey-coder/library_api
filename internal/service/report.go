package service

import (
	"context"
	"library_api/internal/model"
	"library_api/internal/repository"
)

type ReportService struct {
	repo *repository.Repository
}

func NewReportService(repo *repository.Repository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) GetTopAuthors(ctx context.Context, limit int) ([]*model.AuthorBookCount, error) {
	return s.repo.GetTopAuthors(ctx, limit)
}

func (s *ReportService) GetAuthorsWithMoreThanNBooks(ctx context.Context, n int) ([]*model.AuthorBookCount, error) {
	return s.repo.GetAuthorsWithMoreThanNBooks(ctx, n)
}

func (s *ReportService) GetBooksNeverBorrowed(ctx context.Context) ([]*model.Book, error) {
	return s.repo.GetBooksNeverBorrowed(ctx)
}
