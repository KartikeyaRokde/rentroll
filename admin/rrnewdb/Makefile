rrnewdb: *.sh

clean:
	rm -f rrnewdb
	@echo "\t*** CLEAN completed in admin/rrnewdb ***"

install: rrnewdb
	@echo "\t*** PACKAGE in admin/rrnewdb ***"

package: rrnewdb
	cp rrnewdb ../../tmp/rentroll/
	cp *.1 ../../tmp/rentroll/man/man1/
	@echo "\t*** PACKAGE completed in admin/rrnewdb ***"

packageqa: rrnewdb
	cp rrnewdb ../../tmp/rentroll/
	cp *.1 ../../tmp/rentrollqa/man/man1/
	@echo "\t*** PACKAGE completed in admin/rrnewdb ***"

test:
	@echo "*** TEST completed in admin/rrnewdb ***"

man:
	nroff -man rrnewdb.1
