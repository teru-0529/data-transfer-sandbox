-- 更新日時の設定
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
  -- 更新日時
  NEW.updated_at := now();
  return NEW;
END;
$$ LANGUAGE plpgsql;

-- DDL変更時に'reload schema'を通知する
-- Create an event trigger function
CREATE OR REPLACE FUNCTION pgrst_watch() RETURNS event_trigger
  LANGUAGE plpgsql
  AS $$
BEGIN
  NOTIFY pgrst, 'reload schema';
END;
$$;

-- This event trigger will fire after every ddl_command_end event
CREATE EVENT TRIGGER pgrst_watch
  ON ddl_command_end
  EXECUTE PROCEDURE pgrst_watch();

CREATE SCHEMA IF NOT EXISTS work;
CREATE SCHEMA IF NOT EXISTS clean;

-- Create Table
DROP TABLE IF EXISTS work.clean_db CASCADE;
CREATE TABLE work.clean_db (
  dump_key varchar(3) NOT NULL default 'key' check (dump_key = 'key'),
  dump_file_name varchar(200) NOT NULL
);

-- Set Table Comment
COMMENT ON TABLE work.clean_db IS 'クリーンデータベース';

-- Set Column Comment
COMMENT ON COLUMN work.clean_db.dump_key IS 'キー';
COMMENT ON COLUMN work.clean_db.dump_file_name IS 'ダンプファイル名';

-- Set PK Constraint
ALTER TABLE work.clean_db ADD PRIMARY KEY (
  dump_key
);
