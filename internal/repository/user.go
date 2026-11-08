package repository

import (
	"context"
	"errors"
	"library_api/internal/apperror"
	"library_api/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) RegisterUser(ctx context.Context, req model.CreateUserRequest, pass_hash string) (*model.Users, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO users 
	(
	 first_name,
	 last_name,
	 email,
	 phone,
	 password_hash,
	 role,
	 gender
	 )
	 VALUES ($1, $2, $3, $4, $5, 'member', $6)
	 RETURNING id, first_name, last_name, email, role, phone, gender, created_at, deleted_at, password_hash
	`

	user := model.Users{}

	err = tx.QueryRow(
		ctx,
		query,
		req.FirstName,
		req.LastName,
		req.Email,
		req.Phone,
		pass_hash,
		req.Gender,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Role,
		&user.Phone,
		&user.Gender,
		&user.CreatedAt,
		&user.DeletedAt,
		&user.PasswordHash,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
				return nil, apperror.ErrEmailTaken
			}

			if pgErr.ConstraintName == "users_role_check" {
				return nil, apperror.ErrInvalidRole
			}

			if pgErr.ConstraintName == "users_gender_check" {
				return nil, apperror.ErrInvalidGender
			}

		}
		return nil, err
	}

	memberQuery := `
	INSERT INTO members (
		user_id, 
		first_name, 
		last_name, 
		email, 
		phone
	) 
	VALUES ($1, $2, $3, $4, $5) 
	RETURNING id
	`

	err = tx.QueryRow(
		ctx,
		memberQuery,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Phone,
	).Scan(&user.MemberID)

	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.Users, error) {
	query := `
	SELECT u.id, u.first_name, u.last_name, u.email, u.role, u.phone, u.gender, u.created_at, u.deleted_at, u.password_hash, m.id AS member_id
	FROM users AS u
	LEFT JOIN members AS m 
	ON u.id = m.user_id
	WHERE u.email = $1
	AND u.deleted_at IS NULL
	`

	user := model.Users{}

	err := r.db.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Role,
		&user.Phone,
		&user.Gender,
		&user.CreatedAt,
		&user.DeletedAt,
		&user.PasswordHash,
		&user.MemberID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrInvalidCredentials
		}
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*model.Users, error) {
	query := `
	SELECT u.id, u.first_name, u.last_name, u.email, u.role, u.phone, u.gender, u.created_at, u.deleted_at, u.password_hash, m.id AS member_id
	FROM users AS u
	LEFT JOIN members AS m 
	ON u.id = m.user_id
	WHERE u.id = $1
	AND u.deleted_at IS NULL
	`

	user := model.Users{}

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Role,
		&user.Phone,
		&user.Gender,
		&user.CreatedAt,
		&user.DeletedAt,
		&user.PasswordHash,
		&user.MemberID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
