-- operation_afert_create_tables

-- 3.受注(order)

-- Set FK Constraint
ALTER TABLE clean.order DROP CONSTRAINT IF EXISTS order_foreignKey_1;
ALTER TABLE clean.order ADD CONSTRAINT order_foreignKey_1 FOREIGN KEY (
  order_pic
) REFERENCES clean.operators (
  operator_name
);
