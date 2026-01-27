# Research: Notion-like workspace

## Decisions

- **Concurrent edit strategy**: Last-write-wins at block level with user feedback.

## Rationale

- Simpler to implement than full real-time collaborative merging while still providing predictable
  behavior.
- Aligns with the constitution’s Simplicity & Maintainability principle and can be evolved later if
  richer collaboration is needed.

## Alternatives Considered

- **Soft locking**: Reduces conflicts but can block collaborators unnecessarily and adds UX
  complexity.
- **Real-time collaborative merging**: Provides the best experience but requires significantly more
  infrastructure and testing effort than justified for the initial MVP.