package handler

import (
	"encoding/json"
	"library_api/internal/model"
	"library_api/internal/utils"
	"net/http"
)

func (h *Handler) CreateAuthor(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	author, err := h.author.CreateAuthor(r.Context(), req)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, "author created successfully!", author)
}

func (h *Handler) GetAuthors(w http.ResponseWriter, r *http.Request) {
	authors, err := h.author.GetAuthors(r.Context())
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "authors fetched successfully!", authors)
}

func (h *Handler) GetAuthor(w http.ResponseWriter, r *http.Request) {
	id, err := utils.PathID(r, "id")

	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}
	author, err := h.author.GetAuthorByID(r.Context(), id)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "author fetched successfully!", author)
}

func (h *Handler) GetAuthorBooks(w http.ResponseWriter, r *http.Request) {
	id, err := utils.PathID(r, "id")

	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}
	authorBooks, err := h.author.GetBooksByAuthorID(r.Context(), id)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "author books fetched successfully!", authorBooks)
}
