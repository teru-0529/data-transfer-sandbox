# note

> [!IMPORTANT]
> go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-mysql@latest

-----

> [!IMPORTANT]
> go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest

-----

> [!TIP]
> Helpful advice for doing things better or more easily.

-----

-----

> [!WARNING]
> Urgent info that needs immediate user attention to avoid problems.

-----

> [!CAUTION]
> Advises about risks or negative outcomes of certain actions.

<!-- Load -->
docker cp dd-dump1.sql.gz work-db:/tmp/dumpfile.sql.gz
docker exec -e PGPASSWORD=password -i work-db bash -c "gzip -d -c /tmp/dump.sql.gz | psql -U postgres -d workDB"

<!-- dump -->
docker exec -e PGPASSWORD=password -i work-db bash -c "pg_dump -U postgres -d workDB --data-only --schema=clean > /tmp/dump.sql && gzip /tmp/dump.sql"
docker cp work-db:/tmp/dump.sql.gz ./dd-dump3.sql.gz

``` bat
ALTER TABLE inventories.inventory_histories ADD CONSTRAINT inventory_histories_inventory_type_check CHECK (
  CASE
    -- 在庫変動種類が「倉庫間移動入庫」「仕入入庫」「売上返品入庫」の場合、変動数量が1以上であること
    WHEN inventory_type = 'MOVE_WAREHOUSEMENT' AND variable_quantity <= 0 THEN FALSE
    WHEN inventory_type = 'PURCHASE' AND variable_quantity <= 0 THEN FALSE
    WHEN inventory_type = 'SALES_RETURN' AND variable_quantity <= 0 THEN FALSE
    -- 在庫変動種類が「倉庫間移動出庫」「売上出庫」「仕入返品出庫」の場合、変動数量が-1以下であること
    WHEN inventory_type = 'MOVE_SHIPPMENT' AND variable_quantity >= 0 THEN FALSE
    WHEN inventory_type = 'SELES' AND variable_quantity >= 0 THEN FALSE
    WHEN inventory_type = 'PURCHASE_RETURN' AND variable_quantity >= 0 THEN FALSE
    -- 在庫変動種類が「倉庫間移動入庫」「倉庫間移動出庫」の場合、変動金額が0であること
    WHEN inventory_type = 'MOVE_WAREHOUSEMENT' AND variable_amount != 0.00 THEN FALSE
    WHEN inventory_type = 'MOVE_SHIPPMENT' AND variable_amount != 0.00 THEN FALSE
    -- 在庫変動種類が「仕入入庫」「売上返品入庫」の場合、変動金額が0より大きい値であること
    WHEN inventory_type = 'PURCHASE' AND variable_amount <= 0.00 THEN FALSE
    WHEN inventory_type = 'SALES_RETURN' AND variable_amount <= 0.00 THEN FALSE
    -- 在庫変動種類が「売上出庫」「仕入返品出庫」の場合、変動金額が0より小さい値であること
    WHEN inventory_type = 'SELES' AND variable_amount >= 0.00 THEN FALSE
    WHEN inventory_type = 'PURCHASE_RETURN' AND variable_amount >= 0.00 THEN FALSE
    ELSE TRUE
  END
);
```
