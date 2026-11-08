CREATE TABLE IF NOT EXISTS loans (
    id SERIAL PRIMARY KEY,
    book_id INT NOT NULL,
    member_id INT NOT NULL,
    loan_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    due_date TIMESTAMPTZ NOT NULL,
    return_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

    FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE RESTRICT,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE RESTRICT,

    CONSTRAINT chk_loans_due_date CHECK (due_date > loan_date),
    CONSTRAINT chk_loans_return_date CHECK (return_date IS NULL OR return_date >= loan_date)
);

CREATE INDEX idx_loans_member ON loans(member_id);
CREATE INDEX idx_overdue_loans ON loans(due_date) WHERE return_date IS NULL;
CREATE INDEX idx_open_loans_by_book ON loans(book_id) WHERE return_date IS NULL;
CREATE INDEX idx_open_loans_by_member ON loans(member_id) WHERE return_date IS NULL;
CREATE UNIQUE INDEX idx_one_active_per_book ON loans(book_id) WHERE return_date IS NULL;