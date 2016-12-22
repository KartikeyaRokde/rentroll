#!/bin/bash

source ../../confreader.sh

#  Create sql file to create new user
touch create_accord_user.sql
echo "CREATE USER IF NOT EXISTS '$rrdbuser'@'$rrdbhost';" > create_accord_user.sql
echo "GRANT ALL ON $rrdbname.* TO '$rrdbuser'@'$rrdbhost';" >> create_accord_user.sql

#  Create accord user
if [ "$dbpass" ]; then
	echo -e "\t** SQL >> Creating RentRoll user using admin username & password **"
	mysql -h "$dbhost" -u "$dbuser" "-p$dbpass" "$dbname" < "create_accord_user.sql"
	cat create_accord_user.sql;
else
	echo -e "\t** SQL >> Creating RentRoll user without admin password **"
	mysql -h "$dbhost" -u "$dbuser" "$dbname" < "create_accord_user.sql"
	cat create_accord_user.sql;
fi;

rm ~/.mylogin.cnf
#  Create mysql config using mysql config editor
if [ "$rrdbpass" ]; then
	echo -e "\t** SQL >> Creating mysql credentials config with username and password **"
	mysql_config_editor set --user=$rrdbuser --password
else
	echo -e "\t** SQL >> Creating mysql credentials config without password **"
	mysql_config_editor set --user=$rrdbuser
fi;

#  Make sure that the test machine has the same dependent
#  databases as the one(s) we're using to test with...
echo -e "\n\n**** SQL >> Firing up accord database ****"
mysql --no-defaults accord < accord.sql
echo -e "**** Database fired up successfully ****\n\n"
