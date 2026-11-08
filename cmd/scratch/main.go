package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"library_api/internal/apperror"
	"library_api/internal/db"
	"library_api/internal/model"
	"library_api/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env not loaded, using system environment")
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer pool.Close()

	repo := repository.New(pool)

	fmt.Println("=================================================================")
	fmt.Println("       LIBRARY REPOSITORY MASTER COMPREHENSIVE TEST SUITE        ")
	fmt.Println("=================================================================")

	// 0. RESET FIXTURE
	_, err = pool.Exec(ctx, `
		TRUNCATE TABLE loans, book_authors, books, members, authors RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		log.Fatal("truncate failed:", err)
	}

	// =================================================================
	// PART 1: AUTHORS
	// =================================================================
	fmt.Println("\n--- [PART 1: AUTHORS] ---")

	// 1.1 CreateAuthor OK
	ursula, err := repo.CreateAuthor(ctx, model.CreateAuthorRequest{
		FirstName: "Ursula",
		LastName:  "Le Guin",
		Email:     "ursula@example.com",
		Bio:       "Speculative fiction author",
	})
	if err != nil {
		log.Fatalf("failed to create author Ursula: %v", err)
	}
	fmt.Printf("[1.1] CreateAuthor OK -> ID=%d, Name=%s %s\n", ursula.ID, ursula.FirstName, ursula.LastName)

	octavia, err := repo.CreateAuthor(ctx, model.CreateAuthorRequest{
		FirstName: "Octavia",
		LastName:  "Butler",
		Email:     "octavia@example.com",
		Bio:       "Science fiction author",
	})
	if err != nil {
		log.Fatalf("failed to create author Octavia: %v", err)
	}
	fmt.Printf("[1.1] CreateAuthor OK -> ID=%d, Name=%s %s\n", octavia.ID, octavia.FirstName, octavia.LastName)

	// 1.2 CreateAuthor Duplicate Email
	_, err = repo.CreateAuthor(ctx, model.CreateAuthorRequest{
		FirstName: "Imposter",
		LastName:  "Ursula",
		Email:     "ursula@example.com",
	})
	if errors.Is(err, apperror.ErrAuthorEmailExists) {
		fmt.Printf("[1.2] CreateAuthor Duplicate -> Expected error: %v\n", err)
	} else {
		log.Fatalf("expected ErrAuthorEmailExists, got: %v", err)
	}

	// 1.3 GetAuthors
	authors, err := repo.GetAuthors(ctx)
	if err != nil || len(authors) != 2 {
		log.Fatalf("GetAuthors failed or count mismatch: len=%d err=%v", len(authors), err)
	}
	fmt.Printf("[1.3] GetAuthors -> Returned count=%d authors\n", len(authors))

	// 1.4 GetAuthorByID OK & NotFound
	gotAuthor, err := repo.GetAuthorByID(ctx, ursula.ID)
	if err != nil || gotAuthor.Email != "ursula@example.com" {
		log.Fatalf("GetAuthorByID failed: %v", err)
	}
	fmt.Printf("[1.4] GetAuthorByID OK -> ID=%d, Email=%s\n", gotAuthor.ID, gotAuthor.Email)

	_, err = repo.GetAuthorByID(ctx, 999999)
	if errors.Is(err, apperror.ErrAuthorNotFound) {
		fmt.Printf("[1.4] GetAuthorByID Not Found -> Expected error: %v\n", err)
	} else {
		log.Fatalf("expected ErrAuthorNotFound, got: %v", err)
	}

	ursulaID := int64(ursula.ID)
	octaviaID := int64(octavia.ID)

	// =================================================================
	// PART 2: BOOKS
	// =================================================================
	fmt.Println("\n--- [PART 2: BOOKS] ---")

	// 2.1 CreateBook Valid & Hydrate Authors in GetBookByID
	b1, err := repo.CreateBook(ctx, model.CreateBookRequest{
		Title:         "A Wizard of Earthsea",
		ISBN:          "9780547773742",
		PublishedDate: time.Date(1968, 11, 1, 0, 0, 0, 0, time.UTC),
		Genre:         "fantasy",
		Price:         1299,
		AuthorIds:     []int64{ursulaID},
	})
	if err != nil {
		log.Fatalf("CreateBook 1 failed: %v", err)
	}
	fmt.Printf("[2.1] CreateBook OK -> ID=%d, Title=%q\n", b1.ID, b1.Title)

	gotB1, err := repo.GetBookByID(ctx, b1.ID)
	if err != nil || len(gotB1.Authors) != 1 {
		log.Fatalf("GetBookByID failed: len(Authors)=%d err=%v", len(gotB1.Authors), err)
	}
	fmt.Printf("[2.1] GetBookByID OK -> Title=%q, AuthorsCount=%d (%s)\n", gotB1.Title, len(gotB1.Authors), gotB1.Authors[0].LastName)

	// 2.2 CreateBook with Bad Author ID
	_, err = repo.CreateBook(ctx, model.CreateBookRequest{
		Title:         "Ghost Book",
		ISBN:          "9780000000001",
		PublishedDate: time.Now(),
		Genre:         "sci-fi",
		Price:         999,
		AuthorIds:     []int64{999999},
	})
	if errors.Is(err, apperror.ErrAuthorNotFound) {
		fmt.Printf("[2.2] CreateBook Bad Author -> Expected error: %v\n", err)
	} else {
		log.Fatalf("expected ErrAuthorNotFound, got: %v", err)
	}

	// 2.3 CreateBook with Duplicate ISBN
	_, err = repo.CreateBook(ctx, model.CreateBookRequest{
		Title:         "Duplicate ISBN Book",
		ISBN:          "9780547773742",
		PublishedDate: time.Now(),
		Genre:         "fantasy",
		Price:         1500,
		AuthorIds:     []int64{ursulaID},
	})
	if errors.Is(err, apperror.ErrBookISBNExists) {
		fmt.Printf("[2.3] CreateBook Duplicate ISBN -> Expected error: %v\n", err)
	} else {
		log.Fatalf("expected ErrBookISBNExists, got: %v", err)
	}

	// 2.4 GetBookByID Not Found
	_, err = repo.GetBookByID(ctx, 999999)
	if errors.Is(err, apperror.ErrBookNotFound) {
		fmt.Printf("[2.4] GetBookByID Not Found -> Expected error: %v\n", err)
	} else {
		log.Fatalf("expected ErrBookNotFound, got: %v", err)
	}

	// Seed Book 2 (Ursula) and Book 3 (Octavia)
	b2, err := repo.CreateBook(ctx, model.CreateBookRequest{
		Title:         "The Left Hand of Darkness",
		ISBN:          "9780441478125",
		PublishedDate: time.Date(1969, 3, 1, 0, 0, 0, 0, time.UTC),
		Genre:         "science fiction",
		Price:         1399,
		AuthorIds:     []int64{ursulaID},
	})
	if err != nil {
		log.Fatalf("CreateBook 2 failed: %v", err)
	}

	b3, err := repo.CreateBook(ctx, model.CreateBookRequest{
		Title:         "Kindred",
		ISBN:          "9780807083697",
		PublishedDate: time.Date(1979, 6, 1, 0, 0, 0, 0, time.UTC),
		Genre:         "science fiction",
		Price:         1599,
		AuthorIds:     []int64{octaviaID},
	})
	if err != nil {
		log.Fatalf("CreateBook 3 failed: %v", err)
	}
	fmt.Printf("[2.1] CreateBook 3 OK -> ID=%d, Title=%q\n", b3.ID, b3.Title)

	// 2.5 GetBooksByAuthorID
	ursulaBooksLegacy, err := repo.GetBooksByAuthorID(ctx, ursulaID)
	if err != nil || len(ursulaBooksLegacy) != 2 {
		log.Fatalf("GetBooksByAuthorID failed: len=%d err=%v", len(ursulaBooksLegacy), err)
	}
	fmt.Printf("[2.5] GetBooksByAuthorID(Ursula) -> Count=%d books\n", len(ursulaBooksLegacy))

	// =================================================================
	// PART 3: MEMBERS
	// =================================================================
	fmt.Println("\n--- [PART 3: MEMBERS] ---")

	// 3.1 CreateMember OK
	phone := "1234567890"
	member1, err := repo.CreateMember(ctx, model.CreateMemberRequest{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@example.com",
		Phone:     &phone,
	})
	if err != nil {
		log.Fatalf("CreateMember 1 failed: %v", err)
	}
	fmt.Printf("[3.1] CreateMember OK -> ID=%d, Name=%s %s, Email=%s\n", member1.ID, member1.FirstName, member1.LastName, member1.Email)

	member2, err := repo.CreateMember(ctx, model.CreateMemberRequest{
		FirstName: "Charles",
		LastName:  "Babbage",
		Email:     "charles@example.com",
	})
	if err != nil {
		log.Fatalf("CreateMember 2 failed: %v", err)
	}
	fmt.Printf("[3.1] CreateMember OK -> ID=%d, Name=%s %s\n", member2.ID, member2.FirstName, member2.LastName)

	// 3.2 CreateMember Duplicate Email
	_, err = repo.CreateMember(ctx, model.CreateMemberRequest{
		FirstName: "Imposter",
		LastName:  "Ada",
		Email:     "ada@example.com",
	})
	if errors.Is(err, apperror.ErrMemberEmailExists) {
		fmt.Printf("[3.2] CreateMember Duplicate Email -> Expected error: %v\n", err)
	} else {
		log.Fatalf("expected ErrMemberEmailExists, got: %v", err)
	}

	// =================================================================
	// PART 4: LOANS & BORROWERS
	// =================================================================
	fmt.Println("\n--- [PART 4: LOANS & BORROWERS] ---")

	// Loan 1: Ada borrows Book 2 (The Left Hand of Darkness) -> ACTIVE
	var loan1ID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO loans (book_id, member_id, loan_date, due_date, return_date)
		VALUES ($1, $2, CURRENT_TIMESTAMP - INTERVAL '3 days', CURRENT_TIMESTAMP + INTERVAL '11 days', NULL)
		RETURNING id;
	`, b2.ID, member1.ID).Scan(&loan1ID)
	if err != nil {
		log.Fatalf("seed loan 1 failed: %v", err)
	}
	fmt.Printf("[SETUP] Created Active Loan ID=%d: Member %s borrows %q\n", loan1ID, member1.FirstName, b2.Title)

	// Loan 2: Ada borrowed Book 1 (A Wizard of Earthsea) -> RETURNED
	var loan2ID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO loans (book_id, member_id, loan_date, due_date, return_date)
		VALUES ($1, $2, CURRENT_TIMESTAMP - INTERVAL '20 days', CURRENT_TIMESTAMP - INTERVAL '6 days', CURRENT_TIMESTAMP - INTERVAL '7 days')
		RETURNING id;
	`, b1.ID, member1.ID).Scan(&loan2ID)
	if err != nil {
		log.Fatalf("seed loan 2 failed: %v", err)
	}
	fmt.Printf("[SETUP] Created Returned Loan ID=%d: Member %s returned %q\n", loan2ID, member1.FirstName, b1.Title)

	// 4.1 GetMembers with OpenLoanCount
	membersList, err := repo.GetMembers(ctx, 1, 20)
	if err != nil {
		log.Fatalf("GetMembers failed: %v", err)
	}
	fmt.Printf("[4.1] GetMembers -> Total members: %d\n", len(membersList))
	for _, m := range membersList {
		fmt.Printf("    - ID=%d, %s %s: OpenLoanCount=%d\n", m.ID, m.FirstName, m.LastName, m.OpenLoanCount)
		if m.ID == member1.ID && m.OpenLoanCount != 1 {
			log.Fatalf("expected member 1 to have 1 open loan, got %d", m.OpenLoanCount)
		}
		if m.ID == member2.ID && m.OpenLoanCount != 0 {
			log.Fatalf("expected member 2 to have 0 open loans, got %d", m.OpenLoanCount)
		}
	}

	// 4.2 GetBookBorrowers for Book 2 (Active Loan)
	b2BorrowersActive, err := repo.GetBookBorrowers(ctx, b2.ID, false)
	if err != nil || len(b2BorrowersActive) != 1 {
		log.Fatalf("GetBookBorrowers(b2, false) failed: len=%d err=%v", len(b2BorrowersActive), err)
	}
	fmt.Printf("[4.2] GetBookBorrowers(Book 2, includeReturned=false) -> Count=%d (Borrower: %s %s)\n",
		len(b2BorrowersActive), b2BorrowersActive[0].Member.FirstName, b2BorrowersActive[0].Member.LastName)

	// 4.3 GetBookBorrowers for Book 1 (Returned Loan only)
	b1BorrowersActive, err := repo.GetBookBorrowers(ctx, b1.ID, false)
	if err != nil || len(b1BorrowersActive) != 0 {
		log.Fatalf("GetBookBorrowers(b1, false) failed: expected 0, got len=%d err=%v", len(b1BorrowersActive), err)
	}
	fmt.Printf("[4.3] GetBookBorrowers(Book 1, includeReturned=false) -> Count=%d (Excludes returned loans)\n", len(b1BorrowersActive))

	b1BorrowersHistory, err := repo.GetBookBorrowers(ctx, b1.ID, true)
	if err != nil || len(b1BorrowersHistory) != 1 {
		log.Fatalf("GetBookBorrowers(b1, true) failed: expected 1, got len=%d err=%v", len(b1BorrowersHistory), err)
	}
	fmt.Printf("[4.3] GetBookBorrowers(Book 1, includeReturned=true) -> Count=%d (History preserved: %s %s)\n",
		len(b1BorrowersHistory), b1BorrowersHistory[0].Member.FirstName, b1BorrowersHistory[0].Member.LastName)

	// 4.4 GetMemberLoans for Ada (Active vs All)
	adaActiveLoans, err := repo.GetMemberLoans(ctx, member1.ID, false)
	if err != nil || len(adaActiveLoans) != 1 {
		log.Fatalf("GetMemberLoans(ada, false) failed: len=%d err=%v", len(adaActiveLoans), err)
	}
	fmt.Printf("[4.4] GetMemberLoans(Ada, includeReturned=false) -> Count=%d active loan:\n", len(adaActiveLoans))
	for _, l := range adaActiveLoans {
		fmt.Printf("    - Loan ID=%d: Book=%q (ISBN: %s), Authors=%d\n",
			l.ID, l.Book.Title, l.Book.ISBN, len(l.Book.Authors))
		if l.Book.ID != b2.ID {
			log.Fatalf("expected loan for book %d, got %d", b2.ID, l.Book.ID)
		}
	}

	adaAllLoans, err := repo.GetMemberLoans(ctx, member1.ID, true)
	if err != nil || len(adaAllLoans) != 2 {
		log.Fatalf("GetMemberLoans(ada, true) failed: len=%d err=%v", len(adaAllLoans), err)
	}
	fmt.Printf("[4.4] GetMemberLoans(Ada, includeReturned=true) -> Count=%d total loans (active + returned):\n", len(adaAllLoans))
	for _, l := range adaAllLoans {
		status := "Active"
		if l.ReturnDate != nil {
			status = "Returned"
		}
		fmt.Printf("    - Loan ID=%d [%s]: Book=%q, Author=%s\n",
			l.ID, status, l.Book.Title, l.Book.Authors[0].LastName)
	}

	// 4.5 GetMemberLoans for Non-Existent Member
	_, err = repo.GetMemberLoans(ctx, 999999, false)
	if errors.Is(err, apperror.ErrMemberNotFound) {
		fmt.Printf("[4.5] GetMemberLoans(NonExistent) -> Expected error: %v\n", err)
	} else {
		log.Fatalf("expected ErrMemberNotFound, got: %v", err)
	}

	// =================================================================
	// PART 5: GetBooks DYNAMIC FILTERS
	// =================================================================
	fmt.Println("\n--- [PART 5: GetBooks DYNAMIC FILTERS] ---")

	// 5.1 No Filters (All Books)
	allBooks, err := repo.GetBooks(ctx, 0, false, 1, 20, "")
	if err != nil || len(allBooks) != 3 {
		log.Fatalf("GetBooks (no filters) failed: len=%d err=%v", len(allBooks), err)
	}
	fmt.Printf("[5.1] GetBooks (no filters) -> Count=%d books (All hydrated with authors)\n", len(allBooks))

	// 5.2 Filter by Author ID
	ursulaBooks, err := repo.GetBooks(ctx, ursulaID, false, 1, 20, "")
	if err != nil || len(ursulaBooks) != 2 {
		log.Fatalf("GetBooks(ursulaID) failed: len=%d err=%v", len(ursulaBooks), err)
	}
	fmt.Printf("[5.2] GetBooks?author_id=Ursula -> Count=%d (Expected 2)\n", len(ursulaBooks))

	octaviaBooks, err := repo.GetBooks(ctx, octaviaID, false, 1, 20, "")
	if err != nil || len(octaviaBooks) != 1 {
		log.Fatalf("GetBooks(octaviaID) failed: len=%d err=%v", len(octaviaBooks), err)
	}
	fmt.Printf("[5.2] GetBooks?author_id=Octavia -> Count=%d (Expected 1: %q)\n", len(octaviaBooks), octaviaBooks[0].Title)

	// 5.3 Filter by Search / Title
	searchBooks, err := repo.GetBooks(ctx, 0, false, 1, 20, "wizard")
	if err != nil || len(searchBooks) != 1 {
		log.Fatalf("GetBooks(search='wizard') failed: len=%d err=%v", len(searchBooks), err)
	}
	fmt.Printf("[5.3] GetBooks?title=wizard -> Count=%d (Found: %q)\n", len(searchBooks), searchBooks[0].Title)

	// 5.4 Filter by Availability (Excludes Book 2 which is actively borrowed)
	availableBooks, err := repo.GetBooks(ctx, 0, true, 1, 20, "")
	if err != nil || len(availableBooks) != 2 {
		log.Fatalf("GetBooks(available=true) failed: expected 2, got len=%d err=%v", len(availableBooks), err)
	}
	fmt.Printf("[5.4] GetBooks?available=true -> Count=%d (Excluded active borrowed Book 2: %q)\n",
		len(availableBooks), b2.Title)
	for _, b := range availableBooks {
		fmt.Printf("    - ID=%d: %q (Available)\n", b.ID, b.Title)
	}

	// 5.5 Pagination Bounding
	pagedBooks, err := repo.GetBooks(ctx, 0, false, 1, 1, "")
	if err != nil || len(pagedBooks) != 1 {
		log.Fatalf("GetBooks(pageSize=1) failed: len=%d err=%v", len(pagedBooks), err)
	}
	fmt.Printf("[5.5] GetBooks(page=1, pageSize=1) -> Count=%d (First item: %q)\n", len(pagedBooks), pagedBooks[0].Title)

	pagedPage2, err := repo.GetBooks(ctx, 0, false, 2, 1, "")
	if err != nil || len(pagedPage2) != 1 {
		log.Fatalf("GetBooks(page=2, pageSize=1) failed: len=%d err=%v", len(pagedPage2), err)
	}
	fmt.Printf("[5.5] GetBooks(page=2, pageSize=1) -> Count=%d (Second item: %q)\n", len(pagedPage2), pagedPage2[0].Title)

	// =================================================================
	// PART 6: BORROWING & RETURNING
	// =================================================================
	fmt.Println("\n--- [PART 6: BORROWING & RETURNING] ---")

	// 6.1 BorrowBook happy path
	loan3, err := repo.BorrowBook(ctx, model.LoanReq{
		BookID:   b3.ID,
		MemberID: member2.ID,
		LoanDate: time.Now(),
		DueDate:  time.Now().AddDate(0, 0, 14),
	})
	if err != nil {
		log.Fatalf("[6.1] BorrowBook failed: %v", err)
	}
	if loan3.ReturnDate != nil {
		log.Fatalf("[6.1] expected nil return_date, got %v", loan3.ReturnDate)
	}
	fmt.Printf("[6.1] BorrowBook OK -> Loan ID=%d, Book=%q, Member=%s %s\n",
		loan3.ID, b3.Title, member2.FirstName, member2.LastName)

	// 6.2 BorrowBook on nonexistent book
	_, err = repo.BorrowBook(ctx, model.LoanReq{
		BookID: 999999, MemberID: member2.ID,
		LoanDate: time.Now(), DueDate: time.Now().AddDate(0, 0, 14),
	})
	if errors.Is(err, apperror.ErrBookNotFound) {
		fmt.Printf("[6.2] BorrowBook(nonexistent book) -> Expected: %v\n", err)
	} else {
		log.Fatalf("[6.2] expected ErrBookNotFound, got: %v", err)
	}

	// 6.3 BorrowBook on nonexistent member
	_, err = repo.BorrowBook(ctx, model.LoanReq{
		BookID: b1.ID, MemberID: 999999,
		LoanDate: time.Now(), DueDate: time.Now().AddDate(0, 0, 14),
	})
	if errors.Is(err, apperror.ErrMemberNotFound) {
		fmt.Printf("[6.3] BorrowBook(nonexistent member) -> Expected: %v\n", err)
	} else {
		log.Fatalf("[6.3] expected ErrMemberNotFound, got: %v", err)
	}

	// 6.4 BorrowBook on already-borrowed book (book 2 is active on loan)
	_, err = repo.BorrowBook(ctx, model.LoanReq{
		BookID: b2.ID, MemberID: member2.ID,
		LoanDate: time.Now(), DueDate: time.Now().AddDate(0, 0, 14),
	})
	if errors.Is(err, apperror.ErrBookAlreadyOnLoan) {
		fmt.Printf("[6.4] BorrowBook(already on loan) -> Expected: %v\n", err)
	} else {
		log.Fatalf("[6.4] expected ErrBookAlreadyOnLoan, got: %v", err)
	}

	// 6.5 ReturnBook happy path (return loan3)
	returned, err := repo.ReturnBook(ctx, loan3.ID)
	if err != nil {
		log.Fatalf("[6.5] ReturnBook failed: %v", err)
	}
	if returned.ReturnDate == nil {
		log.Fatalf("[6.5] expected non-nil return_date after return")
	}
	fmt.Printf("[6.5] ReturnBook OK -> Loan ID=%d, ReturnDate=%s\n",
		returned.ID, returned.ReturnDate.Format(time.RFC3339))

	// 6.6 ReturnBook already returned
	_, err = repo.ReturnBook(ctx, loan3.ID)
	if errors.Is(err, apperror.ErrLoanAlreadyReturned) {
		fmt.Printf("[6.6] ReturnBook(already returned) -> Expected: %v\n", err)
	} else {
		log.Fatalf("[6.6] expected ErrLoanAlreadyReturned, got: %v", err)
	}

	// 6.7 ReturnBook nonexistent
	_, err = repo.ReturnBook(ctx, 999999)
	if errors.Is(err, apperror.ErrLoanNotFound) {
		fmt.Printf("[6.7] ReturnBook(nonexistent) -> Expected: %v\n", err)
	} else {
		log.Fatalf("[6.7] expected ErrLoanNotFound, got: %v", err)
	}

	// 6.8 Seed an overdue open loan for GetLoans filter
	var overdueLoanID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO loans (book_id, member_id, loan_date, due_date, return_date)
		VALUES ($1, $2, CURRENT_TIMESTAMP - INTERVAL '20 days', CURRENT_TIMESTAMP - INTERVAL '6 days', NULL)
		RETURNING id;
	`, b1.ID, member2.ID).Scan(&overdueLoanID)
	if err != nil {
		log.Fatalf("[6.8] seed overdue loan failed: %v", err)
	}
	fmt.Printf("[SETUP] Seeded overdue open loan ID=%d on Book %q\n", overdueLoanID, b1.Title)

	// 6.9 GetLoans (open only)
	openLoans, err := repo.GetLoans(ctx, 1, 20, false)
	if err != nil {
		log.Fatalf("[6.9] GetLoans failed: %v", err)
	}
	fmt.Printf("[6.9] GetLoans(open) -> Count=%d\n", len(openLoans))

	// 6.10 GetLoans?overdue=true (must include only the overdue open loan)
	overdueLoans, err := repo.GetLoans(ctx, 1, 20, true)
	if err != nil {
		log.Fatalf("[6.10] GetLoans(overdue) failed: %v", err)
	}
	fmt.Printf("[6.10] GetLoans(overdue) -> Count=%d\n", len(overdueLoans))
	for _, l := range overdueLoans {
		if l.ID == overdueLoanID {
			fmt.Printf("    - Overdue loan matched: ID=%d, Book=%q, Due=%s\n",
				l.ID, l.Book.Title, l.DueDate.Format("2006-01-02"))
		}
	}

	// 6.11 GetLoanByID OK + not found
	gotLoan, err := repo.GetLoan(ctx, overdueLoanID)
	if err != nil || gotLoan.ID != overdueLoanID {
		log.Fatalf("[6.11] GetLoan failed: %v", err)
	}
	fmt.Printf("[6.11] GetLoan OK -> ID=%d, Book=%q, Member=%s %s\n",
		gotLoan.ID, gotLoan.Book.Title, gotLoan.Member.FirstName, gotLoan.Member.LastName)

	_, err = repo.GetLoan(ctx, 999999)
	if errors.Is(err, apperror.ErrLoanNotFound) {
		fmt.Printf("[6.11] GetLoan(not found) -> Expected: %v\n", err)
	} else {
		log.Fatalf("[6.11] expected ErrLoanNotFound, got: %v", err)
	}

	// 6.12 CONCURRENCY: two simultaneous borrows of the same book
	// Book 1 is now on loan (overdueLoanID). Use a fresh book.
	concurrencyBook, err := repo.CreateBook(ctx, model.CreateBookRequest{
		Title:         "The Dispossessed",
		ISBN:          "9780060512750",
		PublishedDate: time.Date(1974, 5, 1, 0, 0, 0, 0, time.UTC),
		Genre:         "science fiction",
		Price:         1499,
		AuthorIds:     []int64{ursulaID},
	})
	if err != nil {
		log.Fatalf("[6.12] seed concurrency book failed: %v", err)
	}

	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.BorrowBook(ctx, model.LoanReq{
				BookID:   concurrencyBook.ID,
				MemberID: member1.ID,
				LoanDate: time.Now(),
				DueDate:  time.Now().AddDate(0, 0, 14),
			})
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	successes, conflicts, unexpected := 0, 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, apperror.ErrBookAlreadyOnLoan):
			conflicts++
		default:
			unexpected++
			log.Printf("[6.12] unexpected error: %v", err)
		}
	}
	fmt.Printf("[6.12] Concurrency: successes=%d, conflicts=%d, unexpected=%d\n",
		successes, conflicts, unexpected)
	if successes != 1 || conflicts != 1 {
		log.Fatalf("[6.12] FAIL: expected exactly 1 success + 1 conflict")
	}

	// =================================================================
	// PART 7: REPORTS
	// =================================================================
	fmt.Println("\n--- [PART 7: REPORTS] ---")

	// 7.1 GetTopAuthors
	topAuthors, err := repo.GetTopAuthors(ctx, 3)
	if err != nil {
		log.Fatalf("[7.1] GetTopAuthors failed: %v", err)
	}
	fmt.Printf("[7.1] GetTopAuthors(3) -> Count=%d\n", len(topAuthors))
	for i, a := range topAuthors {
		fmt.Printf("    #%d: %s %s (%d books)\n", i+1, a.FirstName, a.LastName, a.BookCount)
	}
	if len(topAuthors) < 2 || topAuthors[0].LastName != "Le Guin" {
		log.Fatalf("[7.1] FAIL: expected Le Guin #1, got %+v", topAuthors)
	}

	// 7.2 GetAuthorsWithMoreThanNBooks
	authorsMoreThan1, err := repo.GetAuthorsWithMoreThanNBooks(ctx, 1)
	if err != nil {
		log.Fatalf("[7.2] GetAuthorsWithMoreThanNBooks(1) failed: %v", err)
	}
	fmt.Printf("[7.2] GetAuthorsWithMoreThanNBooks(1) -> Count=%d (expect only Le Guin)\n", len(authorsMoreThan1))
	if len(authorsMoreThan1) != 1 || authorsMoreThan1[0].LastName != "Le Guin" {
		log.Fatalf("[7.2] FAIL: expected only Le Guin, got %+v", authorsMoreThan1)
	}

	authorsMoreThan0, err := repo.GetAuthorsWithMoreThanNBooks(ctx, 0)
	if err != nil || len(authorsMoreThan0) < 2 {
		log.Fatalf("[7.2] GetAuthorsWithMoreThanNBooks(0) failed: len=%d err=%v", len(authorsMoreThan0), err)
	}
	fmt.Printf("[7.2] GetAuthorsWithMoreThanNBooks(0) -> Count=%d (both authors)\n", len(authorsMoreThan0))

	// 7.3 GetBooksNeverBorrowed — seed a fresh book first
	freshBook, err := repo.CreateBook(ctx, model.CreateBookRequest{
		Title:         "The Farthest Shore",
		ISBN:          "9780689845346",
		PublishedDate: time.Date(1972, 9, 1, 0, 0, 0, 0, time.UTC),
		Genre:         "fantasy",
		Price:         1299,
		AuthorIds:     []int64{ursulaID},
	})
	if err != nil {
		log.Fatalf("[7.3] seed fresh book failed: %v", err)
	}

	neverBorrowed, err := repo.GetBooksNeverBorrowed(ctx)
	if err != nil {
		log.Fatalf("[7.3] GetBooksNeverBorrowed failed: %v", err)
	}
	fmt.Printf("[7.3] GetBooksNeverBorrowed -> Count=%d (expect 1: %q)\n", len(neverBorrowed), freshBook.Title)
	if len(neverBorrowed) != 1 || neverBorrowed[0].ID != freshBook.ID {
		log.Fatalf("[7.3] FAIL: expected only %q, got %+v", freshBook.Title, neverBorrowed)
	}

	// 7.4 Soft-delete filter: GetTopAuthors and GetAuthorsWithMoreThanNBooks
	// must agree once a book is soft-deleted.
	//
	// Before: Le Guin has 4 books (Earthsea, Left Hand, Dispossessed, Farthest Shore).
	// After soft-deleting Farthest Shore: 3.
	_, err = pool.Exec(ctx, `UPDATE books SET deleted_at = NOW() WHERE id = $1`, freshBook.ID)
	if err != nil {
		log.Fatalf("[7.4] soft-delete failed: %v", err)
	}

	topAfterDelete, err := repo.GetTopAuthors(ctx, 3)
	if err != nil {
		log.Fatalf("[7.4] GetTopAuthors failed: %v", err)
	}
	var leGuinCount int64
	for _, a := range topAfterDelete {
		if a.LastName == "Le Guin" {
			leGuinCount = int64(a.BookCount)
		}
	}
	fmt.Printf("[7.4] After soft-delete, Le Guin book_count = %d (expected 3)\n", leGuinCount)
	if leGuinCount != 3 {
		log.Fatalf("[7.4] FAIL: expected 3, got %d", leGuinCount)
	}

	// Cross-report check: both queries must agree on the count.
	moreThan3, err := repo.GetAuthorsWithMoreThanNBooks(ctx, 3)
	if err != nil {
		log.Fatalf("[7.4] GetAuthorsWithMoreThanNBooks(3) failed: %v", err)
	}
	fmt.Printf("[7.4] GetAuthorsWithMoreThanNBooks(3) -> Count=%d (expected 0)\n", len(moreThan3))
	if len(moreThan3) != 0 {
		log.Fatalf("[7.4] FAIL: expected 0 authors with >3 books, got %+v", moreThan3)
	}

	// Verify soft-deleted book is gone from GetBooks too.
	allAfter, err := repo.GetBooks(ctx, 0, false, 1, 100, "")
	if err != nil {
		log.Fatalf("[7.4] GetBooks failed: %v", err)
	}
	for _, b := range allAfter {
		if b.ID == freshBook.ID {
			log.Fatalf("[7.4] FAIL: soft-deleted book still returned by GetBooks")
		}
	}
	fmt.Printf("[7.4] GetBooks excludes soft-deleted book (total now %d)\n", len(allAfter))

	fmt.Println("\n=================================================================")
	fmt.Println("       MASTER SCRATCH SUITE: ALL CHECKS PASSED                  ")
	fmt.Println("=================================================================")
}
