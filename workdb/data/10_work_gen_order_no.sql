-- is_master_table=false

-- 1.受注番号(gen_order_no)

-- Create Table
DROP TABLE IF EXISTS work.gen_order_no CASCADE;
CREATE TABLE work.gen_order_no (
  order_now varchar(10) NOT NULL check (order_now ~* '^RO-[0-9]{7}$'),
  product_idw varchar(5) NOT NULL check (product_idw ~* '^P[0-9]{4}$'),
  order_no integer NOT NULL,
  order_detail_no integer NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp,
  updated_at timestamp NOT NULL DEFAULT current_timestamp,
  created_by varchar(58),
  updated_by varchar(58)
);

-- Set Table Comment
COMMENT ON TABLE work.gen_order_no IS '受注番号';

-- Set Column Comment
COMMENT ON COLUMN work.gen_order_no.order_now IS '受注番号(WORK)';
COMMENT ON COLUMN work.gen_order_no.product_idw IS '商品ID(WORK)';
COMMENT ON COLUMN work.gen_order_no.order_no IS '受注番号';
COMMENT ON COLUMN work.gen_order_no.order_detail_no IS '受注明細番号';
COMMENT ON COLUMN work.gen_order_no.created_at IS '作成日時';
COMMENT ON COLUMN work.gen_order_no.updated_at IS '更新日時';
COMMENT ON COLUMN work.gen_order_no.created_by IS '作成者';
COMMENT ON COLUMN work.gen_order_no.updated_by IS '更新者';

-- Set PK Constraint
ALTER TABLE work.gen_order_no ADD PRIMARY KEY (
  order_now,
  product_idw
);

-- Set Unique Constraint
ALTER TABLE work.gen_order_no ADD CONSTRAINT gen_order_no_unique_1 UNIQUE (
  order_no,
  order_detail_no
);

-- Create 'set_update_at' Trigger
CREATE TRIGGER set_updated_at
  BEFORE UPDATE
  ON work.gen_order_no
  FOR EACH ROW
EXECUTE PROCEDURE set_updated_at();
