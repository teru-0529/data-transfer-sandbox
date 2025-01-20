/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cleansing

import (
	"fmt"
	"log"
	"time"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/clean"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/legacy"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// STRUCT:
type ProductsClensing struct {
	conns   infra.DbConnection
	Result  Result
	Details []ProductPiece
}

type ProductPiece struct {
	ProductName string
	status      Status
	approved    bool
	message     string
}

// FUNCTION:
func NewProducts(conns infra.DbConnection) ProductsClensing {
	s := time.Now()

	cs := ProductsClensing{conns: conns, Result: Result{
		TableNameJp: "商品",
		TableNameEn: "products",
	}}

	// PROCESS: 入力データ量
	cs.setEntryCount()

	// PROCESS: クレンジング先のtruncate
	_, err := queries.Raw(cs.Result.truncateSql()).ExecContext(ctx, cs.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}

	// PROCESS: 1行ごと処理
	cs.iterate()

	log.Printf("cleansing completed [%s] … %s\n", cs.Result.TableNameEn, time.Since(s))
	return cs
}

// FUNCTION: 入力データ量
func (cs *ProductsClensing) setEntryCount() {
	num, err := legacy.Products().Count(ctx, cs.conns.LegacyDB)
	if err != nil {
		log.Fatalln(err)
	}
	cs.Result.EntryCount = int(num)
}

// FUNCTION: データ繰り返し取得(1000件単位で分割)
func (cs *ProductsClensing) iterate() {

	// PROCESS: 1000件単位でのSQL実行に分割する
	for section := 0; section < cs.Result.sectionCount(); section++ {
		records, err := legacy.Products(qm.Limit(LIMIT), qm.Offset(section*LIMIT)).All(ctx, cs.conns.LegacyDB)
		if err != nil {
			log.Fatalln(err)
		}

		for _, record := range records {
			// PROCESS: レコード毎のチェック
			piece := cs.checkAndClensing(record)
			// PROCESS: レコード毎のクレンジング後データ登録
			cs.saveData(record, piece)
		}
	}
}

// FUNCTION: レコード毎のチェック
func (cs *ProductsClensing) checkAndClensing(record *legacy.Product) ProductPiece {
	costPrice := record.CostPrice

	// PROCESS: cost_priceが負の数字の場合は、0にクレンジングする。
	if costPrice < 0 {
		record.CostPrice = 0
		return ProductPiece{
			ProductName: record.ProductName,
			status:      MODIFY,
			approved:    false,
			message:     fmt.Sprintf("cost_price(商品原価) が負の数。`%d` → `0`(固定値) にクレンジング。", costPrice),
		}
	}

	return ProductPiece{status: NO_CHANGE}
}

// FUNCTION: レコード毎のクレンジング後データ登録
func (cs *ProductsClensing) saveData(record *legacy.Product, piece ProductPiece) {
	// PROCESS: REMOVEDの場合はDBに登録しない
	if piece.status == REMOVE {
		cs.setResult(piece)
		return
	}

	// PROCESS: データ登録
	rec := clean.Product{
		ProductName: record.ProductName,
		CostPrice:   record.CostPrice,
		CreatedBy:   OPERATION_USER,
		UpdatedBy:   OPERATION_USER,
	}
	err := rec.Insert(ctx, cs.conns.WorkDB, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納、未承認扱い)
	if err != nil {
		p := ProductPiece{
			ProductName: record.ProductName,
			status:      REMOVE,
			approved:    false,
			message:     fmt.Sprintf("%v", err),
		}
		cs.setResult(p)
		return
	}
	cs.setResult(piece)
}

// FUNCTION: クレンジング結果の登録
func (cs *ProductsClensing) setResult(piece ProductPiece) {
	switch piece.status {
	case NO_CHANGE:
		cs.Result.UnchangeCount++
	case MODIFY:
		cs.Result.ModifyCount++
		cs.Details = append(cs.Details, piece)
	case REMOVE:
		cs.Result.RemoveCount++
		cs.Details = append(cs.Details, piece)
	}
}

// FUNCTION: 詳細情報の出力
func (cs *ProductsClensing) ShowDetails() string {
	if len(cs.Details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", cs.Result.TableName())

	msg += "  | # | product_name | … | RESULT | APPROVED | MESSAGE |\n"
	msg += "  |--:|---|---|:-:|:-:|---|\n"
	for i, piece := range cs.Details {
		msg += fmt.Sprintf("  | %d | %s | … | %s | %s | %s |\n",
			i+1,
			piece.ProductName,
			piece.status,
			approveStr(piece.approved),
			piece.message,
		)
	}

	return msg
}
