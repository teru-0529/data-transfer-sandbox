-- ext insert functions

-- 4.受注明細(order_details)

-- 導出属性:出荷済数(WORK)
-- 導出属性:キャンセル数(WORK)
-- 導出属性:受注残数(WORK)
-- 導出属性:受注金額(WORK)
-- 導出属性:受注残額(WORK)

-- Create Function
CREATE OR REPLACE FUNCTION clean.order_details_pre_process() RETURNS TRIGGER AS $$
BEGIN
  IF NEW.shipping_flag THEN
    NEW.w_shipping_quantity:=NEW.receiving_quantity;
  ELSEIF NEW.cancel_flag THEN
    NEW.w_cancel_quantity:=NEW.receiving_quantity;
  ELSE
    NEW.w_remaining_quantity:=NEW.receiving_quantity;
  END IF;

  NEW.w_total_order_price:=NEW.receiving_quantity * NEW.selling_price;
  NEW.w_remaining_order_price:=NEW.w_remaining_quantity * NEW.selling_price;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create Trigger
CREATE TRIGGER pre_process
  BEFORE INSERT
  ON clean.order_details
  FOR EACH ROW
EXECUTE PROCEDURE clean.order_details_pre_process();
