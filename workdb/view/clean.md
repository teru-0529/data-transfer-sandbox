# テーブル定義

----------

## #1 担当者(operators)

### Fields

| # | 名称 | データ型 | NOT NULL | 初期値 | 制約 |
| -- | -- | -- | -- | -- | -- |
| 1 | 担当者ID(operator_id) | varchar(5) | true |  | (LENGTH(operator_id) = 5) |
| 2 | 担当者名(operator_name) | varchar(30) | true |  |  |

### Constraints

#### Primary Key

* 担当者ID(operator_id)

#### Uniques

#### operators_unique_1

* 担当者名(operator_name)

----------

## #2 商品(products)

### Fields

| # | 名称 | データ型 | NOT NULL | 初期値 | 制約 |
| -- | -- | -- | -- | -- | -- |
| 1 | 商品名(product_name) | varchar(30) | true |  |  |
| 2 | 商品原価(cost_price) | integer | true |  | (cost_price >= 0) |

### Constraints

#### Primary Key

* 商品名(product_name)

----------

## #3 受注(orders)

### Fields

| # | 名称 | データ型 | NOT NULL | 初期値 | 制約 |
| -- | -- | -- | -- | -- | -- |
| 1 | 受注番号(order_no) | integer | true |  |  |
| 2 | 受注日付(order_date) | date | true |  |  |
| 3 | 受注担当者名(order_pic) | varchar(30) | true |  |  |
| 4 | 得意先名称(customer_name) | varchar(50) | true |  |  |

### Constraints

#### Primary Key

* 受注番号(order_no)

#### Foreign Keys

#### orders_foreignKey_1

* 参照先テーブル : 担当者(operators)
* 削除時オプション : RESTRICT(デフォルト値)
* 更新時オプション : RESTRICT(デフォルト値)

| # | フィールド | 参照先フィールド |
| -- | -- | -- |
| 1 | 受注担当者名(order_pic) | 担当者名(operator_name) |

----------

## #4 受注明細(order_details)

### Fields

| # | 名称 | データ型 | NOT NULL | 初期値 | 制約 |
| -- | -- | -- | -- | -- | -- |
| 1 | 受注番号(order_no) | integer | true |  |  |
| 2 | 受注明細番号(order_detail_no) | integer | true |  |  |
| 3 | 商品名(product_name) | varchar(30) | true |  |  |
| 4 | 受注数量(receiving_quantity) | integer | true |  | (receiving_quantity >= 0) |
| 5 | 出荷済フラグ(shipping_flag) | boolean | true |  |  |
| 6 | キャンセルフラグ(cancel_flag) | boolean | true |  |  |
| 7 | 販売単価(selling_price) | integer | true |  | (selling_price >= 0) |
| 8 | 商品原価(cost_price) | integer | true |  | (cost_price >= 0) |

### Constraints

#### Primary Key

* 受注番号(order_no)
* 受注明細番号(order_detail_no)

#### Foreign Keys

#### order_details_foreignKey_1

* 参照先テーブル : 受注(orders)
* 削除時オプション : CASCADE
* 更新時オプション : CASCADE

| # | フィールド | 参照先フィールド |
| -- | -- | -- |
| 1 | 受注番号(order_no) | 受注番号(order_no) |

#### order_details_foreignKey_2

* 参照先テーブル : 商品(products)
* 削除時オプション : RESTRICT(デフォルト値)
* 更新時オプション : RESTRICT(デフォルト値)

| # | フィールド | 参照先フィールド |
| -- | -- | -- |
| 1 | 商品名(product_name) | 商品名(product_name) |

----------
