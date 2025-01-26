-- create view

-- 3.受注(orders)

CREATE OR REPLACE VIEW clean.w_orders AS

  -- 集約
  WITH od AS (
    SELECT
      w_order_no,
      order_no,
      SUM(w_total_order_price) AS w_total_order_price,
      SUM(w_remaining_order_price) AS w_remaining_order_price,
      SUM(w_shipping_quantity) > 0 AS is_shipped,
      SUM(w_remaining_quantity) = 0 AS is_remaining
    FROM
      clean.order_details
    GROUP BY
      w_order_no, order_no
  )

  SELECT
    od.w_order_no,
    od.order_no,
    o.order_date,
    op.operator_id,
    o.order_pic,
    o.customer_name,
    od.w_total_order_price,
    od.w_remaining_order_price,
    od.is_shipped,
    od.is_remaining
  FROM
    od
  INNER JOIN
    clean.orders o
  ON
    od.order_no = o.order_no
  INNER JOIN
    clean.operators op
  ON
    o.order_pic = op.operator_name
  ORDER BY od.w_order_no;
