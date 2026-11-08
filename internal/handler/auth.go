package handler

import (
	"encoding/json"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/utils"
	"net/http"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	user, err := h.user.Login(r.Context(), req)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "login successfull", user)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	user, err := h.user.CreateUser(r.Context(), req)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, "registration successfull", user)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(utils.UserClaimsKey).(*model.Claims)
	if !ok || claims == nil {
		utils.ErrorResponse(w, r, http.StatusUnauthorized, apperror.ErrUnauthorized.Error())
		return
	}

	user, err := h.user.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "profile retrieved successfully", user)
}
