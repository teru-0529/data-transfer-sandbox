-- ext check constructions

-- 4.受注明細(order_details)

-- Add Check Constraint
--  属性相関チェック制約(出荷済フラグ/キャンセルフラグ)
ALTER TABLE clean.order_details ADD CONSTRAINT order_details_shipping_and_cancel_flag_check CHECK (
  NOT(shipping_flag AND cancel_flag)
);
