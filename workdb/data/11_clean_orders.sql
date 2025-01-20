-- operation_afert_create_tables

-- 3.受注(orders)

-- Set FK Constraint
ALTER TABLE clean.orders DROP CONSTRAINT IF EXISTS orders_foreignKey_1;
ALTER TABLE clean.orders ADD CONSTRAINT orders_foreignKey_1 FOREIGN KEY (
  order_pic
) REFERENCES clean.operators (
  operator_name
);
