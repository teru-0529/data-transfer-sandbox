/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cleansing

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/clean"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/legacy"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// STRUCT: 詳細メッセージ
type ProductMsg struct {
	ProductName string
	bp          *Piece
}

func NewProductMsg(productName string) *ProductMsg {
	return &ProductMsg{
		ProductName: productName,
		bp:          NewPiece(),
	}
}

// STRUCT: レコード
type ProductRecord struct {
	record legacy.Product
	msg    *ProductMsg
	setTo  *[]ProductMsg
}

// FUNCTION: 更新
func (r *ProductRecord) checkAndPersist(ctx infra.AppCtx, db *sql.DB, refData *RefData) Piece {

	// PROCESS:TODO: check #2-01:
	r.checkCostPrice()

	// PROCESS: REMOVE判定時は登録なし
	if !r.msg.bp.isRemove() {
		r.persiste(ctx, db)
	}

	// PROCESS: REMOVE/MODIFY判定時は詳細情報の出力あり
	if r.msg.bp.isWarn() {
		*r.setTo = append(*r.setTo, *r.msg)
	}

	// PROCESS:INFO: 正常登録時に[商品名]登録
	if !r.msg.bp.isRemove() {
		refData.ProductNameSet[r.record.ProductName] = struct{}{}
	}

	return *r.msg.bp
}

// FUNCTION: データ登録
func (r *ProductRecord) persiste(ctx infra.AppCtx, db *sql.DB) {
	// PROCESS: データ登録
	rec := clean.Product{
		ProductName: r.record.ProductName,
		CostPrice:   r.record.CostPrice,
		CreatedBy:   ctx.OperationUser,
		UpdatedBy:   ctx.OperationUser,
		// INFO: w_product_id はtrigger function
	}
	err := rec.Insert(ctx.Ctx, db, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
	if err != nil {
		r.msg.bp.dbError(err)
	}
}

// FUNCTION: #2-01(MODIFY): cost_priceが負の数字の場合は、0にクレンジングする。
func (r *ProductRecord) checkCostPrice() {
	const ID = "#2-01"
	costPrice := r.record.CostPrice
	if costPrice < 0 {
		r.record.CostPrice = 0
		r.msg.bp.approveStay() //TODO: 承認確認中
		r.msg.bp.modified().addMessage(
			fmt.Sprintf("cost_price(商品原価) が負の数です`%d`。<br>【クレンジング】`0`に変換", costPrice), "#2-01")
	}
}

// STRUCT: コマンド
type ProductsCmd struct {
	details []ProductMsg
}

// FUNCTION: New
func NewProductsCmd() *ProductsCmd {
	return &ProductsCmd{}
}

// FUNCTION: テーブル名設定
func (cmd *ProductsCmd) getTableInfo() TableInfo {
	return TableInfo{
		tableJp: "商品",
		tableEn: legacy.TableNames.Products,
	}
}

// FUNCTION: 入力データ量
func (cmd *ProductsCmd) entryCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := legacy.Products().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	return int(num)
}

// FUNCTION: 処理対象レコードのフェッチ
func (cmd *ProductsCmd) fetchRecords(ctx infra.AppCtx, con *sql.DB, qmArray []qm.QueryMod) []Record {
	//INFO: 商品名:昇順
	qmArray = append(qmArray, qm.OrderBy("product_name ASC"))
	records, err := legacy.Products(qmArray...).All(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}

	results := make([]Record, len(records))
	for i, record := range records {
		results[i] = Record{rec: &ProductRecord{
			record: *record,
			msg:    NewProductMsg(record.ProductName),
			setTo:  &cmd.details,
		}}
	}
	return results
}

// FUNCTION: 追加データ登録
func (r *ProductsCmd) extInsert(ctx infra.AppCtx, db *sql.DB, refData *RefData) {}

// FUNCTION: 詳細メッセージの出力
func (cmd *ProductsCmd) showDetails(ctx infra.AppCtx, tableName string) string {
	if len(cmd.details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", tableName)
	msg += "  | # | product_name | … | RESULT | APPROVED | MESSAGE |\n"
	msg += "  |--:|---|---|:-:|:-:|---|\n"
	for i, piece := range cmd.details {
		msg += fmt.Sprintf("  | %d | %s | … | %s | %s | %s |\n",
			i+1,
			piece.ProductName,
			piece.bp.status,
			piece.bp.approve,
			piece.bp.msg,
		)
	}

	return msg
}
