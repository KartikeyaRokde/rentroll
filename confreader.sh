function readConfigJson {
  
  UNAMESTR=`uname`
  if [[ "$UNAMESTR" == 'Linux' ]]; then
    SED_EXTENDED='-r'
  elif [[ "$UNAMESTR" == 'Darwin' ]]; then
    SED_EXTENDED='-E'
  fi; 


  VALUE=`grep -m 1 "\"${2}\"" ${1} | sed ${SED_EXTENDED} 's/^ *//;s/.*: *"//;s/",?//'`

  if [ ! "$VALUE" ]; then
    echo ""
  else
    echo $VALUE
  fi; 
}

#  Reading database configuration
dbhost=$(readConfigJson conf.json Dbhost);
dbuser=$(readConfigJson conf.json Dbuser);
dbpass=$(readConfigJson conf.json Dbpass);
rrdbhost=$(readConfigJson conf.json RRDbhost);
rrdbname=$(readConfigJson conf.json RRDbname);
rrdbuser=$(readConfigJson conf.json RRDbuser);
rrdbpass=$(readConfigJson conf.json RRDbpass);