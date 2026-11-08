Does the relation produce multiple child rows for 1 parent row?
│
├── NO (1:1 or N:1) ──> Plain JOIN & Plain Scan
│                       Example: Loan -> Book, Member -> Address
│                       Select columns directly: SELECT l.id, b.id, b.title ...
│                       Scan directly: rows.Scan(&loan.ID, &loan.Book.ID, &loan.Book.Title)
│
└── YES (1:N or M:N) ─> COALESCE(json_agg(json_build_object(...) FILTER ...) ... )
                        Example: Book -> []*Author, Member -> []*Loan
                        Why? Standard SQL JOIN produces duplicate parent rows:
                        Row 1: [Book 1, Author A]
                        Row 2: [Book 1, Author B]  <-- Book 1 is duplicated!
                        json_agg collapses them into 1 SQL row per parent.


// SCENARIO A: Target field is a SLICE -> USE json_agg
type Book struct {
    ID      int64
    Authors []*Author // <-- SLICE: 1 Book has MANY Authors
}

// SCENARIO B: Target field is a SINGLE STRUCT -> DO NOT USE json_agg
type Loan struct {
    ID   int64
    Book *Book // <-- SINGLE: 1 Loan has EXACTLY 1 Book
}

Anatomy of the expresssion 
COALESCE(                                  -- 4. Guarantees '[]' instead of SQL NULL if 0 children
    json_agg(                              -- 2. Combines objects into a JSON array: [{...}, {...}]
        json_build_object(                 -- 1. Maps columns to JSON keys: {"id": a.id, "name": a.first_name}
            'id', a.id,
            'first_name', a.first_name
        ) ORDER BY a.id ASC
    ) FILTER (WHERE a.id IS NOT NULL),     -- 3. Prevents [null] array when LEFT JOIN finds no match
    '[]'::json
) AS authors
