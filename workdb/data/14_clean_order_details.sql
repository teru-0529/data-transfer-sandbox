-- create view

-- 4.受注明細(order_details)

CREATE OR REPLACE VIEW clean.w_order_details AS

  -- 集約
  WITH od AS (
    SELECT
      w_order_no,
      order_no,
      product_name,
      SUM(receiving_quantity) AS receiving_quantity,
      SUM(w_shipping_quantity) AS w_shipping_quantity,
      SUM(w_cancel_quantity) AS w_cancel_quantity,
      SUM(w_remaining_quantity) AS w_remaining_quantity,
      selling_price,
      cost_price,
      SUM(w_shipping_quantity) > 0 AS is_shipped,
      SUM(w_remaining_quantity) = 0 AS is_remaining
    FROM
      clean.order_details
    GROUP BY
      w_order_no, order_no, product_name, selling_price, cost_price
  )

  SELECT
    od.w_order_no,
    od.order_no,
    pr.w_product_id,
    od.product_name,
    od.receiving_quantity,
    od.w_shipping_quantity,
    od.w_cancel_quantity,
    od.w_remaining_quantity,
    od.selling_price,
    od.cost_price,
    od.is_shipped,
    od.is_remaining
  FROM
    od
  INNER JOIN
    clean.products pr
  ON
    od.product_name = pr.product_name
  ORDER BY od.w_order_no, pr.w_product_id;
