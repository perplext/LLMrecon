# Atticked v0.9.0 unused retry/error-classification helpers

These four files (`retry.go`, `errors.go`, and their tests) implemented a
generic retry helper and a transient/permanent error taxonomy in
`src/provider/core`:

- `RetryableQuery[T]` — generic exponential-backoff retry loop.
- `RetryPolicy` / `DefaultRetryPolicy`.
- `TransientError` / `PermanentError` + `IsTransient` / `IsPermanent`.

## Why atticked (#304)

They were **fully dead and redundant**. Investigation while scoping #304
found:

- `RetryableQuery`, `TransientError`, `PermanentError`, `IsTransient`, and
  `IsPermanent` had **zero production callers** (only their own unit tests
  referenced them). `*TransientError{}` was never constructed anywhere
  outside tests.
- The providers' **actual** retry path does not use any of this. Each
  provider runs requests through `executeWithResilience` →
  `middleware.RetryMiddleware` → circuit breaker → rate limiter. The live
  retry decision is `middleware.RetryMiddleware.isRetryableError`, which
  **already** classifies `*core.ProviderError` by HTTP status (429 + 5xx →
  retryable) and honors `Retry-After`.

So #304's actionable premise — "RetryableQuery is dead code" — was correct.
Its other framing ("429s don't actually retry today") was **not**: a 429
surfaces as a `*core.ProviderError` and `middleware.RetryMiddleware`
retries it. The two retry systems were redundant; this one was the unused
duplicate.

## Disposition

Atticked rather than deleted, following the v0.10.0 #177 / v0.11.0 #228
pattern. If a generic, provider-agnostic retry helper is wanted later
(e.g. for call sites outside the middleware-wrapped provider methods), this
is a reasonable starting point — but it should be reconciled with
`middleware.RetryMiddleware` rather than reintroduced as a second parallel
system.

The live retry classification is covered by tests in
`src/provider/middleware/retry_test.go`.
