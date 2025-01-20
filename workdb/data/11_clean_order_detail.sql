-- operation_afert_create_tables

-- 4.受注明細(order_detail)

-- Set FK Constraint
ALTER TABLE clean.order_detail DROP CONSTRAINT IF EXISTS order_detail_foreignKey_1;
ALTER TABLE clean.order_detail ADD CONSTRAINT order_detail_foreignKey_1 FOREIGN KEY (
  order_no
) REFERENCES clean.order (
  order_no
) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE clean.order_detail DROP CONSTRAINT IF EXISTS order_detail_foreignKey_2;
ALTER TABLE clean.order_detail ADD CONSTRAINT order_detail_foreignKey_2 FOREIGN KEY (
  product_name
) REFERENCES clean.products (
  product_name
);
