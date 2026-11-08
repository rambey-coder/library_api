package service

import (
	"context"
	"library_api/internal/model"
	"library_api/internal/repository"
	"library_api/internal/validation"
	"time"
)

type LoanService struct {
	repo *repository.Repository
}

func NewLoanService(r *repository.Repository) *LoanService {
	return &LoanService{repo: r}
}

func (s *LoanService) GetLoans(ctx context.Context, page, limit int, overdue bool) ([]*model.Loan, error) {
	return s.repo.GetLoans(ctx, page, limit, overdue)
}

func (s *LoanService) GetLoan(ctx context.Context, loanID int64) (*model.Loan, error) {
	return s.repo.GetLoan(ctx, loanID)
}

func (s *LoanService) BorrowBook(ctx context.Context, loanReq model.LoanReq) (*model.Loan, error) {
	if err := validation.ValidateLoanRequest(loanReq, time.Now()); err != nil {
		return nil, err
	}

	return s.repo.BorrowBook(ctx, loanReq)
}

func (s *LoanService) ReturnBook(ctx context.Context, loanID int64) (*model.Loan, error) {
	if err := validation.ValidateLoanReturn(loanID); err != nil {
		return nil, err
	}

	return s.repo.ReturnBook(ctx, loanID)
}
