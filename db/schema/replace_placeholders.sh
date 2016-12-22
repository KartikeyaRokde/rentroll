#!/bin/bash

source ../../confreader.sh

# Replace SQL file contents
sed -i.bak s/{{rrdbuser}}/$rrdbuser/g schema.sql
sed -i.bak s/{{rrdbhost}}/$rrdbhost/g schema.sql
