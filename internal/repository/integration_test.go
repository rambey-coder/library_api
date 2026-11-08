//go:build integration

package repository

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestGetBooksIntegration(t *testing.T) {
	dns := os.Getenv("TEST_DB_URL")
	if dns == "" {
		t.Skip("TEST_DB_URL is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dns)
	if err != nil {
		t.Fatal("failed to connect to database:", err)
	}
	t.Cleanup(pool.Close)

	repo := New(pool)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	var authorID int64
	err = tx.QueryRow(ctx, `
	INSERT INTO authors (first_name, last_name, email)
	VALUES ('Chinua', 'Achebe', 'achebe_test@example.com')
	RETURNING id
`).Scan(&authorID)
	if err != nil {
		t.Fatalf("failed to seed author: %v", err)
	}

	var bookID int64
	err = tx.QueryRow(ctx, `
	INSERT INTO books (title, isbn, published_date, genre, price)
	VALUES ('Things Fall Apart', '9780385474542', NOW(), 'Fiction', 1500)
	RETURNING id
`).Scan(&bookID)
	if err != nil {
		t.Fatalf("failed to seed book: %v", err)
	}

	_, err = tx.Exec(ctx, `
	INSERT INTO book_authors (book_id, author_id)
	VALUES ($1, $2)
`, bookID, authorID)
	if err != nil {
		t.Fatalf("failed to seed book_authors: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("failed to commit seed tx: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM book_authors WHERE book_id = $1`, bookID)
		_, _ = pool.Exec(ctx, `DELETE FROM books WHERE id = $1`, bookID)
		_, _ = pool.Exec(ctx, `DELETE FROM authors WHERE id = $1`, authorID)
	})

	books, err := repo.GetBooks(ctx, authorID, false, 1, 10, "")
	if err != nil {
		t.Fatalf("GetBooks failed: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("want 1 book, got %d", len(books))
	}
	if len(books[0].Authors) != 1 {
		t.Fatalf("want 1 author on book, got %d", len(books[0].Authors))
	}
	if books[0].Authors[0].ID != authorID {
		t.Fatalf("want author ID %d, got %d", authorID, books[0].Authors[0].ID)
	}
	if books[0].Authors[0].FirstName != "Chinua" {
		t.Fatalf("want author FirstName %q, got %q", "Chinua", books[0].Authors[0].FirstName)
	}
}
