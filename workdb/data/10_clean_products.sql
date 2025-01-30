-- is_master_table=false

-- 2.商品(products)

-- Create Table
DROP TABLE IF EXISTS clean.products CASCADE;
CREATE TABLE clean.products (
  product_name varchar(30) NOT NULL,
  cost_price integer NOT NULL check (cost_price >= 0),
  w_product_id varchar(5) NOT NULL check (w_product_id ~* '^P[0-9]{4}$'),
  created_at timestamp NOT NULL DEFAULT current_timestamp,
  updated_at timestamp NOT NULL DEFAULT current_timestamp,
  created_by varchar(58),
  updated_by varchar(58)
);

-- Set Table Comment
COMMENT ON TABLE clean.products IS '商品';

-- Set Column Comment
COMMENT ON COLUMN clean.products.product_name IS '商品名';
COMMENT ON COLUMN clean.products.cost_price IS '商品原価';
COMMENT ON COLUMN clean.products.w_product_id IS '商品ID(WORK)';
COMMENT ON COLUMN clean.products.created_at IS '作成日時';
COMMENT ON COLUMN clean.products.updated_at IS '更新日時';
COMMENT ON COLUMN clean.products.created_by IS '作成者';
COMMENT ON COLUMN clean.products.updated_by IS '更新者';

-- Set PK Constraint
ALTER TABLE clean.products ADD PRIMARY KEY (
  product_name
);

-- Set Unique Constraint
ALTER TABLE clean.products ADD CONSTRAINT products_unique_1 UNIQUE (
  w_product_id
);

-- Create 'set_update_at' Trigger
CREATE TRIGGER set_updated_at
  BEFORE UPDATE
  ON clean.products
  FOR EACH ROW
EXECUTE PROCEDURE set_updated_at();
