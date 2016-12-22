DIRS = db rlib rcsv rrpt admin test
.PHONY:  test

rentroll: *.go ver.go
	@echo "\n\n**** Running target rentroll ****"
	@echo "Copying confdev.json to conf.json"
	cp confdev.json conf.json
	cp conf.json test/conf.json
	for dir in $(DIRS); do make -C $$dir;done
	go vet
	golint
	./mkver.sh
	go build

ver.go:
	./mkver.sh

clean:
	@echo "\n\n**** Running CLEAN for all directories ****"
	for dir in $(DIRS); do make -C $$dir clean;done
	@echo "\t*** GO CLEAN ***"
	go clean
	@echo "\t*** REMOVING FILES ***"
	rm -f rentroll ver.go conf.json rentroll.log *.out restore.sql rrbkup rrnewdb rrrestore example
	@echo "**** CLEAN COMPLETED ****"

test: package
	@echo "\n\n**** Running TESTS for all directories ****"
	rm -f test/*/err.txt
	for dir in $(DIRS); do make -C $$dir test;done
	@./errcheck.sh
	@echo "**** TEST COMPLETED ****"

man: rentroll.1
	cp rentroll.1 /usr/local/share/man/man1

instman:
	pushd tmp/rentroll;./installman.sh;popd

package: rentroll
	@echo "\n\n**** Running PACKAGE for rentroll ****"
	rm -rf tmp
	mkdir -p tmp/rentroll
	mkdir -p tmp/rentroll/man/man1/
	mkdir -p tmp/rentroll/example/csv
	cp rentroll.1 tmp/rentroll/man/man1
	for dir in $(DIRS); do make -C $$dir package;done
	cp rentroll ./tmp/rentroll/
	cp conf.json ./tmp/rentroll/
	cp -r html ./tmp/rentroll/
	if [ -e js ]; then cp -r js ./tmp/rentroll/ ; fi
	cp activate.sh update.sh ./tmp/rentroll/
	rm -f ./rrnewdb ./rrbkup ./rrrestore
	ln -s tmp/rentroll/rrnewdb
	ln -s tmp/rentroll/rrbkup
	ln -s tmp/rentroll/rrrestore
	@echo "**** PACKAGE COMPLETED ****"

publish: package
	cd tmp;tar cvf rentroll.tar rentroll; gzip rentroll.tar
	cd tmp;/usr/local/accord/bin/deployfile.sh rentroll.tar.gz jenkins-snapshot/rentroll/latest

pubimages:
	cd tmp/rentroll;find . -name "*.png" | tar -cf rrimages.tar -T - ;gzip rrimages.tar ;/usr/local/accord/bin/deployfile.sh rrimages.tar.gz jenkins-snapshot/rentroll/latest

pubjs:
	cd tmp/rentroll;tar czvf rrjs.tar.gz ./js;/usr/local/accord/bin/deployfile.sh rrjs.tar.gz jenkins-snapshot/rentroll/latest

pub: pubjs pubimages


all: clean rentroll test


try: clean rentroll package
