GETVERSION = export VERSION=$$(git describe --tags --always --dirty)

# Falls back to the redocly/cli container when there's no npx or only a Windows npx (WSL interop),
# as the latter can't work on UNC paths like \\wsl.localhost\...
REDOCLY = case "$$(command -v npx)" in \
		""|/mnt/*) redocly() { docker run --rm --user "$$(id -u):$$(id -g)" -v "$$PWD":/spec -w /spec redocly/cli "$$@"; } ;; \
		*) redocly() { npm_config_yes=true npx @redocly/cli "$$@"; } ;; \
	esac

# Drops the unconditional "now renders docs with Redoc 3" banner, keeping the exit code.
GENAPIDOC = $(REDOCLY); out=$$(redocly build-docs -t api/openapi-spec/redoc.hbs -o docs/api/index.html api/openapi-spec/ghostwire-v1.yaml 2>&1); rc=$$?; printf '%s\n' "$$out" | sed '/╔/,/╚/d;/^\s*$$/d'; exit $$rc

tools := gostwire gostdump lsallnifs lsnetdev

DOCKERFILE := deployments/gostwire/Dockerfile
TRIVY_IMAGE := aquasec/trivy:0.76.1@sha256:790fd66d1697395d8693d1acaa1ea702406a20f109b6fc7bac8db623eca6c06c

.PHONY: help clean redocsify deploy undeploy bump-debian

help: ## list available targets
	@# Derived from Gomega's Makefile (github.com/onsi/gomega) under MIT License
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

clean: ## cleans up build and testing artefacts
	rm -f $(tools) coverage.html coverage.out __debug_bin

redocsify: ## generate the API docs in docs/api
	@$(GENAPIDOC)

deploy: ## deploy Gostwire service exposed on host port 5000
	@$(GENAPIDOC)
	( \
		$(GETVERSION) \
		&& echo "deploying version" $$VERSION \
		&& scripts/docker-build.sh $(DOCKERFILE) \
			-t gostwire \
			--build-arg VERSION=$$VERSION \
			--build-context webappsrc=./webui \
	)
	docker compose -p gostwire -f deployments/gostwire/docker-compose.yaml up

undeploy: ## remove any Gostwire service deployment
	docker compose -p gostwire -f deployments/gostwire/docker-compose.yaml down

# Trivy exits with 3 only when it finds fixable vulnerabilities, so that other
# failures still abort.
bump-debian: ## list Debian security fixes missing from the image, then bump the Debian snapshot
	@set -e; \
	tar=$$(mktemp); trap 'rm -f "$$tar"' EXIT; \
	scan() { \
		scripts/docker-build.sh $(DOCKERFILE) -q --load -t gostwire:bump-debian \
			--build-context webappsrc=./webui >/dev/null && \
		docker save gostwire:bump-debian -o "$$tar" && \
		docker run --rm -v "$$tar:/image.tar:ro" -v gostwire-trivy-cache:/root/.cache $(TRIVY_IMAGE) \
			image --quiet --input /image.tar --scanners vuln --pkg-types os --ignore-unfixed --exit-code 3; \
	}; \
	echo "Scanning image with Debian snapshot $$(sed -n 's/^ARG DEBIAN_SNAPSHOT=//p' $(DOCKERFILE))..."; \
	scan || [ $$? -eq 3 ]; \
	snapshot=$$(date -u +%Y%m%dT%H%M%SZ); \
	sed -i "s/^ARG DEBIAN_SNAPSHOT=.*/ARG DEBIAN_SNAPSHOT=$$snapshot/" $(DOCKERFILE); \
	echo "Bumped to Debian snapshot $$snapshot, rescanning..."; \
	scan && echo "No Debian security fixes missing."
