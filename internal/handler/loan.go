package handler

import (
	"encoding/json"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/utils"
	"net/http"
)

func (h *Handler) CreateLoan(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(utils.UserClaimsKey).(*model.Claims)
	if !ok || claims == nil {
		utils.ErrorHandler(w, r, apperror.ErrUnauthorized)
		return
	}

	if claims.MemberID == nil {
		utils.ErrorHandler(w, r, apperror.ErrForbidden)
		return
	}

	var req model.LoanReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.MemberID = *claims.MemberID

	loan, err := h.loan.BorrowBook(r.Context(), req)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, "loan created successfully!", loan)
}

func (h *Handler) ReturnLoan(w http.ResponseWriter, r *http.Request) {
	id, err := utils.PathID(r, "id")
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	loan, err := h.loan.ReturnBook(r.Context(), id)

	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "loan returned successfully!", loan)
}

func (h *Handler) GetLoans(w http.ResponseWriter, r *http.Request) {
	var overdue bool

	limit, err := utils.QueryInt(r, "limit", utils.DefaultLimit)
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	page, err := utils.QueryInt(r, "page", utils.DefaultPage)
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	overdue, err = utils.QueryBool(r, "overdue", false)
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, "invalid overdue parameter")
		return
	}

	loans, err := h.loan.GetLoans(r.Context(), page, limit, overdue)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "loans fetched successfully!", loans)

}

func (h *Handler) GetLoan(w http.ResponseWriter, r *http.Request) {
	id, err := utils.PathID(r, "id")
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	loan, err := h.loan.GetLoan(r.Context(), id)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "loan fetched successfully!", loan)

}
