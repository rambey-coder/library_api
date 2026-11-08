package service

import (
	"context"
	"library_api/internal/model"
	"library_api/internal/repository"
	"library_api/internal/validation"
)

type MemberService struct {
	repo *repository.Repository
}

func NewMemberService(r *repository.Repository) *MemberService {
	return &MemberService{repo: r}
}

func (s *MemberService) CreateMember(ctx context.Context, member model.CreateMemberRequest) (*model.Member, error) {

	if err := validation.ValidateCreateMemberRequest(member); err != nil {
		return nil, err
	}
	return s.repo.CreateMember(ctx, member)
}

func (s *MemberService) GetMembers(ctx context.Context, page, pageSize int) ([]*model.Member, error) {
	return s.repo.GetMembers(ctx, page, pageSize)
}

func (s *MemberService) GetMemberLoans(ctx context.Context, memberID int64, includeReturned bool) ([]*model.Loan, error) {
	return s.repo.GetMemberLoans(ctx, memberID, includeReturned)
}
