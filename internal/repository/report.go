package repository

import (
	"context"
	"encoding/json"
	"library_api/internal/model"
)

func (r *Repository) GetTopAuthors(ctx context.Context, limit int) ([]*model.AuthorBookCount, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
	SELECT a.id, a.first_name, a.last_name, COUNT(ba.book_id) AS book_count
	FROM authors AS a 
	JOIN book_authors AS ba 
	ON ba.author_id = a.id
	JOIN books AS b 
	ON b.id = ba.book_id AND b.deleted_at IS NULL
	WHERE a.deleted_at IS NULL
	GROUP BY a.id

	ORDER BY book_count DESC, a.id ASC
	LIMIT $1;
`
	top_authors := []*model.AuthorBookCount{}

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		author := &model.AuthorBookCount{}

		err := rows.Scan(
			&author.ID,
			&author.FirstName,
			&author.LastName,
			&author.BookCount,
		)

		if err != nil {
			return nil, err
		}

		top_authors = append(top_authors, author)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(top_authors) == 0 {
		return []*model.AuthorBookCount{}, nil
	}

	return top_authors, nil

}

func (r *Repository) GetAuthorsWithMoreThanNBooks(ctx context.Context, n int) ([]*model.AuthorBookCount, error) {
	query := `
	SELECT a.id, a.first_name, a.last_name, COUNT(ba.book_id) AS book_count
FROM authors a
JOIN book_authors ba ON ba.author_id = a.id
JOIN books b ON b.id = ba.book_id AND b.deleted_at IS NULL
WHERE a.deleted_at IS NULL
GROUP BY a.id
HAVING COUNT(ba.book_id) > $1
ORDER BY book_count DESC, a.id ASC
	`
	top_authors := []*model.AuthorBookCount{}

	rows, err := r.db.Query(ctx, query, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		author := &model.AuthorBookCount{}

		err := rows.Scan(
			&author.ID,
			&author.FirstName,
			&author.LastName,
			&author.BookCount,
		)

		if err != nil {
			return nil, err
		}

		top_authors = append(top_authors, author)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(top_authors) == 0 {
		return []*model.AuthorBookCount{}, nil
	}

	return top_authors, nil
}

// "never borrowed" means zero loan rows ever, not "currently on loan".
// Compare GetBooks(available=true), which filters by return_date IS NULL.
func (r *Repository) GetBooksNeverBorrowed(ctx context.Context) ([]*model.Book, error) {
	books := []*model.Book{}
	var authorsJSON []byte

	query := `
	SELECT b.id, 
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
	
 	AND NOT EXISTS (SELECT 1 FROM loans l WHERE l.book_id = b.id)
	GROUP BY b.id
	ORDER BY b.id ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err

	}
	defer rows.Close()

	for rows.Next() {
		book := &model.Book{}

		err = rows.Scan(
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

	return books, nil
}
