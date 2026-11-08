package handler

import (
	"encoding/json"
	"library_api/internal/model"
	"library_api/internal/utils"
	"net/http"
	"strconv"
)

func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	book, err := h.book.CreateBook(r.Context(), req)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}
	utils.SuccessResponse(w, http.StatusCreated, "book created", book)
}

func (h *Handler) GetBooks(w http.ResponseWriter, r *http.Request) {
	var (
		authorId  int64
		available bool
		search    string
	)

	page, err := utils.QueryInt(r, "page", utils.DefaultPage)
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	limit, err := utils.QueryInt(r, "limit", utils.DefaultLimit)
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if authorID := r.URL.Query().Get("author_id"); authorID != "" {
		aid, err := strconv.ParseInt(authorID, 10, 64)
		if err != nil {
			utils.ErrorResponse(w, r, http.StatusBadRequest, "invalid author_id parameter")
			return
		}
		authorId = aid
	}

	available, err = utils.QueryBool(r, "available", false)
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, "invalid available parameter")
		return
	}

	if s := r.URL.Query().Get("search"); s != "" {
		search = s
	}

	books, err := h.book.GetBooks(r.Context(), authorId, available, page, limit, search)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}
	utils.SuccessResponse(w, http.StatusOK, "books fetched successfully", books)
}

func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	id, err := utils.PathID(r, "id")

	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	book, err := h.book.GetBookByID(r.Context(), id)

	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "book fetched successfully", book)

}

func (h *Handler) GetBookBorrowers(w http.ResponseWriter, r *http.Request) {
	var includeReturned bool

	id, err := utils.PathID(r, "id")

	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	includeReturned, err = utils.QueryBool(r, "include_returned", false)
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, "invalid include_returned parameter")
		return
	}

	borrowers, err := h.book.GetBookBorrowers(r.Context(), id, includeReturned)

	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "book borrowers fetched successfully", borrowers)

}
