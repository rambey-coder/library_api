package model

import "time"

type Loan struct {
	ID         int64      `json:"id"`
	BookID     int64      `json:"book_id"`
	MemberID   int64      `json:"member_id"`
	LoanDate   time.Time  `json:"loan_date"`
	DueDate    time.Time  `json:"due_date"`
	ReturnDate *time.Time `json:"return_date"`
	CreatedAt  time.Time  `json:"created_at"`
	Book       *Book      `json:"book,omitempty"`
	Member     *Member    `json:"member,omitempty"`
}

type LoanMember struct {
	Loan   Loan   `json:"loan"`
	Member Member `json:"member"`
}


type BookBorrowers struct {
	Book      *Book         `json:"book"`
	Borrowers []*LoanMember `json:"borrowers"`
}

type LoanReq struct {
	BookID   int64 `json:"book_id"`
	MemberID int64 `json:"-"`
	LoanDate time.Time `json:"loan_date"`
	DueDate  time.Time `json:"due_date"`
}