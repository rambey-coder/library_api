# Auth Milestone Notes (`NOTES.md`)

## 1. Token Lifetime Rationale
- **Lifetime**: `24h` (`time.Now().Add(24 * time.Hour)`), signed with `HS256` (`jwt.SigningMethodHS256`) using `JWT_SECRET` from the environment.
- **Rationale**: Since the scope forbids session stores and refresh token tables, a 24-hour stateless access token balances usability (a user or staff member does not get logged out mid-shift) against revocation risk. Claims embed `UserID`, `MemberID` (nullable `*int64`), `Role`, `iat`, and `exp`.

## 2. Schema Linking (`users` ↔ `members`)
- Instead of duplicating loan foreign keys or merging `members` into `users`, `000006_users.up.sql` creates the `users` table and adds a nullable `user_id INT UNIQUE REFERENCES users(id)` column to `members`.
- Self-registration via `POST /auth/register` hardcodes `role = 'member'` and rejects any request body that attempts to supply a `role` field (`400 Bad Request`). Inside `Repository.RegisterUser`, the `users` row and linked `members` row are inserted atomically within a single PostgreSQL transaction.
- `GetUserByEmail` and `GetUserByID` perform a `LEFT JOIN members m ON u.id = m.user_id` so `MemberID` is populated on login and included in the JWT claims.

## 3. Out-of-Band Admin Seeding & Staff Role Decomposition (`librarian`)
- **CLI Generator (`cmd/hashpw/main.go`)**: Admin and librarian accounts cannot be self-registered via the API. They are provisioned out-of-band by generating a `bcrypt.DefaultCost` hash with `go run ./cmd/hashpw <password>` and inserting directly into `users` via SQL.
- **Addition of `librarian` Role**: Added `librarian` alongside `admin` and `member` in the schema (`users_role_check`) and route protection matrix as a staff role decomposition so operational library staff can manage catalog items and loans without holding full database/system admin identity.

## 4. Known Gap / Deferred Item: `POST /members` Onboarding Path
- Members created by staff via `POST /members` only have a `members` row (`user_id IS NULL`) and no `users` credentials row, whereas members created via `POST /auth/register` have both `users` and `members` rows in a single transaction.
- Deferred for a future iteration: either remove `POST /members` in favor of self-registration (`/auth/register`), or update `POST /members` to transactionally create a `users` row with a temporary generated password.
