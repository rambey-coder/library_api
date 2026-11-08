package utils

const (
	DefaultPage            int = 1
	DefaultLimit           int = 20
	MaxLimit               int = 100
	DefaultTopAuthorsLimit int = 10
	MaxActiveLoans         int = 3
)

type Role string

const (
	RoleAdmin     = "admin"
	RoleMember    = "member"
	RoleLibrarian = "librarian"
)

type Gender string

const (
	Male   = "male"
	Female = "female"
)

type ctxKey string

const RequestIDKey ctxKey = "request_id"

const UserClaimsKey ctxKey = "user_claims"
