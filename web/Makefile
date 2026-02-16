.PHONY: dist test

build: dep dist

build-admin:
	@pnpm install && pnpm exec turbo run build --filter=admin

dist:
	@pnpm run build

test:
	@pnpm run test

dep:
	@pnpm install
