## Description

<!-- Provide a brief summary of the changes made and the motivation/context behind them. -->

## Type of Change

- [ ] `feat`: A new feature
- [ ] `fix`: A bug fix
- [ ] `refactor`: Code change that neither fixes a bug nor adds a feature
- [ ] `perf`: A code change that improves performance
- [ ] `docs`: Documentation updates
- [ ] `style`: Formatting, missing semi-colons, etc.; no production code change
- [ ] `chore`: Tooling, build process, dependency updates, maintenance

## Scope / Areas Affected

- [ ] Backend (`cmd/api`, `internal/auth`, `internal/middleware`, `pkg/*`)
- [ ] Database / Ent (`ent/schema/`, migrations)
- [ ] Frontend (`ui/src/`)
- [ ] Observability (`observability/`, metrics, logs)
- [ ] CI / Tooling (`.github/`, `Makefile`)

## Related Issues

<!-- Link any related issues: e.g. Closes #123, Fixes #456 -->

## Verification & Pre-flight Checklist

### Backend Gate (Required for Go changes)
- [ ] `make check` ran freshly and passed (vet + golangci-lint + tests)
- [ ] `make build` built `./tmp/main` successfully
- [ ] Ent schemas regenerated if modified (`make ent-generate`)
- [ ] No destructive auto-migrations in production paths

### Frontend Gate (Required for UI changes)
- [ ] `bun run typecheck` passed (no TypeScript errors)
- [ ] `bun run lint` passed (no ESLint errors/warnings)
- [ ] `bun run build` built `dist/` successfully
- [ ] Frontend contracts match Go models in `internal/auth/auth_model.go`

### General
- [ ] Conventional commit messages used
- [ ] No sensitive credentials, secrets, or unneeded logs committed
- [ ] Documentation / `AGENTS.md` updated if workflows or contracts changed

## Screenshots / Demos (Optional for UI changes)

<!-- If applicable, paste screenshots, GIFs, or walkthrough links showing the UI changes in action. -->
