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

CREATE SCHEMA IF NOT EXISTS clean;
