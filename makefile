SHELL := /bin/bash

VERSION ?= v3.0.8

ifeq ($(UNIQ_TAG), true)
	VERSION := $()-$(shell date +"%m-%d-%y_%H-%M-%S")
endif

REPO ?= harbor.shopeemobile.com/rds-kube-operator
IMG_ALY ?= ${REPO}/slowquery-analyzer:${VERSION}
IMG_PLAT ?= ${REPO}/slowquery-platform:${VERSION}
IMG_OPENAPI ?= ${REPO}/slowquery-openapi:${VERSION}
IMG_ALERT ?= ${REPO}/slowquery-alert:${VERSION}
IMG_CRON ?= ${REPO}/slowquery-crontab:${VERSION}
TEST_DIR ?= $(shell go list ./... | grep -v /thrid-party/ | grep -v /cmd/ | grep -v /conf/ | grep -v /doc/)

TEST_PARALLEL_NUM := $(shell [ -n "${GO_TEST_PARALLEL_NUM}" ] && echo "${GO_TEST_PARALLEL_NUM}" || echo "1")
TEST_PROGRAMS_NUM := $(shell [ -n "${GO_TEST_PROGRAMS_NUM}" ] && echo "${GO_TEST_PROGRAMS_NUM}" || echo "4")
TEST := go test -timeout=0 -v -parallel=${TEST_PARALLEL_NUM} -p=${TEST_PROGRAMS_NUM} -count=1 -cover -race
PKGS := $$(go list ./... | grep -v /thrid-party/ | grep -v /cmd/ | grep -v /conf/ | grep -v /doc/)

GOCOVERDIR := $(shell [ -n "${CI_PROJECT_DIR}" ] && echo "${CI_PROJECT_DIR}/coveragebin" || echo "$(shell pwd)/coveragebin")
DIFF_COVER_FAIL_UNDER = $(shell [ -n "${CI_DIFF_COVER_FAIL_UNDER}" ] && echo "${CI_DIFF_COVER_FAIL_UNDER}" || ([[ "${DIFF_COVER_BRANCH}" == "master" ]] && echo "70" || echo "50"))
DIFF_COVER_REF = $(shell [[ -n "${CI_PROJECT_NAMESPACE}" && "${CI_PROJECT_NAMESPACE}" != "shopee/cloud/dbaas/smart-slowquery" ]] && echo "ciupstream" || echo "origin")
DIFF_COVER_BRANCH = $(shell [ -n "${CI_MERGE_REQUEST_TARGET_BRANCH_NAME}" ] && echo "${CI_MERGE_REQUEST_TARGET_BRANCH_NAME}" || echo "master")
DIFF_COVER_PROJECT_PATH = $(shell [ -n "${CI_MERGE_REQUEST_PROJECT_PATH}" ] && echo "${CI_MERGE_REQUEST_PROJECT_PATH}" || echo "shopee/cloud/dbaas/smart-slowquery")

build-all: build-analyzer build-platform build-openapi build-alert build-cronjob

.PHONY: format
format:
	go vet $(PKGS)
	go fmt $(PKGS)

.PHONY: lint
lint: format
	golangci-lint run --max-issues-per-linter 0 --max-same-issues 0

.PHONY: deps
deps:
	env GONOSUMDB="git.garena.com" GONOPROXY="git.garena.com" GOPRIVATE="git.garena.com" go mod download
	env GONOSUMDB="git.garena.com" GONOPROXY="git.garena.com" GOPRIVATE="git.garena.com" go mod vendor

.PHONY: test
test:
	mkdir -p $(GOCOVERDIR)
	$(TEST) $(PKGS) -args -test.gocoverdir="$(GOCOVERDIR)"
	$(TEST) $(REMOVE_PKGS) -args -test.gocoverdir="$(GOCOVERDIR)"

	# generate combined coverage report in textfmt
	go tool covdata textfmt -i=$(GOCOVERDIR) -o ./coverage.out
	go tool covdata percent -i=$(GOCOVERDIR)

