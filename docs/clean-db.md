# クレンジング仕様

----------

## #1 担当者(operators)

<details><summary>(open)</summary>

### <u>●Table layout</u>

| # | 名称 | データ型 | NOT NULL | 初期値 | 制約 |
| -- | -- | -- | -- | -- | -- |
| 1 | 担当者ID(operator_id) | varchar(5) | true |  | (LENGTH(operator_id) = 5) |
| 2 | 担当者名(operator_name) | varchar(30) | true |  |  |

### <u>●Constraints</u>

| # | 状況 | 対応方針 | 承認 | BacklogId |
| -- | -- | -- | :--: | -- |
| #1-01 | 担当者名が一意ではない | ⛔REMOVE | 〇 | xxxxx |
| #1-02 | 担当者IDが5桁に満たない | ⚠MODIFY<br>末尾に`X`を追加しクレンジング |  | xxxxx |

</details>

----------

## #2 商品(products)

<details><summary>(open)</summary>

### <u>●Table layout</u>

| # | 名称 | データ型 | NOT NULL | 初期値 | 制約 |
| -- | -- | -- | -- | -- | -- |
| 1 | 商品名(product_name) | varchar(30) | true |  |  |
| 2 | 商品原価(cost_price) | integer | true |  | (cost_price >= 0) |

### <u>●Constraints</u>

| # | 状況 | 対応方針 | 承認 | BacklogId |
| -- | -- | -- | :--: | -- |
| #2-01 | 商品原価がマイナス | ⚠MODIFY<br>固定値(0)に変換しクレンジング |  | xxxxx |

</details>

----------

## #3 受注(orders)

<details><summary>(open)</summary>

### <u>●Table layout</u>

| # | 名称 | データ型 | NOT NULL | 初期値 | 制約 |
| -- | -- | -- | -- | -- | -- |
| 1 | 受注番号(order_no) | integer | true |  |  |
| 2 | 受注日付(order_date) | date | true |  |  |
| 3 | 受注担当者名(order_pic) | varchar(30) | true |  |  |
| 4 | 得意先名称(customer_name) | varchar(50) | true |  |  |

### <u>●Constraints</u>

| # | 状況 | 対応方針 | 承認 | BacklogId |
| -- | -- | -- | :--: | -- |
| #3-01 | 受注日付が日付型ではない | ⚠MODIFY<br>固定値(20250101)に変換しクレンジング | 〇 | xxxxx |
| #3-02 | 受注担当者名が「担当者」に存在しない | ⚠MODIFY<br>固定値(N/A)※に変換しクレンジング | 〇 | xxxxx |

※担当者(Z9999、N/A)を「担当者」に固定で登録する。

</details>

----------

## #4 受注明細(order_details)

<details><summary>(open)</summary>

### <u>●Table layout</u>

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

### <u>●Constraints</u>

| # | 状況 | 対応方針 | 承認 | BacklogId |
| -- | -- | -- | :--: | -- |
| #4-01 | 出荷済フラグ/キャンセルフラグが両方ともTrue | ⛔REMOVE | 〇 | xxxxx |
| #4-02 | 受注番号が「受注」に存在しない | ⛔REMOVE | 〇 | xxxxx |
| #4-03 | 商品名が「商品」に存在しない | ⛔REMOVE |  |  |

</details>

----------
