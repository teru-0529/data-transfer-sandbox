-- is_master_table=false

-- 1.担当者(operators)

-- Create Table
DROP TABLE IF EXISTS cleansing.operators CASCADE;
CREATE TABLE cleansing.operators (
  operator_id varchar(5) NOT NULL check (LENGTH(operator_id) = 5),
  operator_name varchar(30) NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp,
  updated_at timestamp NOT NULL DEFAULT current_timestamp,
  created_by varchar(58),
  updated_by varchar(58)
);

-- Set Table Comment
COMMENT ON TABLE cleansing.operators IS '担当者';

-- Set Column Comment
COMMENT ON COLUMN cleansing.operators.operator_id IS '担当者ID';
COMMENT ON COLUMN cleansing.operators.operator_name IS '担当者名';
COMMENT ON COLUMN cleansing.operators.created_at IS '作成日時';
COMMENT ON COLUMN cleansing.operators.updated_at IS '更新日時';
COMMENT ON COLUMN cleansing.operators.created_by IS '作成者';
COMMENT ON COLUMN cleansing.operators.updated_by IS '更新者';

-- Set PK Constraint
ALTER TABLE cleansing.operators ADD PRIMARY KEY (
  operator_id
);

-- Create 'set_update_at' Trigger
CREATE TRIGGER set_updated_at
  BEFORE UPDATE
  ON cleansing.operators
  FOR EACH ROW
EXECUTE PROCEDURE set_updated_at();
