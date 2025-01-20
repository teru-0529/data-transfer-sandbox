-- is_master_table=false

-- 3.受注(orders)

-- Create Table
DROP TABLE IF EXISTS clean.orders CASCADE;
CREATE TABLE clean.orders (
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
COMMENT ON TABLE clean.orders IS '受注';

-- Set Column Comment
COMMENT ON COLUMN clean.orders.order_no IS '受注番号';
COMMENT ON COLUMN clean.orders.order_date IS '受注日付';
COMMENT ON COLUMN clean.orders.order_pic IS '受注担当者名';
COMMENT ON COLUMN clean.orders.customer_name IS '得意先名称';
COMMENT ON COLUMN clean.orders.created_at IS '作成日時';
COMMENT ON COLUMN clean.orders.updated_at IS '更新日時';
COMMENT ON COLUMN clean.orders.created_by IS '作成者';
COMMENT ON COLUMN clean.orders.updated_by IS '更新者';

-- Set PK Constraint
ALTER TABLE clean.orders ADD PRIMARY KEY (
  order_no
);

-- Create 'set_update_at' Trigger
CREATE TRIGGER set_updated_at
  BEFORE UPDATE
  ON clean.orders
  FOR EACH ROW
EXECUTE PROCEDURE set_updated_at();
