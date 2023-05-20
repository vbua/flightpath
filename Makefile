UNIT_COVERAGE_MIN = 80
REPO_NAME = flightpath

fmt:
	$(info run go fmt)
	@go fmt ./...

gofumpt:
	$(info run gofumpt)
	@gofumpt -w .

vet:
	$(info run go vet)
	@go vet ./...

deps_check:
	$(info run dependency analyzer)
	@go list -json -deps ./cmd/flightpath | docker run -i --rm \
      sonatypecommunity/nancy:v1-alpine nancy sleuth

lint:
	$(info run linter)
	@go mod vendor
	@docker run --rm -v "$$(pwd)":/work:ro -w /work -it golangci/golangci-lint:latest golangci-lint run -v
	@rm -Rf vendor

security_scan:
	$(info run security scan)
	@go mod download
	@CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o flightpath ./cmd/flightpath
	@docker build -t "$(REPO_NAME):local" .
	@rm ./flightpath
	@docker save "$(REPO_NAME):local" > image.tar
	@docker run --rm -it -v "$$(pwd):/work" aquasec/trivy image --input /work/image.tar
	@rm image.tar

unit:
	$(info run tests with coverage)
	@go test -race -coverpkg=./... -coverprofile=cover_profile.out.tmp $$(go list ./...)
	@grep -v -e "mock" < cover_profile.out.tmp > cover_profile.out
	@rm cover_profile.out.tmp
	@CUR_COVERAGE=$$(go tool cover -func=cover_profile.out | tail -n 1 | awk '{ print $$3 }' | sed -e 's/^\([0-9]*\).*$$/\1/g'); \
	rm cover_profile.out; \
	if [ $$CUR_COVERAGE -lt $(UNIT_COVERAGE_MIN) ]; then \
		echo "coverage is not enough $$CUR_COVERAGE < $(UNIT_COVERAGE_MIN)"; \
		exit 1; \
	else \
		echo "coverage is enough $$CUR_COVERAGE >= $(UNIT_COVERAGE_MIN)"; \
	fi

test: gofumpt fmt vet unit deps_check lint security_scan