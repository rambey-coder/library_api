package service

import (
	"context"
	"library_api/internal/model"
	"library_api/internal/validation"
)

type BookRepo interface {
	GetBookByID(ctx context.Context, id int64) (*model.Book, error)
	CreateBook(ctx context.Context, book model.CreateBookRequest) (*model.Book, error)
	GetBooks(ctx context.Context, authorID int64, available bool, page, pageSize int, search string) ([]*model.Book, error)
	GetBookBorrowers(ctx context.Context, id int64, includeReturned bool) ([]*model.LoanMember, error)
}

type BookService struct {
	repo BookRepo
}

func NewBookService(r BookRepo) *BookService {
	return &BookService{repo: r}
}

func (s *BookService) CreateBook(ctx context.Context, book model.CreateBookRequest) (*model.Book, error) {
	if err := validation.ValidateCreateBookRequest(book); err != nil {
		return nil, err
	}
	return s.repo.CreateBook(ctx, book)
}

func (s *BookService) GetBooks(ctx context.Context, authorID int64, available bool, page, pageSize int, search string) ([]*model.Book, error) {
	return s.repo.GetBooks(ctx, authorID, available, page, pageSize, search)
}

func (s *BookService) GetBookByID(ctx context.Context, id int64) (*model.Book, error) {
	return s.repo.GetBookByID(ctx, id)
}

func (s *BookService) GetBookBorrowers(ctx context.Context, id int64, includeReturned bool) ([]*model.LoanMember, error) {
	return s.repo.GetBookBorrowers(ctx, id, includeReturned)
}
