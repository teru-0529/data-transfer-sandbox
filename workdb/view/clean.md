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