.PHONY: clean
clean:
	rm -fr bin/*
	rm -fr $(GOCOVERDIR)/cov*
	rm -fr coverage.out coverage.xml coverage.html diff_coverage.html diff_coverage.json
	cd ./ut_output && rm -rf coverage.out coverage.txt

.PHONY: coverage
coverage:
	gocov convert coverage.out | gocov-xml > coverage.xml
	gocov convert coverage.out | gocov-html > coverage.html
	echo "total test coverage:"
	go tool cover -func=coverage.out | tail -n 1

	echo "diff-cover with branch: $(DIFF_COVER_REF)/$(DIFF_COVER_BRANCH), fail-under: $(DIFF_COVER_FAIL_UNDER)"
	@if [ ${DIFF_COVER_REF} == "ciupstream" ]; then \
		git remote -v; \
		echo "diff-cover upstream path: $(DIFF_COVER_PROJECT_PATH)"; \
		git remote add ciupstream https://git.garena.com/$(DIFF_COVER_PROJECT_PATH).git || true; \
	fi
	git fetch $(DIFF_COVER_REF) $(DIFF_COVER_BRANCH)
	diff-cover coverage.xml \
	--compare-branch $(DIFF_COVER_REF)/$(DIFF_COVER_BRANCH) \
	--html-report diff_coverage.html \
	--json-report diff_coverage.json \
	--diff-range-notation '..' \
	--show-uncovered \
	--fail-under=$(DIFF_COVER_FAIL_UNDER)
	echo "diff coverage report in html: diff_coverage.html"
	echo "diff-cover result: `cat diff_coverage.json`"

.PHONY: build-analyzer
build-analyzer:
	go build -o bin/slowquery-analyzer ./cmd/analyzer/analyzer.go

.PHONY: build-platform
build-platform:
	go build -o bin/slowquery-platform ./cmd/platform/platform.go

.PHONY: build-openapi
build-openapi:
	go build -o bin/slowquery-openapi ./cmd/openapi/openapi.go

.PHONY: build-alert
build-alert:
	go build -o bin/slowquery-alert ./cmd/alert/alert.go

.PHONY: build-cronjob
build-cronjob:
	go build -o bin/slowquery-cronjob ./cmd/cronjob/daily_report.go

.PHONY: docker-build-analyzer
docker-build-analyzer:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/slowquery-analyzer ./cmd/analyzer/analyzer.go
	docker build -t ${IMG_ALY} -f ./Dockerfile-analyzer .

.PHONY: docker-push-analyzer
docker-push-analyzer:
	docker push ${IMG_ALY}

.PHONY: docker-build-platform
docker-build-platform:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/slowquery-platform ./cmd/platform/platform.go
	docker build -t ${IMG_PLAT} -f ./Dockerfile-platform .

.PHONY: docker-push-platform
docker-push-platform:
	docker push ${IMG_PLAT}

.PHONY: docker-build-openapi
docker-build-openapi:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/slowquery-openapi ./cmd/openapi/openapi.go
	docker build -t ${IMG_OPENAPI} -f ./Dockerfile-openapi .

.PHONY: docker-push-openapi
docker-push-openapi:
	docker push ${IMG_OPENAPI}

.PHONY: docker-build-cronjob
docker-build-cronjob:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/slowquery-cronjob ./cmd/cronjob/daily_report.go
	docker build -t ${IMG_CRON} -f ./Dockerfile-cronjob .

.PHONY: docker-push-cronjob
docker-push-cronjob:
	docker push ${IMG_CRON}

.PHONY: docker-build-alert
docker-build-alert:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/slowquery-alert ./cmd/alert/alert.go
	docker build -t ${IMG_ALERT} -f ./Dockerfile-alert .

.PHONY: docker-push-alert
docker-push-alert:
	docker push ${IMG_ALERT}
