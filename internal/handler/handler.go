package handler

import (
	"context"
	"library_api/internal/model"
	"library_api/internal/repository"
	"library_api/internal/service"
	"library_api/internal/utils"
	"net/http"
)

type Handler struct {
	repo   *repository.Repository
	author *service.AuthorService
	book   BookService
	member *service.MemberService
	loan   *service.LoanService
	report *service.ReportService
	user   UserService
}

type BookService interface {
	GetBookByID(ctx context.Context, id int64) (*model.Book, error)
	CreateBook(ctx context.Context, book model.CreateBookRequest) (*model.Book, error)
	GetBooks(ctx context.Context, authorID int64, available bool, page, pageSize int, search string) ([]*model.Book, error)
	GetBookBorrowers(ctx context.Context, id int64, includeReturned bool) ([]*model.LoanMember, error)
}

type UserService interface {
	CreateUser(ctx context.Context, user model.CreateUserRequest) (*model.Users, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.UserResponse, error)
	GetProfile(ctx context.Context, id int64) (*model.Users, error)
}

func NewHandler(repo *repository.Repository, author *service.AuthorService, book BookService, member *service.MemberService, loan *service.LoanService, report *service.ReportService, user UserService) *Handler {
	return &Handler{
		repo:   repo,
		author: author,
		book:   book,
		member: member,
		loan:   loan,
		report: report,
		user:   user,
	}
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Ping(r.Context()); err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "Database connection successful", "ok")
}
