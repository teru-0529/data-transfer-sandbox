-- is_master_table=false

-- 4.受注明細(order_detail)

-- Create Table
DROP TABLE IF EXISTS clean.order_detail CASCADE;
CREATE TABLE clean.order_detail (
  order_no integer NOT NULL,
  order_detail_no integer NOT NULL,
  product_name varchar(30) NOT NULL,
  order_quantity integer NOT NULL check (order_quantity >= 0),
  shipping_flag boolean NOT NULL,
  cancel_flag boolean NOT NULL,
  selling_price integer NOT NULL check (selling_price >= 0),
  cost_price integer NOT NULL check (cost_price >= 0),
  created_at timestamp NOT NULL DEFAULT current_timestamp,
  updated_at timestamp NOT NULL DEFAULT current_timestamp,
  created_by varchar(58),
  updated_by varchar(58)
);

-- Set Table Comment
COMMENT ON TABLE clean.order_detail IS '受注明細';

-- Set Column Comment
COMMENT ON COLUMN clean.order_detail.order_no IS '受注番号';
COMMENT ON COLUMN clean.order_detail.order_detail_no IS '受注明細番号';
COMMENT ON COLUMN clean.order_detail.product_name IS '商品名';
COMMENT ON COLUMN clean.order_detail.order_quantity IS '受注数量';
COMMENT ON COLUMN clean.order_detail.shipping_flag IS '出荷済フラグ';
COMMENT ON COLUMN clean.order_detail.cancel_flag IS 'キャンセルフラグ';
COMMENT ON COLUMN clean.order_detail.selling_price IS '販売単価';
COMMENT ON COLUMN clean.order_detail.cost_price IS '商品原価';
COMMENT ON COLUMN clean.order_detail.created_at IS '作成日時';
COMMENT ON COLUMN clean.order_detail.updated_at IS '更新日時';
COMMENT ON COLUMN clean.order_detail.created_by IS '作成者';
COMMENT ON COLUMN clean.order_detail.updated_by IS '更新者';

-- Set PK Constraint
ALTER TABLE clean.order_detail ADD PRIMARY KEY (
  order_no,
  order_detail_no
);

-- Create 'set_update_at' Trigger
CREATE TRIGGER set_updated_at
  BEFORE UPDATE
  ON clean.order_detail
  FOR EACH ROW
EXECUTE PROCEDURE set_updated_at();
