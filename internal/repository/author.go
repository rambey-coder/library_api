package repository

import (
	"context"
	"errors"
	"library_api/internal/apperror"
	"library_api/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) CreateAuthor(ctx context.Context, val model.CreateAuthorRequest) (*model.Author, error) {
	author := &model.Author{}

	query := `
	INSERT INTO authors (first_name, last_name, email, bio)
	VALUES ($1, $2, $3, $4)
	RETURNING id, first_name, last_name, email, bio, created_at, deleted_at;
	`

	err := r.db.QueryRow(ctx, query, val.FirstName, val.LastName, val.Email, val.Bio).Scan(
		&author.ID,
		&author.FirstName,
		&author.LastName,
		&author.Email,
		&author.Bio,
		&author.CreatedAt,
		&author.DeletedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {

			if pgErr.Code == "23505" && pgErr.ConstraintName == "authors_email_key" {
				return &model.Author{}, apperror.ErrAuthorEmailExists
			}
		}

		return &model.Author{}, err
	}

	return author, nil
}

func (r *Repository) GetAuthors(ctx context.Context) ([]*model.Author, error) {
	authors := []*model.Author{}

	query := `
	SELECT id, first_name, last_name, email, bio, created_at, deleted_at
	FROM authors
	WHERE deleted_at IS NULL;
	`

	rows, err := r.db.Query(ctx, query)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(authors) == 0 {
		return []*model.Author{}, nil
	}

	return authors, nil
}

func (r *Repository) GetAuthorByID(ctx context.Context, id int64) (*model.Author, error) {
	author := &model.Author{}

	query := `
	SELECT id, first_name, last_name, email, bio, created_at, deleted_at
	FROM authors
	WHERE id = $1 AND deleted_at IS NULL;
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&author.ID,
		&author.FirstName,
		&author.LastName,
		&author.Email,
		&author.Bio,
		&author.CreatedAt,
		&author.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrAuthorNotFound
		}

		return nil, err
	}

	return author, nil
}

func (r *Repository) GetBooksByAuthorID(ctx context.Context, id int64) ([]*model.Book, error) {
	books := []*model.Book{}

	query := `
	SELECT b.id, b.title, b.isbn, b.published_date, b.genre, b.price, b.created_at, b.deleted_at
	FROM books AS b
	 JOIN book_authors AS ba 
		ON ba.book_id = b.id
	WHERE ba.author_id = $1
		AND b.deleted_at IS NULL;
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		book := &model.Book{}

		err := rows.Scan(
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
			return nil, err
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

