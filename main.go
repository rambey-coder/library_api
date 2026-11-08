package main

import (
	"context"
	"library_api/internal/db"
	"library_api/internal/handler"
	"library_api/internal/middleware"
	"library_api/internal/repository"
	"library_api/internal/service"
	"library_api/internal/utils"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env")
	}
	ctx := context.Background()

	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	mux := http.NewServeMux()

	repo := repository.New(pool)
	authorService := service.NewAuthorService(repo)
	bookService := service.NewBookService(repo)
	memberService := service.NewMemberService(repo)
	loanService := service.NewLoanService(repo)
	reportService := service.NewReportService(repo)
	userService := service.NewUserService(repo)

	h := handler.NewHandler(repo, authorService, bookService, memberService, loanService, reportService, userService)

	mux.HandleFunc("GET /health", h.HealthHandler)

	mux.Handle("POST /authors", middleware.Protect(h.CreateAuthor, utils.RoleAdmin, utils.RoleLibrarian))
	mux.Handle("GET /authors", middleware.Protect(h.GetAuthors))
	mux.Handle("GET /authors/{id}", middleware.Protect(h.GetAuthor))
	mux.Handle("GET /authors/{id}/books", middleware.Protect(h.GetAuthorBooks))

	mux.Handle("POST /books", middleware.Protect(h.CreateBook, utils.RoleAdmin, utils.RoleLibrarian))
	mux.Handle("GET /books", middleware.Protect(h.GetBooks))
	mux.Handle("GET /books/{id}", middleware.Protect(h.GetBook))
	mux.Handle("GET /books/{id}/borrowers", middleware.Protect(h.GetBookBorrowers, utils.RoleAdmin, utils.RoleLibrarian))

	mux.Handle("POST /members", middleware.Protect(h.CreateMember, utils.RoleAdmin, utils.RoleLibrarian))
	mux.Handle("GET /members", middleware.Protect(h.GetMembers, utils.RoleAdmin, utils.RoleLibrarian))
	mux.Handle("GET /members/{id}/loans", middleware.Protect(h.GetMemberLoans, utils.RoleAdmin, utils.RoleLibrarian, utils.RoleMember))

	mux.Handle("POST /loans", middleware.Protect(h.CreateLoan, utils.RoleMember))
	mux.Handle("GET /loans", middleware.Protect(h.GetLoans, utils.RoleAdmin, utils.RoleLibrarian))
	mux.Handle("GET /loans/{id}", middleware.Protect(h.GetLoan, utils.RoleAdmin, utils.RoleLibrarian))
	mux.Handle("PATCH /loans/{id}", middleware.Protect(h.ReturnLoan, utils.RoleAdmin, utils.RoleLibrarian))

	mux.Handle("GET /reports/top-authors", middleware.Protect(h.GetTopAuthors, utils.RoleAdmin, utils.RoleLibrarian))
	mux.Handle("GET /reports/authors-with-more-than", middleware.Protect(h.GetAuthorsWithMoreThanNBooks, utils.RoleAdmin, utils.RoleLibrarian))
	mux.Handle("GET /reports/books-never-borrowed", middleware.Protect(h.GetBooksNeverBorrowed, utils.RoleAdmin, utils.RoleLibrarian))

	mux.HandleFunc("POST /auth/login", h.Login)
	mux.HandleFunc("POST /auth/register", h.Register)
	mux.Handle("GET /auth/me", middleware.Protect(h.GetProfile))

	timeoutMW := middleware.Timeout(5 * time.Second)

	finalHandler := middleware.Recover(
		middleware.RequestID(
			middleware.Logging(
				timeoutMW(mux),
			),
		),
	)

	srv := &http.Server{
		Addr:    ":8082",
		Handler: finalHandler,
	}
	log.Fatal(srv.ListenAndServe())
}
