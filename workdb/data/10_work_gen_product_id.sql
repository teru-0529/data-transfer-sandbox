-- is_master_table=false

-- 2.商品ID(gen_product_id)

-- Create Table
DROP TABLE IF EXISTS work.gen_product_id CASCADE;
CREATE TABLE work.gen_product_id (
  product_idw varchar(5) NOT NULL check (product_idw ~* '^P[0-9]{4}$'),
  product_name varchar(30) NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp,
  updated_at timestamp NOT NULL DEFAULT current_timestamp,
  created_by varchar(58),
  updated_by varchar(58)
);

-- Set Table Comment
COMMENT ON TABLE work.gen_product_id IS '商品ID';

-- Set Column Comment
COMMENT ON COLUMN work.gen_product_id.product_idw IS '商品ID(WORK)';
COMMENT ON COLUMN work.gen_product_id.product_name IS '商品名';
COMMENT ON COLUMN work.gen_product_id.created_at IS '作成日時';
COMMENT ON COLUMN work.gen_product_id.updated_at IS '更新日時';
COMMENT ON COLUMN work.gen_product_id.created_by IS '作成者';
COMMENT ON COLUMN work.gen_product_id.updated_by IS '更新者';

-- Set PK Constraint
ALTER TABLE work.gen_product_id ADD PRIMARY KEY (
  product_idw
);

-- Set Unique Constraint
ALTER TABLE work.gen_product_id ADD CONSTRAINT gen_product_id_unique_1 UNIQUE (
  product_name
);

-- Create 'set_update_at' Trigger
CREATE TRIGGER set_updated_at
  BEFORE UPDATE
  ON work.gen_product_id
  FOR EACH ROW
EXECUTE PROCEDURE set_updated_at();
