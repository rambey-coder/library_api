package handler

import (
	"library_api/internal/utils"
	"net/http"
	"strconv"
)

func (h *Handler) GetTopAuthors(w http.ResponseWriter, r *http.Request) {
	limit, err := utils.QueryInt(r, "limit", utils.DefaultTopAuthorsLimit)
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	topAuthors, err := h.report.GetTopAuthors(r.Context(), limit)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}
	utils.SuccessResponse(w, http.StatusOK, "top authors fetched successfully", topAuthors)

}

func (h *Handler) GetAuthorsWithMoreThanNBooks(w http.ResponseWriter, r *http.Request) {

	numberOfBooks, err := utils.QueryInt(r, "no_of_books", 0)
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	authors, err := h.report.GetAuthorsWithMoreThanNBooks(r.Context(), numberOfBooks)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}
	utils.SuccessResponse(w, http.StatusOK, "authors with more than "+strconv.Itoa(numberOfBooks)+" books fetched successfully", authors)

}

func (h *Handler) GetBooksNeverBorrowed(w http.ResponseWriter, r *http.Request) {

	books, err := h.report.GetBooksNeverBorrowed(r.Context())
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}
	utils.SuccessResponse(w, http.StatusOK, "data fetched successfully", books)

}
