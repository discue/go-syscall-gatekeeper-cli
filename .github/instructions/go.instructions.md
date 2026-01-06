---
applyTo: "**/*.go"
---

# Libraries
- Use https://github.com/kelseyhightower/envconfig for environment configuration
- Use https://github.com/stfsy/go-api-kit for http handler configuration, endpoints, middlewares, sending http responses
- Use https://github.com/stfsy/go-api-key for API key creation and valiation
- Use https://github.com/stfsy/go-argon2id for secure hashing and hash verification
- Use https://github.com/stretchr/testify for making assertions in test files
- Use https://github.com/go-playground/validator for struct validation

# Testing
- Creates test cases for all new functions, test happy paths and edge cases.

# Error Handling
- Always check and handle errors returned from called functions. When propagating an error, wrap it with concise, useful context using `fmt.Errorf` and the `%w` verb so callers can inspect the original error (`errors.Is` / `errors.As`).
- Assign and check errors immediately (the two-line pattern). Do not pre-declare `err` (or result) and assign later — that pattern is error-prone and harder to read.
- Include actionable context in the wrapping message (what operation failed and any non-sensitive identifiers). Do not include secrets or full tokens in error messages.
- Keep error checks explicit and readable; avoid hiding checks inside nested expressions or single-line idioms.

## Preferred Error Handling Style
```go
result, err := someMethod()
if err != nil {
	return nil, fmt.Errorf("someMethod failed: %w", err)
}
```

- **Do NOT use the pre-declare-and-assign-later pattern:**
```go
var err error
var result SomeType
result, err = someMethod()
if err != nil {
	// handle error
}
```

- When returning wrapped errors, ensure you return the correct zero values for other return types (for example `return nil, fmt.Errorf(...)` when the first return value is a pointer or slice).

- Example of adding helpful context while avoiding secrets:
```go
user, err := repo.GetUserByID(ctx, id)
if err != nil {
	return nil, fmt.Errorf("get user by id %s: %w", id, err)
}
```

# Code Style
- Use idiomatic Go style: follow effective Go and Go community conventions.
- Use gofmt for formatting and goimports for import management.
- Use clear, descriptive names for variables, functions, and types.
- Keep functions small and focused; prefer composition over inheritance.
- Add comments for exported functions, types, and complex logic.
- Avoid global variables; use dependency injection where possible.
- Group related code into packages; avoid circular dependencies.
- Use error wrapping and context for error handling.
- Prefer explicitness over cleverness; optimize for readability and maintainability.

# Context Management
- Only create a new context with `context.Background()` in exceptional situations or in test cases. In all other cases, use the parent context, e.g., the context of the http request. If beneficial, create a new child context to be able to cancel only the child request.

# Commit Guidelines
- Use Conventional Commits for all commit messages. 
- The format is:
```
<type>[scope]: <description>
[optional body]
[optional footer(s)]
```

- The description should be concise and in the imperative mood (e.g., "add new endpoint" not "added new endpoint").
- Use the body to explain what and why vs. how, if necessary.
- Reference issues or PRs in the footer when relevant (e.g., "Closes #123").

## Common scopes:
- feat:     A new feature
- fix:      A bug fix
- docs:     Documentation only changes
- style:    Changes that do not affect the meaning of the code (white-space, formatting, etc)
- refactor: A code change that neither fixes a bug nor adds a feature
- perf:     A code change that improves performance
- test:     Adding or correcting tests
- chore:    Other changes that don't modify src or test files
