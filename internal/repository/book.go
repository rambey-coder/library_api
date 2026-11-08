package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/utils"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) CreateBook(ctx context.Context, val model.CreateBookRequest) (*model.Book, error) {
	seen := make(map[int64]struct{})
	uniqueAuthorIDs := make([]int64, 0, len(val.AuthorIds))

	for _, id := range val.AuthorIds {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			uniqueAuthorIDs = append(uniqueAuthorIDs, id)
		}
	}
	val.AuthorIds = uniqueAuthorIDs

	tx, err := r.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	book := &model.Book{}

	var authors []*model.Author

	query := `
	SELECT id, first_name, last_name, email,  bio, created_at, deleted_at
	
	FROM authors
	WHERE id = ANY($1::bigint[])
	AND deleted_at IS NULL
	`

	rows, err := tx.Query(ctx, query, val.AuthorIds)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		author := &model.Author{}
		err := rows.Scan(
			&author.ID,
			&author.FirstName,
			&author.LastName,
			&author.Email,
			&author.Bio,
			&author.CreatedAt,
			&author.DeletedAt,
		)

		if err != nil {
			return nil, err
		}

		authors = append(authors, author)
	}

	rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(authors) != len(val.AuthorIds) {
		return nil, apperror.ErrAuthorNotFound
	}

	query = `
	INSERT INTO books (title, isbn, published_date, genre, price)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, title, isbn, published_date, genre, price, created_at, deleted_at
	`

	err = tx.QueryRow(ctx, query, val.Title, val.ISBN, val.PublishedDate, val.Genre, val.Price).Scan(
		&book.ID,
		&book.Title,
		&book.ISBN,
		&book.PublishedDate,
		&book.Genre,
		&book.Price,
		&book.CreatedAt,
		&book.DeletedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "books_isbn_key" {
				return nil, apperror.ErrBookISBNExists
			}
		}

		return nil, err
	}

	_, err = tx.Exec(ctx, `
    INSERT INTO book_authors (book_id, author_id)
    SELECT $1, unnest($2::bigint[])
`, book.ID, val.AuthorIds)

	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return nil, err
	}

	book.Authors = authors

	return book, nil
}

func (r *Repository) GetBooks(ctx context.Context, authorID int64, available bool, page, pageSize int, search string) ([]*model.Book, error) {
	if pageSize <= 0 {
		pageSize = utils.DefaultLimit
	} else if pageSize > 100 {
		pageSize = utils.DefaultLimit
	}

	if page <= 0 {
		page = utils.DefaultPage
	}
	offset := (page - 1) * pageSize

	var args []any
	var query strings.Builder

	query.WriteString(`
	SELECT 
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
	FROM books AS b
	LEFT JOIN book_authors AS ba ON ba.book_id = b.id
	LEFT JOIN authors AS a ON ba.author_id = a.id AND a.deleted_at IS NULL
	WHERE b.deleted_at IS NULL
	`)

	if authorID > 0 {
		args = append(args, authorID)
		query.WriteString(fmt.Sprintf(`
		AND EXISTS (
			SELECT 1 FROM book_authors sub_ba
			WHERE sub_ba.book_id = b.id AND sub_ba.author_id = $%d
		)`, len(args)))
	}

	if available {
		query.WriteString(`
		AND NOT EXISTS (
			SELECT 1 FROM loans l 
			WHERE l.book_id = b.id AND l.return_date IS NULL
		)`)
	}

	trimmedSearch := strings.TrimSpace(search)
	if trimmedSearch != "" {
		args = append(args, "%"+trimmedSearch+"%")
		searchPlaceholder := fmt.Sprintf("$%d", len(args))
		query.WriteString(fmt.Sprintf(`
		AND (
			b.title ILIKE %s
			OR b.isbn ILIKE %s
			OR b.genre ILIKE %s
			OR EXISTS (
				SELECT 1 
				FROM book_authors sub_ba
				JOIN authors sub_a ON sub_ba.author_id = sub_a.id
				WHERE sub_ba.book_id = b.id
				  AND sub_a.deleted_at IS NULL
				  AND (
					sub_a.first_name ILIKE %s
					OR sub_a.last_name ILIKE %s
				  )
			)
		)`, searchPlaceholder, searchPlaceholder, searchPlaceholder, searchPlaceholder, searchPlaceholder))
	}

	query.WriteString(`
	GROUP BY b.id
	ORDER BY b.id ASC
	`)

	args = append(args, pageSize)
	query.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))

	args = append(args, offset)
	query.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))

	rows, err := r.db.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]*model.Book, 0)
	for rows.Next() {
		book := &model.Book{}
		var authorsJSON []byte

		err := rows.Scan(
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

		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(books) == 0 {
		return []*model.Book{}, nil
	}

	return books, nil
}

func (r *Repository) GetBookByID(ctx context.Context, id int64) (*model.Book, error) {
	book := &model.Book{}
	var authorsJSON []byte

	query := `
	SELECT 
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
	FROM books AS b
	LEFT JOIN book_authors AS ba ON ba.book_id = b.id
	LEFT JOIN authors AS a ON ba.author_id = a.id AND a.deleted_at IS NULL
	WHERE b.deleted_at IS NULL
	AND b.id = $1
	GROUP BY b.id
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrBookNotFound
		}
		return nil, err
	}

	book.Authors = make([]*model.Author, 0)
	if len(authorsJSON) > 0 {
		if err := json.Unmarshal(authorsJSON, &book.Authors); err != nil {
			return nil, err
		}
	}

	return book, nil
}

func (r *Repository) GetBookBorrowers(ctx context.Context, id int64, includeReturned bool) ([]*model.LoanMember, error) {
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
		m.id, 
		m.first_name, 
		m.last_name, 
		m.email, 
		m.phone, 
		m.created_at, 
		m.removed_at
	FROM loans AS l
	JOIN members AS m ON l.member_id = m.id
	WHERE l.book_id = $1 
	  AND m.removed_at IS NULL

	`)

	if !includeReturned {
		query.WriteString("  AND l.return_date IS NULL")
	}

	query.WriteString(` ORDER BY l.loan_date DESC`)

	rows, err := r.db.Query(ctx, query.String(), id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	borrowers := make([]*model.LoanMember, 0)
	for rows.Next() {
		lm := &model.LoanMember{}
		err := rows.Scan(
			&lm.Loan.ID,
			&lm.Loan.BookID,
			&lm.Loan.MemberID,
			&lm.Loan.LoanDate,
			&lm.Loan.DueDate,
			&lm.Loan.ReturnDate,
			&lm.Loan.CreatedAt,
			&lm.Member.ID,
			&lm.Member.FirstName,
			&lm.Member.LastName,
			&lm.Member.Email,
			&lm.Member.Phone,
			&lm.Member.CreatedAt,
			&lm.Member.RemovedAt,
		)
		if err != nil {
			return nil, err
		}

		borrowers = append(borrowers, lm)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return borrowers, nil
}

