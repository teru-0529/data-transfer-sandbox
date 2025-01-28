-- create view

-- 3.受注(orders)

CREATE OR REPLACE VIEW clean.w_orders AS

  -- 集約
  WITH oda AS (
    SELECT
      ROW_NUMBER() OVER (PARTITION BY order_no ORDER BY w_order_no ASC) > 1 AS logging, -- ログの対象(受注を分割した場合)
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
  ),

  oc AS (
    SELECT
      order_no,
      COUNT(*) AS order_count -- order_noの出現数を計算
    FROM
      oda
    GROUP BY
      order_no
  )

  SELECT
    CASE
      WHEN oda.order_no IS NOT NULL THEN true
      ELSE false
    END AS register, -- oda(受注明細の集約データ)がある場合登録対象
    COALESCE(oda.logging, true) AS logging, -- ログの対象(odaがない場合は強制Log)
    oda.w_order_no,
    o.order_no,
    COALESCE(oc.order_count - 1, -1) AS change_count, -- 入力情報からの変動数
    o.order_date,
    op.operator_id,
    o.order_pic,
    o.customer_name,
    oda.w_total_order_price,
    oda.w_remaining_order_price,
    oda.is_shipped,
    oda.is_remaining
  FROM
    clean.orders o
  LEFT OUTER JOIN
    oda
  ON
    o.order_no = oda.order_no
  INNER JOIN
    clean.operators op
  ON
    o.order_pic = op.operator_name
  LEFT OUTER JOIN
    oc
  ON
    o.order_no = oc.order_no
  ORDER BY oda.w_order_no;
