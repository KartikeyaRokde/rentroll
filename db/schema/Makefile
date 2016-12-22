newdb:
	#mysql --no-defaults < schema.sql
	cp ../../conf.json conf.json
	./replace_placeholders.sh
	@echo "*** completed in db/schema ***"

doit:
	mysql --no-defaults < schema.sql

clean:
	rm -f conf.json
	@echo "*** CLEAN completed in db/schema ***"

test:
	@echo "*** TEST completed in db/schema ***"

package:
	cp schema.sql ../../tmp/rentroll/
	@echo "*** PACKAGE completed in db/schema ***"
