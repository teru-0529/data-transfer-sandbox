@echo off

sqlboiler mysql -c config/legacy.yaml
sqlboiler psql -c config/clean.yaml

sqlboiler psql -c config/product-orders.yaml
