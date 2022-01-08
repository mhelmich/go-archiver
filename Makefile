test:
	[ -d test-fixtures/tree2/empty-dir ] || mkdir -p test-fixtures/tree2/empty-dir
	go test -v -race ./...
.PHONY: test
