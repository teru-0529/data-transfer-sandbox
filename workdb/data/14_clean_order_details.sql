-- create view

-- 4.受注明細(order_details)

CREATE OR REPLACE VIEW clean.w_order_details AS

  -- 集約
  WITH oda AS (
    SELECT
      w_order_no,
      order_no,
      STRING_AGG(order_detail_no::TEXT, ',') AS aggregated_details, --カンマ区切りでw_order_noを結合
      COUNT(*) AS detail_count, -- 集約件数を計算
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
    ROW_NUMBER() OVER (PARTITION BY od.w_order_no, od.product_name ORDER BY od.order_detail_no DESC) = 1 AS register,
    oda.w_order_no,
    oda.order_no,
    oda.aggregated_details,
    oda.detail_count,
    pr.w_product_id,
    oda.product_name,
    oda.receiving_quantity,
    oda.w_shipping_quantity,
    oda.w_cancel_quantity,
    oda.w_remaining_quantity,
    oda.selling_price,
    oda.cost_price,
    oda.is_shipped,
    oda.is_remaining
  FROM
    clean.order_details od
  LEFT OUTER JOIN
    oda
  ON
    od.w_order_no = oda.w_order_no AND od.product_name = oda.product_name
  INNER JOIN
    clean.products pr
  ON
    oda.product_name = pr.product_name
  ORDER BY oda.w_order_no, pr.w_product_id;
