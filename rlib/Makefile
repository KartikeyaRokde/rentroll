rrlib: *.go
	go vet
	golint
	go build
	go test
	go install

test:
	go test
	@echo "\t*** TEST completed in rrlib -- ALL TESTS PASS ***"

clean:
	go clean
	@echo "\t*** CLEAN completed in rrlib ***"

package:
	@echo "\t*** PACKAGE completed in rlib ***"
