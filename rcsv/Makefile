rrlib: *.go
	go vet
	golint
	go build
	go test
	go install

test:
	go test
	@echo "\t*** TEST completed in rcsv -- ALL TESTS PASS ***"

clean:
	go clean
	@echo "\t*** CLEAN completed in rcsv ***"

package:
	@echo "\t*** PACKAGE completed in rlib ***"
