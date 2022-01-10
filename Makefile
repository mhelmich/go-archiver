test:
	[ -d test-fixtures/tree2/empty-dir ] || mkdir -p test-fixtures/tree2/empty-dir
	[ -d test-fixtures/tree5/.git ] || mkdir -p test-fixtures/tree5/.git
	[ -d test-fixtures/tree5/.git/.holder ] || touch test-fixtures/tree5/.git/.holder
	go test -v -race ./...
.PHONY: test
