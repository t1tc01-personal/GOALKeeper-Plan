<!--
Sync Impact Report
- Version change: N/A → 1.0.0
- Modified principles:
  - Principle 1 → Code Quality as a First-Class Citizen
  - Principle 2 → Testing Discipline & Coverage
  - Principle 3 → Consistent, Accessible User Experience
  - Principle 4 → Performance & Reliability Budgets
  - Principle 5 → Simplicity & Maintainability
- Added sections:
  - Non-Functional Standards (Performance, Reliability, Accessibility)
  - Development Workflow, Quality Gates, and Reviews
- Removed sections:
  - None
- Templates requiring updates:
  - ✅ .specify/templates/plan-template.md (Constitution Check aligned with principles)
  - ✅ .specify/templates/tasks-template.md (Testing expectations aligned with principles)
  - ✅ .specify/templates/spec-template.md (Success criteria guidance aligned with performance and UX requirements)
  - ✅ .specify/templates/agent-file-template.md (no constitution conflicts detected)
  - ✅ .specify/templates/checklist-template.md (no constitution conflicts detected)
- Follow-up TODOs:
  - None
-->

# GOALKeeper Plan Constitution

## Core Principles

### Code Quality as a First-Class Citizen

All production code MUST:
- Pass automated linting and formatting checks with zero errors.
- Avoid new unresolved `TODO`/`FIXME` comments unless linked to a tracked ticket.
- Be reviewed by at least one other engineer before merge, with explicit confirmation that
  naming, structure, and error handling are clear and maintainable.
- Isolate side effects and keep functions small (ideally under ~20 lines) to enable focused
  testing and easy refactoring.

Rationale: High, consistent code quality reduces defects, onboarding time, and long-term
maintenance cost.

### Testing Discipline & Coverage

Every meaningful behavior MUST be backed by automated tests proportional to its risk:
- P1 and security-sensitive user stories MUST have automated tests (unit and/or integration)
  that fail before implementation and pass before merge.
- New or modified backend logic SHOULD reach at least 80% line coverage, with critical paths
  (auth, money-like flows, data integrity) explicitly exercised.
- Bug fixes MUST include a regression test that fails without the fix.
- Tests MUST run in CI on every pull request and MUST be green before merge.

Rationale: A disciplined, risk-based testing strategy is the primary defense against
regressions and enables fast, safe iteration.

### Consistent, Accessible User Experience

User-facing features MUST:
- Follow a shared design system (typography, spacing, colors, components) defined at the
  project level; new components extend rather than bypass this system.
- Provide consistent behavior across screens for comparable actions (e.g., navigation,
  confirmations, error messages).
- Meet baseline accessibility expectations: keyboard navigability for all interactive
  controls, meaningful labels, and sufficient color contrast.
- Include clear, user-centered acceptance criteria in the spec that cover happy paths,
  error states, and empty states.

Rationale: Consistency and accessibility make the product predictable, inclusive, and
easier to evolve without surprising users.

### Performance & Reliability Budgets

Each feature MUST define and respect explicit performance and reliability expectations:
- Specs MUST record success criteria for latency, throughput, and perceived responsiveness
  where applicable (for example: target p95 latency, maximum acceptable frame drops).
- Plans MUST capture any non-default performance or reliability requirements and describe
  the strategy to meet them (caching, pagination, background work, etc.).
- Critical paths MUST be observable: logs, metrics, and/or traces are in place to detect
  and diagnose performance regressions.
- Changes that can materially impact performance or reliability MUST include a rollback or
  mitigation plan in the implementation tasks.

Rationale: Making performance and reliability explicit and observable prevents accidental
regressions and supports data-driven tradeoffs.

### Simplicity & Maintainability

The simplest solution that satisfies the principles above MUST be preferred:
- New abstractions, layers, or dependencies REQUIRE a short justification in the plan,
  especially if they add cross-cutting complexity.
- Implementation plans MUST document where complexity is concentrated and how it will be
  tested and owned.
- Features SHOULD be deliverable in small, independently testable slices that can ship
  without flag days or big-bang releases.
- Refactoring to reduce accidental complexity is a valid and encouraged task when paired
  with strong tests.

Rationale: Intentional simplicity keeps the system understandable and changeable as the
project grows.

## Non-Functional Standards (Performance, Reliability, Accessibility)

Non-functional requirements are first-class citizens in this project:
- Every spec MUST explicitly call out relevant non-functional constraints in the
  Requirements and Success Criteria sections (for example, latency, data retention,
  security, accessibility).
- For user-facing features, UX and accessibility acceptance criteria are REQUIRED, not
  optional.
- For backend or data-heavy features, performance and reliability expectations MUST be
  documented and traceable to monitoring or testing strategies in the plan and tasks.
- Any deviation from agreed non-functional standards (for example, temporary performance
  degradation) MUST be time-bounded and tracked as a follow-up task.

## Development Workflow, Quality Gates, and Reviews

The development workflow enforces the principles above via explicit gates:
- Before Phase 0 research: confirm that the feature can be described in independently
  testable user stories and that non-functional goals are roughly understood.
- Before Phase 1 design: complete the Constitution Check in the plan, verifying code
  quality, testing, UX, and performance expectations are captured.
- Before implementation tasks start: specs and plans MUST define acceptance criteria,
  test strategy, and performance/UX success metrics for P1 stories.
- Before merge: CI MUST be green, core tests MUST pass, and reviewers MUST confirm that
  relevant constitutional principles are respected (or explicitly justified if deviated).
- Before release: any known deviations from principles MUST be documented with owners and
  timelines for remediation.

## Governance

The constitution governs how we work and evolve this project:
- This constitution supersedes ad-hoc practices or historical habits that conflict with
  the principles defined here.
- Amendments MUST be proposed via pull request that clearly explains the rationale,
  impact on existing features, and version bump (major, minor, or patch).
- Material changes that affect live systems MUST include a migration or rollout plan,
  including any required performance or UX validations.
- All feature plans, specs, and task lists MUST include an explicit Constitution Check
  section or equivalent notes that reference the principles where relevant.
- Compliance is reviewed during code review and at major milestones; repeated deviations
  without justification are considered process failures to be addressed.

**Version**: 1.0.0 | **Ratified**: 2026-01-27 | **Last Amended**: 2026-01-27

