notes: *.go
	cp ../../confdev.json conf.json
	go vet
	golint
	go build
	@echo "*** Completed in Notes ***"

clean:
	rm -f notes.log log notes [a-z] conf.json
	@echo "*** CLEAN completed in Notes ***"

test:	notes
	./functest.sh
	@echo "*** TEST completed in Notes ***"

package:
	@echo "*** PACKAGE completed in Notes ***"
