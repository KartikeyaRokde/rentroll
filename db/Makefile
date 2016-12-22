DIRS = schema

db:
	for dir in $(DIRS); do make -C $$dir;done

clean:
	for dir in $(DIRS); do make -C $$dir clean;done
	@echo "\t*** CLEAN completed in db ***"

test:
	for dir in $(DIRS); do make -C $$dir test;done
	@echo "\t*** TEST completed in db ***"

package:
	for dir in $(DIRS); do make -C $$dir package;done
	@echo "\t*** PACKAGE completed in db/schema ***"
