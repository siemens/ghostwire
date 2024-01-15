# - Statically compiling Go programs: https://www.arp242.net/static-go.html
# - Why your Go programs can surprisingly be dynamically linked,
#   https://utcc.utoronto.ca/~cks/space/blog/programming/GoWhyNotStaticLinked
# - Go programs and Linux glibc versioning,
#   https://utcc.utoronto.ca/~cks/space/blog/programming/GoAndGlibcVersioning

GOSTATIC = -ldflags="-s -w -extldflags=-static" \
	-tags=osusergo,netgo,sqlite_omit_load_extension

GOGEN = go generate .

GETGITVERSION = export GIT_VERSION=$$(awk -n 'match($$0, /^const SemVersion = "(.*)"$$/, v) { print v[1]; }' defs_version.go)

GENAPIDOC = npm_config_yes=true npx @redocly/cli build-docs -o docs/api/index.html api/openapi-spec/ghostwire-v1.yaml

tools := gostwire gostdump lsallnifs

.PHONY: help pkgsite docsify redocsify build rebuild test test-all report clean deploy undeploy coverage grype lsallnifs install-tools vuln dist

help: ## list available targets
	@# Derived from Gomega's Makefile (github.com/onsi/gomega) under MIT License
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

clean: ## cleans up build and testing artefacts
	rm -f $(tools)
	rm -f coverage.html coverage.out __debug_bin gostwire

test: ## run tests as root as well as an ordinary user, without KinD
	go test -v -p 1 -tags=matchers -exec sudo ./...
	go test -v -p 1 -tags=matchers ./...

test-all: ## run tests as root as well as an ordinary user, including KinD
	go test -v -p 1 -tags=matchers,kind -exec sudo ./...
	go test -v -p 1 -tags=matchers,kind ./...

grype: ## run grype vul scan on sources
	@scripts/grype.sh

docsify: ## run a docsify HTTP server on port 3300 (and 3301)
	@$(GENAPIDOC)
	@scripts/docsify.sh ./docs

redocsify: ## run redoc-cli in serve mode on port 3400
	@(cd docs/api && npx @redocly/openapi-cli preview-docs --port 3400 --config=../../redoc-theme.json ../../api/openapi-spec/ghostwire-v1.yaml)
report: ## run goreportcard on this module
	@./scripts/goreportcard.sh

build: ## build the Gostwire stripped static binary
	@$(GENAPIDOC)
	go build -v $(GOSTATIC) ./cmd/gostwire
	@file gostwire

build-embedded: ## build the Gostwire stripped static binary with embedded web UI
	@$(GENAPIDOC)
	( \
		$(GETGITVERSION) \
		cd webui \
		REACT_APP_GIT_VERSION=$$GIT_VERSION yarn build \
	)
	go build -v $(GOSTATIC),webui ./cmd/gostwire
	@file gostwire

pprof: ## build the Gostwire static binary with pprof support enabled
	go run -exec sudo -v -tags osusergo,netgo,pprof -ldflags="-extldflags=-static" ./cmd/gostwire

dist: ## build multi-arch image (amd64, arm64) and push to local running registry on port 5000.
	$(GOGEN)
	( \
		$(GETGITVERSION) \
		&& scripts/multiarch-builder.sh \
			--build-arg GIT_VERSION=$$GIT_VERSION \
			--build-context webappsrc=./webui \
	)

deploy: ## deploy Gostwire service exposed on host port 5999
	$(GOGEN)
	@$(GENAPIDOC)
	( \
		$(GETGITVERSION) \
		&& echo "deploying version" $$GIT_VERSION \
		&& scripts/docker-build.sh deployments/gostwire/Dockerfile \
			-t gostwire \
			--build-arg GIT_VERSION=$$GIT_VERSION \
			--build-context webappsrc=./webui \
	)
	docker compose -p gostwire -f deployments/gostwire/docker-compose.yaml up

pkgsite: ## serves Go documentation on port 6060
	@echo "navigate to: http://localhost:6060/github.com/siemens/ghostwire/v2"
	@scripts/pkgsite.sh

pprofdeploy: ## deploy Gostwire service with pprof support on host port 5000
	$(GOGEN)
	( \
		$(GETGITVERSION) \
		&& echo "deploying version" $$GIT_VERSION \
		&& docker buildx build -t gostwire -f deployments/gostwire/Dockerfile \
			--build-arg GIT_VERSION=$$GIT_VERSION \
			--build-arg TAGS="pprof,osusergo,netgo" \
			--build-arg LDFLAGS="" \
			--build-arg GITTOKEN="${GITTOKEN}" \
			. \
	)
	docker compose -p gostwire -f deployments/gostwire/docker-compose.yaml up

undeploy: ## remove any Gostwire service deployment
	docker compose -p gostwire -f deployments/gostwire/docker-compose.yaml down

install-tools: ## install nerdctl and CNI plugins
	@scripts/setup-nerdctl-and-friends.sh

coverage: ## run all tests and aggregate coverage results into coverage.out
	@scripts/cov.sh

lsallnifs: ## list all network interfaces with their configuration in all network namespaces
	go run -v -exec sudo ./cmd/lsallnifs -x

vuln: ## run go vulnerabilities check
	@scripts/vuln.sh

yarnsetup: ## set up yarn v4 correctly
	cd webui && \
	rm -f .yarnrc.yml && \
	rm -rf .yarn/ && \
	rm -rf node_modules && \
	yarn set version berry && \
	yarn config set nodeLinker node-modules && \
	yarn install
	#yarn eslint --init
