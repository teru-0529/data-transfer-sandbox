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

--

--exclude-table-data を指定すると、スキーマは含めるが、データは除外できます。
--exclude-table-data はデータのみ除外し、スキーマは出力するため、
リストア時に target_table のテーブル定義は作成されますが、データは空の状態になります。

pg_dump -U postgres -d mydatabase \
  --exclude-table-data=public.target_table \
  -f dump.sql

複数のテーブルを除外する場合

pg_dump -U postgres -d mydatabase \
  --exclude-table-data=public.table1 \
  --exclude-table-data=public.table2 \
  -f dump.sql

解決策: --no-owner --no-privileges --schema=schema_name を使用
デフォルトの pg_dump では CREATE SCHEMA を含むため、
権限がない環境でリストアするとエラーになります。

そのため、以下のオプションを使ってスキーマ作成を省略し、
既存のスキーマ内に テーブルのみ作成 し、データも投入する ダンプを取得できます。

pg_dump -U postgres -d mydatabase \
  --schema=schema_name \
  --no-owner --no-privileges \
  -f dump.sql

各オプションの意味
--schema=schema_name
→ 指定したスキーマ内のテーブルとデータのみダンプ
--no-owner
→ OWNER TO を含めず、リストア時の権限エラーを防ぐ
--no-privileges
→ GRANT / REVOKE を含めず、不要な権限エラーを防ぐ

ダンプファイルの内容
CREATE SCHEMA は含まれない ✅
対象スキーマ内の CREATE TABLE は含まれる ✅
対象スキーマ内の COPY (データ) も含まれる ✅

ーーーーー

<!-- drop -->
docker cp ./dist/dev/v1.0.0(L202501)/clean.sql product-db:/tmp/clean.sql
docker exec -it product-db psql -U postgres -d product_db -f /tmp/clean.sql

<!-- ddl -->
docker cp ./dist/dev/v1.0.0(L202501)/ddl.sql.gz product-db:/tmp/ddl.sql.gz
docker exec -it product-db bash -c "gzip -d -c /tmp/ddl.sql.gz | psql -U postgres -d product_db 2>&1"

<!-- dml -->
docker cp ./dist/dev/v1.0.0(L202501)/dml.sql.gz product-db:/tmp/dml.sql.gz
docker exec -it product-db bash -c "gzip -d -c /tmp/dml.sql.gz | psql -U postgres -d product_db 2>&1"
