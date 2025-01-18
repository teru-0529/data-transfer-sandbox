@echo off

sqlboiler mysql -c config/legacy.yaml
sqlboiler psql -c config/work.yaml

sqlboiler psql -c config/product-orders.yaml
