# Pull Request Guidelines

This document outlines the workflow and quality standards for contributing pull requests to TownHall.

---

## 1. Branching Strategy

- **Base branch**: `main` is the primary production/integration branch.
- **Feature/Fix branches**: Branch off `main` with semantic prefixes:
  - `feat/<feature-name>` — New capabilities or user-facing changes (e.g. `feat/username-onboarding`).
  - `fix/<bug-name>` — Bug fixes (e.g. `fix/refresh-token-race`).
  - `chore/<task-name>` — Dependency upgrades, CI updates, configuration cleanup.
  - `docs/<topic>` — Documentation or guidelines changes.

---

## 2. Commit Message Conventions

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject in imperative, lowercase, no trailing period>

[optional body explaining motivation and breaking changes]
```

- **Allowed types**: `feat`, `fix`, `refactor`, `perf`, `docs`, `style`, `chore`, `test`.
- **Scopes**: `auth`, `ui`, `db`, `middleware`, `config`, `validation`, `ci`, `deps`, etc.
- **Examples**:
  - `feat(auth): implement username availability check endpoint`
  - `feat(ui): add onboarding screen for initial username setup`
  - `chore(ui): remove Vitest and testing library suites`
  - `fix(auth): handle null username serialization in ToUserResponse`

---

## 3. Mandatory Pre-Flight Verification Gates

Before submitting or requesting review on any PR, run and confirm the appropriate verification gates locally.

### Backend Changes (`Go`, `Ent`, `API`)
Run from repo root:
```bash
make check   # Runs: go vet ./... + golangci-lint run ./... + go test ./...
make build   # Builds ./tmp/main
```
Both commands must pass with **0 issues**.

If Ent schemas were altered:
```bash
make ent-generate
```
Never manually edit files outside `ent/schema/`.

### Frontend Changes (`ui/`, React, Vite, TS)
Run from `ui/`:
```bash
bun run typecheck   # tsc --noEmit -p tsconfig.app.json
bun run lint        # eslint .
bun run build       # tsc -b && vite build
```
All three commands must pass cleanly without warnings or errors.

### Backend-Frontend Contract Mirror
When updating request or response models in `internal/auth/auth_model.go`, ensure corresponding TypeScript types in `ui/src/features/auth/authApi.ts` match exact JSON keys (`snake_case`) in the same pull request.

---

## 4. Submitting a Pull Request

1. Push your branch to GitHub:
   ```bash
   git push -u origin <branch-name>
   ```
2. Open a Pull Request targeting `main`.
3. Complete all sections of the Pull Request template:
   - Provide a clear summary of the problem and the solution.
   - Check the relevant checkboxes in the verification checklist.
   - Include UI screenshots or video walkthroughs for visible frontend modifications.
4. Ensure GitHub Actions CI checks (`lint`, `build-vet-test`, `ui-lint-build`) are all green.
5. Address any reviewer feedback iteratively on the same branch.
