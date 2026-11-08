package repository

import (
	"context"
	"encoding/json"
	"errors"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/utils"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) GetMemberLoans(ctx context.Context, memberID int64, includeReturned bool) ([]*model.Loan, error) {
	var memberExists bool

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})

	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	checkQuery := `SELECT EXISTS(SELECT 1 FROM members WHERE id = $1 AND removed_at IS NULL)`
	err = tx.QueryRow(ctx, checkQuery, memberID).Scan(&memberExists)
	if err != nil {
		return nil, err
	}
	if !memberExists {
		return nil, apperror.ErrMemberNotFound
	}

	var query strings.Builder
	query.WriteString(`
	SELECT 
		l.id, 
		l.book_id, 
		l.member_id, 
		l.loan_date, 
		l.due_date, 
		l.return_date, 
		l.created_at,
		b.id, 
		b.title, 
		b.isbn, 
		b.published_date, 
		b.genre, 
		b.price, 
		b.created_at, 
		b.deleted_at,
		COALESCE(
			json_agg(
				json_build_object(
					'id', a.id,
					'first_name', a.first_name,
					'last_name', a.last_name,
					'email', a.email,
					'bio', a.bio,
					'created_at', a.created_at,
					'deleted_at', a.deleted_at
				) ORDER BY a.id ASC
			) FILTER (WHERE a.id IS NOT NULL),
			'[]'::json
		) AS authors
	FROM loans l
	JOIN books b ON l.book_id = b.id
	LEFT JOIN book_authors ba ON ba.book_id = b.id
	LEFT JOIN authors a ON ba.author_id = a.id AND a.deleted_at IS NULL
	WHERE l.member_id = $1
	`)

	if !includeReturned {
		query.WriteString(" AND l.return_date IS NULL")
	}

	query.WriteString(`
	GROUP BY l.id, b.id
	ORDER BY l.loan_date DESC
	`)

	rows, err := tx.Query(ctx, query.String(), memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	loans := make([]*model.Loan, 0)
	for rows.Next() {
		loan := &model.Loan{}
		book := &model.Book{}
		var authorsJSON []byte

		err := rows.Scan(
			&loan.ID,
			&loan.BookID,
			&loan.MemberID,
			&loan.LoanDate,
			&loan.DueDate,
			&loan.ReturnDate,
			&loan.CreatedAt,
			&book.ID,
			&book.Title,
			&book.ISBN,
			&book.PublishedDate,
			&book.Genre,
			&book.Price,
			&book.CreatedAt,
			&book.DeletedAt,
			&authorsJSON,
		)
		if err != nil {
			return nil, err
		}

		book.Authors = make([]*model.Author, 0)
		if len(authorsJSON) > 0 {
			if err := json.Unmarshal(authorsJSON, &book.Authors); err != nil {
				return nil, err
			}
		}

		loan.Book = book
		loans = append(loans, loan)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return loans, nil
}

func (r *Repository) BorrowBook(ctx context.Context, val model.LoanReq) (*model.Loan, error) {
	loan := &model.Loan{}
	tx, err := r.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	var bookID int64

	query := `
	SELECT id 
	FROM books 
	WHERE id = $1 
	AND deleted_at IS NULL
	FOR UPDATE
	`

	err = tx.QueryRow(ctx, query, val.BookID).Scan(&bookID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrBookNotFound
		}
		return nil, err
	}

	var memberID int64

	query = `
	SELECT id 
	FROM members 
	WHERE id = $1 
	AND removed_at IS NULL
	FOR UPDATE
	`

	err = tx.QueryRow(ctx, query, val.MemberID).Scan(&memberID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrMemberNotFound
		}
		return nil, err
	}

	var openLoanCount int

	query = `SELECT count(*) FROM loans WHERE member_id = $1 AND return_date IS NULL`
	err = tx.QueryRow(ctx, query, memberID).Scan(&openLoanCount)
	if err != nil {
		return nil, err
	}
	if openLoanCount >= utils.MaxActiveLoans {
		return nil, apperror.ErrMemberHasMaxLoans
	}

	query = `
	INSERT INTO loans (book_id, member_id, loan_date, due_date) 
	VALUES($1, $2, $3, $4)
	RETURNING id, book_id, member_id, loan_date, due_date, return_date, created_at
	`

	err = tx.QueryRow(ctx, query, val.BookID, val.MemberID, val.LoanDate, val.DueDate).Scan(
		&loan.ID,
		&loan.BookID,
		&loan.MemberID,
		&loan.LoanDate,
		&loan.DueDate,
		&loan.ReturnDate,
		&loan.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_one_active_per_book" {
				return nil, apperror.ErrBookAlreadyOnLoan
			}
		}
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return loan, nil

}

