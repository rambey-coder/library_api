package handler

import (
	"encoding/json"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/utils"
	"net/http"
)

func (h *Handler) CreateMember(w http.ResponseWriter, r *http.Request) {
	var req model.CreateMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	member, err := h.member.CreateMember(r.Context(), req)
	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, "member created successfully!", member)
}

func (h *Handler) GetMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.member.GetMembers(r.Context(), page, limit)

	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "members fetched successfully!", members)
}

func (h *Handler) GetMemberLoans(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(utils.UserClaimsKey).(*model.Claims)
	if !ok || claims == nil {
		utils.ErrorHandler(w, r, apperror.ErrUnauthorized)
		return
	}

	var includeReturned bool
	id, err := utils.PathID(r, "id")

	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if claims.Role == utils.RoleMember {
		if claims.MemberID == nil || *claims.MemberID != id {
			utils.ErrorHandler(w, r, apperror.ErrForbidden)
			return
		}
	}

	includeReturned, err = utils.QueryBool(r, "include_returned", false)
	if err != nil {
		utils.ErrorResponse(w, r, http.StatusBadRequest, "invalid include_returned parameter")
		return
	}

	loans, err := h.member.GetMemberLoans(r.Context(), id, includeReturned)

	if err != nil {
		utils.ErrorHandler(w, r, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, "member loans fetched successfully!", loans)
}
