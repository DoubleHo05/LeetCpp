# One entry point per quality check. `make check` is the unified command:
# CI runs it, and so should every developer before pushing. The individual
# targets below are the single documented command for each check — they are
# what CI executes, byte for byte.

.PHONY: check lint format-check typecheck test \
	backend-lint backend-format-check backend-typecheck backend-test \
	frontend-lint frontend-format-check frontend-typecheck frontend-test

check: lint format-check typecheck test

lint: backend-lint frontend-lint

format-check: backend-format-check frontend-format-check

typecheck: backend-typecheck frontend-typecheck

test: backend-test frontend-test

backend-lint:
	cd backend && go vet ./...

backend-format-check:
	@out="$$(gofmt -l backend)"; if [ -n "$$out" ]; then echo "gofmt needed on:"; echo "$$out"; exit 1; fi

backend-typecheck:
	cd backend && go build ./...

backend-test:
	cd backend && go test ./...

frontend-lint:
	cd frontend && pnpm lint

frontend-format-check:
	cd frontend && pnpm format:check

frontend-typecheck:
	cd frontend && pnpm typecheck

frontend-test:
	cd frontend && pnpm test
