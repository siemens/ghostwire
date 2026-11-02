GOGEN = go generate .

GETGITVERSION = export GIT_VERSION=$$(sed -n 's/^const SemVersion = "\(.*\)"$$/\1/p' defs_version.go)

# Falls back to the redocly/cli container when there's no npx or only a Windows npx (WSL interop),
# as the latter can't work on UNC paths like \\wsl.localhost\...
REDOCLY = case "$$(command -v npx)" in \
		""|/mnt/*) redocly() { docker run --rm --user "$$(id -u):$$(id -g)" -v "$$PWD":/spec -w /spec redocly/cli "$$@"; } ;; \
		*) redocly() { npm_config_yes=true npx @redocly/cli "$$@"; } ;; \
	esac

# Drops the unconditional "now renders docs with Redoc 3" banner, keeping the exit code.
GENAPIDOC = $(REDOCLY); out=$$(redocly build-docs -t api/openapi-spec/redoc.hbs -o docs/api/index.html api/openapi-spec/ghostwire-v1.yaml 2>&1); rc=$$?; printf '%s\n' "$$out" | sed '/╔/,/╚/d;/^\s*$$/d'; exit $$rc

tools := gostwire gostdump lsallnifs lsnetdev

.PHONY: help clean redocsify deploy undeploy

help: ## list available targets
	@# Derived from Gomega's Makefile (github.com/onsi/gomega) under MIT License
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

clean: ## cleans up build and testing artefacts
	rm -f $(tools) coverage.html coverage.out __debug_bin

redocsify: ## generate the API docs in docs/api
	@$(GENAPIDOC)

deploy: ## deploy Gostwire service exposed on host port 5000
	$(GOGEN)
	@$(GENAPIDOC)
	( \
		$(GETGITVERSION) \
		&& echo "deploying version" $$GIT_VERSION \
		&& scripts/docker-build.sh deployments/gostwire/Dockerfile \
			-t gostwire \
			--build-arg REACT_APP_GIT_VERSION=$$GIT_VERSION \
			--build-context webappsrc=./webui \
	)
	docker compose -p gostwire -f deployments/gostwire/docker-compose.yaml up

undeploy: ## remove any Gostwire service deployment
	docker compose -p gostwire -f deployments/gostwire/docker-compose.yaml down
