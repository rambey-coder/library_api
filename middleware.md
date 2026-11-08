Analyzed main.go:L1-81
Analyzed middleware.go:L1-72

Think of middleware as **layers of an onion** (or security checkpoints at an airport) wrapping your router (`mux`).

---

### 1. The Core Building Block: `http.Handler`

In Go's `net/http` package, anything that has a `ServeHTTP(w http.ResponseWriter, r *http.Request)` method is an `http.Handler`:
- Your router `mux` ([main.go](file:///Users/mac/Documents/backend/golang/library_api/main.go#L30)) is an `http.Handler`.
- `http.HandlerFunc(...)` is a type adapter that turns an ordinary function `func(w, r)` into an `http.Handler`.

A middleware is simply a function that **takes the next layer (`next http.Handler`) and returns a new layer (`http.Handler`)** that wraps it:

```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Code BEFORE next.ServeHTTP runs on the way IN
        next.ServeHTTP(w, r) // <-- hands the request to the next inner layer
        // 2. Code AFTER next.ServeHTTP runs on the way OUT
    })
}
```

---

### 2. How a Request Flows Through Your Chain in [main.go](file:///Users/mac/Documents/backend/golang/library_api/main.go#L68-L74)

In [main.go](file:///Users/mac/Documents/backend/golang/library_api/main.go#L68-L74), you stacked them like Russian nesting dolls:

```text
Recover (outermost)
  └── RequestID
        └── Logging
              └── Timeout
                    └── mux -> handler.CreateAuthor (innermost)
```

Here is the exact timeline when a client sends `POST /authors`:

1. **[`Recover`](file:///Users/mac/Documents/backend/golang/library_api/internal/middleware/middleware.go#L27-L37) (Way IN)**:
   - Sets up a `defer func() { if v := recover(); v != nil ... }()` safety net.
   - Calls `next.ServeHTTP(w, r)`, passing the request to `RequestID`.
   - *(Why it's outermost: if anything inside panics, `defer` catches it as the stack unwinds and returns a clean 500 JSON response instead of crashing your server.)*

2. **[`RequestID`](file:///Users/mac/Documents/backend/golang/library_api/internal/middleware/middleware.go#L15-L25) (Way IN)**:
   - Checks if the client sent an `X-Request-ID` header; if not, generates a new UUID (`uuid.NewString()`).
   - Stores that ID inside `r.Context()` (`context.WithValue(...)`) so downstream code (like `Logging` and `utils.ErrorResponse`) can read it.
   - Copies the ID onto the response header (`w.Header().Set("X-Request-ID", id)`).
   - Calls `next.ServeHTTP(w, r.WithContext(ctx))`, passing the updated request to `Logging`.

3. **[`Logging`](file:///Users/mac/Documents/backend/golang/library_api/internal/middleware/middleware.go#L39-L52) (Way IN)**:
   - Records `start := time.Now()`.
   - Wraps `w` in [`statusRecorder`](file:///Users/mac/Documents/backend/golang/library_api/internal/middleware/middleware.go#L54-L61).
   - *(Why [`statusRecorder`](file:///Users/mac/Documents/backend/golang/library_api/internal/middleware/middleware.go#L54-L61) exists: Go's standard `http.ResponseWriter` doesn't let you read back what status code was written! Wrapping `w` and overriding `WriteHeader(code int)` lets `statusRecorder` spy on the status code when your handler writes `201` or `400` and save it in `rec.status`.)*
   - Calls `next.ServeHTTP(rec, r)`, passing the request to `Timeout`.

4. **[`Timeout`](file:///Users/mac/Documents/backend/golang/library_api/internal/middleware/middleware.go#L64-L71) (Way IN)**:
   - Creates a context with a 5-second deadline (`context.WithTimeout(r.Context(), 5*time.Second)`).
   - Calls `next.ServeHTTP(w, r.WithContext(ctx))`, which enters `mux`.
   - *(Why this matters: your database queries use `r.Context()`. If a query takes longer than 5 seconds, the context cancels automatically and `pgx` aborts the slow query.)*

5. **`mux` → [`handler.CreateAuthor`](file:///Users/mac/Documents/backend/golang/library_api/internal/handler/author.go#L10-L24) (The Core)**:
   - Runs your validation and database query, and calls `utils.SuccessResponse` (which calls `w.WriteHeader(201)` — captured by `statusRecorder`!).

6. **Unwinding (Way OUT)**:
   - `Timeout`'s `defer cancel()` cleans up the timer.
   - [`Logging`](file:///Users/mac/Documents/backend/golang/library_api/internal/middleware/middleware.go#L44-L50) resumes right after `next.ServeHTTP(rec, r)`! Now the handler is finished, so `Logging` knows both `rec.status` (`201`) and `time.Since(start)` (how long the handler took), and logs them along with the `request_id`.
   - `RequestID` and `Recover` finish and the HTTP response goes back to the client.

---

Does that flow make sense? Whenever you're ready, we can look at writing your first handler test with `httptest`!