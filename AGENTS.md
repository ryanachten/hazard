## Context
**DO NOT WRITE ANY CODE UNLESS EXPLICITLY REQUESTED.**

This is a learning project, it is especially important that I get familiar with writing Go code.
The agent's role in this project is to help plan, review the code and ensure that the code and architecture aligns
with idiomatic Golang and event-driven architectural principals.

For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan at `docs/plan.md` and the research at `docs/research.md`.

## Code conventions
- A variable's name should be independent of its type. Avoid redundancy in variable and parameter naming that can be inferred from the type.
- Avoid non-descriptive package names like `common` and `utility`. Prefer multiple packages focused on the domain type
- Avoid repetition. i.e. `bytes.Buffer` not `bytes.BytesBuffer` etc. 
- Use `.New()` method naming when returning a pointer to a given type
- Prefer use to structured logs via `slog` rather than `log``
- Avoid use of import aliases, unless it solves a problem - i.e. avoids long package names (like `pathfinding`), prevents awkward references (like `hazard.Hazard`)
- Prefer nil slice values via `var t []string` when declaring a slice over zero-length slices like `t := []string{}`

### References
- Dave Cheney's [Practical Go](https://dave.cheney.net/practical-go)
- Go [Code Review Comments](https://go.dev/wiki/CodeReviewComments)