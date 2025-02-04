-- 関数定義
CREATE OR REPLACE FUNCTION clean_schema(schema_name text) RETURNS void AS $$
DECLARE
    rec RECORD;
BEGIN
    -- 指定されたスキーマ内のテーブルを削除
    FOR rec IN (SELECT tablename FROM pg_tables WHERE schemaname = schema_name) LOOP
        EXECUTE 'DROP TABLE IF EXISTS ' || schema_name || '.' || rec.tablename || ' CASCADE;';
    END LOOP;

    -- 指定されたスキーマ内のシーケンスを削除
    FOR rec IN (SELECT sequencename FROM pg_sequences WHERE schemaname = schema_name) LOOP
        EXECUTE 'DROP SEQUENCE IF EXISTS ' || schema_name || '.' || rec.sequencename || ' CASCADE;';
    END LOOP;

    -- 指定されたスキーマ内の関数を削除
    FOR rec IN (SELECT proname, oidvectortypes(proargtypes) as args FROM pg_proc p
                JOIN pg_namespace n ON p.pronamespace = n.oid
                WHERE n.nspname = schema_name) LOOP
        EXECUTE 'DROP FUNCTION IF EXISTS ' || schema_name || '.' || rec.proname || '(' || rec.args || ') CASCADE;';
    END LOOP;

    -- 指定されたスキーマ内のENUMを削除
    FOR rec IN (SELECT typname FROM pg_type WHERE typtype = 'e' AND typnamespace = (SELECT oid FROM pg_namespace WHERE nspname = schema_name)) LOOP
        EXECUTE 'DROP TYPE IF EXISTS ' || schema_name || '.' || rec.typname || ' CASCADE;';
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 関数を実行(スキーマ毎)
SELECT clean_schema('orders');

-- 関数を削除
DROP FUNCTION IF EXISTS clean_schema;

-- トランケートを実行
TRUNCATE public.operation_histories CASCADE;
