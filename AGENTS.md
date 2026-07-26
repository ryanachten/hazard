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

### References
- Dave Cheney's [Practical Go](https://dave.cheney.net/practical-go)