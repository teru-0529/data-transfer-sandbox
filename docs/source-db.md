# テーブル定義

----------

## #1 受注(orders)

### Fields

| # | 名称 | データ型 | NOT NULL | 初期値 | 制約 |
| -- | -- | -- | -- | -- | -- |
| 1 | 受注番号(order_no) | integer | true |  |  |
| 2 | 受注日(order_date) | varchar(8) | true |  |  |
| 3 | 受注担当者名(order_pic) | varchar(30) | false |  | (LENGTH(order_pic) >= 3) |
| 4 | 得意先名称(customer_name) | varchar(50) | true |  |  |

### Constraints

#### Primary Key

* 受注番号(order_no)

----------

## #2 受注明細(order_details)

### Fields

| # | 名称 | データ型 | NOT NULL | 初期値 | 制約 |
| -- | -- | -- | -- | -- | -- |
| 1 | 受注番号(order_no) | integer | true |  |  |
| 2 | 受注明細番号(order_detail_no) | integer | true |  |  |
| 3 | 商品名(product_name) | varchar(30) | false |  |  |
| 4 | 受注数(receiving_quantity) | integer | false |  | (0 <= receiving_quantity AND receiving_quantity <= 1000) |
| 5 | 出荷済フラグ(shipping_flag) | boolean | false |  |  |
| 6 | キャンセルフラグ(canceled_flag) | boolean | false |  |  |
| 7 | 販売単価(sellling_price) | integer | true |  | (0 <= sellling_price AND sellling_price <= 9999999) |
| 8 | 商品原価(cost_price) | integer | true |  | (0 <= cost_price AND cost_price <= 9999999) |

### Constraints

#### Primary Key

* 受注番号(order_no)
* 受注明細番号(order_detail_no)

----------

## #3 商品(products)

### Fields

| # | 名称 | データ型 | NOT NULL | 初期値 | 制約 |
| -- | -- | -- | -- | -- | -- |
| 1 | 商品名(product_name) | varchar(30) | true |  |  |
| 2 | 商品原価(cost_price) | integer | true |  | (0 <= cost_price AND cost_price <= 9999999) |

### Constraints

#### Primary Key

* 商品名(product_name)

----------

## #4 担当者(operaters)

### Fields

| # | 名称 | データ型 | NOT NULL | 初期値 | 制約 |
| -- | -- | -- | -- | -- | -- |
| 1 | 担当者ID(operator_id) | varchar(5) | true |  |  |
| 2 | 担当者名(operator_name) | varchar(30) | true |  | (LENGTH(operator_name) >= 3) |

### Constraints

#### Primary Key

* 担当者ID(operator_id)

----------
