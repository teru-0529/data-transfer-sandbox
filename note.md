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
docker cp ./dist/dev(L202501)/dml-local.sql.gz product-db:/tmp/dump.sql.gz
docker exec -e PGPASSWORD=password -i product-db bash -c gzip -df -c "/tmp/dump.sql.gz ^| psql -U postgres -d productDB"

docker cp dd-dump1.sql.gz work-db:/tmp/dumpfile.sql.gz
docker exec -e PGPASSWORD=password -i work-db bash -c "gzip -d -c /tmp/dump.sql.gz | psql -U postgres -d workDB"

<!-- dump -->
docker exec -e PGPASSWORD=password -i work-db bash -c "pg_dump -U postgres -d workDB --data-only --schema=clean > /tmp/dump.sql && gzip /tmp/dump.sql"
docker cp work-db:/tmp/dump.sql.gz ./dd-dump3.sql.gz

❌
docker exec -e PGPASSWORD=password -i product-db sh
