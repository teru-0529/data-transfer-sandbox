-- is_master_table=false

-- 3.受注(order)

-- Create Table
DROP TABLE IF EXISTS clean.order CASCADE;
CREATE TABLE clean.order (
  order_no integer NOT NULL,
  order_date date NOT NULL,
  order_pic varchar(30) NOT NULL,
  customer_name varchar(50) NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp,
  updated_at timestamp NOT NULL DEFAULT current_timestamp,
  created_by varchar(58),
  updated_by varchar(58)
);

-- Set Table Comment
COMMENT ON TABLE clean.order IS '受注';

-- Set Column Comment
COMMENT ON COLUMN clean.order.order_no IS '受注番号';
COMMENT ON COLUMN clean.order.order_date IS '受注日付';
COMMENT ON COLUMN clean.order.order_pic IS '受注担当者名';
COMMENT ON COLUMN clean.order.customer_name IS '得意先名称';
COMMENT ON COLUMN clean.order.created_at IS '作成日時';
COMMENT ON COLUMN clean.order.updated_at IS '更新日時';
COMMENT ON COLUMN clean.order.created_by IS '作成者';
COMMENT ON COLUMN clean.order.updated_by IS '更新者';

-- Set PK Constraint
ALTER TABLE clean.order ADD PRIMARY KEY (
  order_no
);

-- Create 'set_update_at' Trigger
CREATE TRIGGER set_updated_at
  BEFORE UPDATE
  ON clean.order
  FOR EACH ROW
EXECUTE PROCEDURE set_updated_at();
