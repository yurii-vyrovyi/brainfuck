.PHONY: deps
deps:
	go mod tidy

.PHONY: unit_test
unit_test:
	go test -count=1 -v ./...


.PHONY: mockgen
mockgen:
	GO111MODULE=off go install github.com/golang/mock/mockgen
	go generate ./...