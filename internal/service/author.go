package service

import (
	"context"
	"library_api/internal/model"
	"library_api/internal/repository"
	"library_api/internal/validation"
)

type AuthorService struct {
	repo *repository.Repository
}

func NewAuthorService(r *repository.Repository) *AuthorService {
	return &AuthorService{repo: r}
}

func (s *AuthorService) CreateAuthor(ctx context.Context, author model.CreateAuthorRequest) (*model.Author, error) {
	if err := validation.ValidateCreateAuthorRequest(author); err != nil {
		return nil, err
	}
	return s.repo.CreateAuthor(ctx, author)
}

func (s *AuthorService) GetAuthors(ctx context.Context) ([]*model.Author, error) {
	return s.repo.GetAuthors(ctx)
}

func (s *AuthorService) GetAuthorByID(ctx context.Context, id int64) (*model.Author, error) {
	return s.repo.GetAuthorByID(ctx, id)
}

func (s *AuthorService) GetBooksByAuthorID(ctx context.Context, id int64) ([]*model.Book, error) {
	return s.repo.GetBooksByAuthorID(ctx, id)
}
