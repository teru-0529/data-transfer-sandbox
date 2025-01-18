-- operation_afert_create_tables

-- 1.受注番号(gen_order_no)

-- Set FK Constraint
ALTER TABLE work.gen_order_no DROP CONSTRAINT IF EXISTS gen_order_no_foreignKey_1;
ALTER TABLE work.gen_order_no ADD CONSTRAINT gen_order_no_foreignKey_1 FOREIGN KEY (
  product_idw
) REFERENCES work.gen_product_id (
  product_idw
) ON DELETE CASCADE ON UPDATE CASCADE;
