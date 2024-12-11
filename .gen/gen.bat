@echo off

sqlboiler psql -c config/source.yaml
sqlboiler psql -c config/work.yaml

sqlboiler psql -c config/dist.yaml
