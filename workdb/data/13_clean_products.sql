-- ext insert functions

-- 2.商品(products)

-- シーケンス
DROP SEQUENCE IF EXISTS clean.product_id_seed;
CREATE SEQUENCE clean.product_id_seed START 1;

-- 導出属性:商品ID(WORK)

-- Create Function
CREATE OR REPLACE FUNCTION clean.products_pre_process() RETURNS TRIGGER AS $$
BEGIN
  NEW.w_product_id:='P'||to_char(nextval('clean.product_id_seed'),'FM0000');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create Trigger
CREATE TRIGGER pre_process
  BEFORE INSERT
  ON clean.products
  FOR EACH ROW
EXECUTE PROCEDURE clean.products_pre_process();
