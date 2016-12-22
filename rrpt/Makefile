rrlib: *.go
	go vet
	golint
	go build
	go test
	go install

test:
	go test
	@echo "\t*** TESTS COMPLETE in rrpt -- ALL TESTS PASS ***"

clean:
	go clean
	@echo "\t*** CLEAN completed in rrpt ***"

package:
	@echo "\t*** PACKAGE completed in rrpt ***"