func (r *Repository) ReturnBook(ctx context.Context, loanID int64) (*model.Loan, error) {
	loan := &model.Loan{}
	var loanReturnDate *time.Time

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	query := `
	SELECT return_date
	FROM loans
	WHERE id = $1
	FOR UPDATE
	`

	err = tx.QueryRow(ctx, query, loanID).Scan(&loanReturnDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrLoanNotFound
		}
		return nil, err
	}
	if loanReturnDate != nil {
		return nil, apperror.ErrLoanAlreadyReturned
	}

	query = `
	UPDATE loans
	SET return_date = NOW()
	WHERE id = $1
	RETURNING id, book_id, member_id, loan_date, due_date, return_date, created_at
	`

	err = tx.QueryRow(ctx, query, loanID).Scan(
		&loan.ID,
		&loan.BookID,
		&loan.MemberID,
		&loan.LoanDate,
		&loan.DueDate,
		&loan.ReturnDate,
		&loan.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return loan, nil
}

// open + past-due
func (r *Repository) GetLoans(ctx context.Context, page, limit int, overdue bool) ([]*model.Loan, error) {
	loans := []*model.Loan{}
	var query strings.Builder

	if page <= 0 {
		page = utils.DefaultPage
	}

	if limit <= 0 {
		limit = utils.DefaultLimit
	}

	offset := (page - 1) * limit

	query.WriteString(`
	
	SELECT 
	l.id, 
	b.title,
	m.first_name, m.last_name,
	l.loan_date,
	l.due_date,
	l.return_date
	FROM loans l
	JOIN books b ON l.book_id = b.id
	JOIN members m ON l.member_id = m.id
	WHERE l.return_date IS NULL
	`)

	if overdue {
		query.WriteString("AND l.due_date < NOW()")
	}

	query.WriteString(`
	GROUP BY l.id, b.id, m.id
	ORDER BY l.loan_date DESC
	LIMIT $1 OFFSET $2
	`)

	rows, err := r.db.Query(ctx, query.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		loan := &model.Loan{}
		book := &model.Book{}
		member := &model.Member{}

		err := rows.Scan(
			&loan.ID,
			&book.Title,
			&member.FirstName,
			&member.LastName,
			&loan.LoanDate,
			&loan.DueDate,
			&loan.ReturnDate,
		)
		if err != nil {
			return nil, err
		}

		loan.Book = book
		loan.Member = member
		loans = append(loans, loan)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return loans, nil
}

func (r *Repository) GetLoan(ctx context.Context, loanID int64) (*model.Loan, error) {
	loan := &model.Loan{}
	loan.Book = &model.Book{}
	loan.Member = &model.Member{}

	query := `
	SELECT 
	l.id, 
	b.title,
	m.first_name, m.last_name,
	l.loan_date,
	l.due_date,
	l.return_date
	FROM loans l
	JOIN books b ON l.book_id = b.id
	JOIN members m ON l.member_id = m.id
	WHERE l.id = $1
	`

	err := r.db.QueryRow(ctx, query, loanID).Scan(
		&loan.ID,
		&loan.Book.Title,
		&loan.Member.FirstName,
		&loan.Member.LastName,
		&loan.LoanDate,
		&loan.DueDate,
		&loan.ReturnDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrLoanNotFound
		}
		return nil, err
	}

	return loan, nil
}
