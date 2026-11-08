package repository

import (
	"context"
	"errors"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/utils"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) CreateMember(ctx context.Context, val model.CreateMemberRequest) (*model.Member, error) {
	member := &model.Member{}

	query := `
	INSERT INTO members (first_name, last_name, email, phone)
	VALUES($1, $2, $3, $4)
	RETURNING id, first_name, last_name, email, phone, created_at, removed_at
	`

	err := r.db.QueryRow(ctx, query, val.FirstName, val.LastName, val.Email, val.Phone).Scan(
		&member.ID,
		&member.FirstName,
		&member.LastName,
		&member.Email,
		&member.Phone,
		&member.CreatedAt,
		&member.RemovedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "members_email_key" {
				return nil, apperror.ErrMemberEmailExists
			}
		}

		return nil, err
	}

	return member, nil
}

func (r *Repository) GetMembers(ctx context.Context, page, limit int) ([]*model.Member, error) {
	members := []*model.Member{}

	if page <= 0 {
		page = utils.DefaultPage
	}

	if limit <= 0 {
		limit = utils.DefaultLimit
	}

	offset := (page - 1) * limit

	query := `
	SELECT 
    m.id, m.first_name, m.last_name, m.email, m.phone, m.created_at, m.removed_at,
    COUNT(l.id) FILTER (WHERE l.return_date IS NULL) AS open_loan_count
FROM members m
LEFT JOIN loans l ON l.member_id = m.id
WHERE m.removed_at IS NULL
GROUP BY m.id
LIMIT $1 
OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		member := &model.Member{}

		err := rows.Scan(
			&member.ID,
			&member.FirstName,
			&member.LastName,
			&member.Email,
			&member.Phone,
			&member.CreatedAt,
			&member.RemovedAt,
			&member.OpenLoanCount,
		)

		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []*model.Member{}, nil
	}

	return members, nil
}
