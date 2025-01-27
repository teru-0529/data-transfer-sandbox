/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package transfer

import (
	"fmt"
	"log"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/spec/product/orders"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/clean"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// STRUCT:
type ProductPiece struct {
	ProductId string
	bp        *Piece
}

// STRUCT:
type ProductTransfer struct {
	conns   infra.DbConnection
	Result  Result
	Details []*ProductPiece
}

// FUNCTION:
func NewProducts(conns infra.DbConnection) ProductTransfer {
	s := time.Now()

	// INFO: 固定値設定
	ts := ProductTransfer{
		conns:  conns,
		Result: Result{Schema: "orders", TableNameJp: "商品", TableNameEn: "products"},
	}
	log.Printf("[%s] table transfer ...", ts.Result.TableNameEn)

	// PROCESS: 入力データ量
	ts.setEntryCount()

	// PROCESS: 移行先のtruncate
	_, err := queries.Raw(ts.Result.truncateSql()).ExecContext(ctx, ts.conns.ProductDB)
	if err != nil {
		log.Fatalln(err)
	}

	// PROCESS: 1行ごと処理
	ts.iterate()

	duration := time.Since(s).Seconds()
	ts.Result.duration = duration
	log.Printf("cleansing completed … %3.2fs\n", duration)
	return ts
}

// FUNCTION: 入力データ量
func (ts *ProductTransfer) setEntryCount() {
	// INFO: Workテーブル名
	num, err := clean.Products().Count(ctx, ts.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}
	ts.Result.EntryCount = int(num)
}

// FUNCTION: データ繰り返し取得(1000件単位で分割)
func (ts *ProductTransfer) iterate() {
	bar := pb.Default.Start(ts.Result.EntryCount)
	bar.SetMaxWidth(80)

	// PROCESS: 1000件単位でのSQL実行に分割する
	for section := 0; section < ts.Result.sectionCount(); section++ {
		// INFO: Workテーブル名
		records, err := clean.Products(qm.Limit(LIMIT), qm.Offset(section*LIMIT)).All(ctx, ts.conns.WorkDB)
		if err != nil {
			log.Fatalln(err)
		}

		for _, record := range records {
			// PROCESS: レコード毎のデータ登録
			ts.saveData(record)

			bar.Increment()
		}
	}
	bar.Finish()
}

// FUNCTION: レコード毎のデータ登録
func (ts *ProductTransfer) saveData(record *clean.Product) {

	// PROCESS: データ登録
	// INFO: productテーブル
	rec := orders.Product{
		ProductID:     record.WProductID,
		ProductName:   record.ProductName,
		CostPrice:     record.CostPrice,
		ProductPic:    "Z9999",
		ProductStatus: orders.ProductStatusON_SALE,
		CreatedBy:     OPERATION_USER,
		UpdatedBy:     OPERATION_USER,
	}
	err := rec.Insert(ctx, ts.conns.ProductDB, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
	if err != nil {
		// INFO: piece
		piece := ProductPiece{
			ProductId: record.WProductID,
			bp:        removedPiece(fmt.Sprintf("%v", err)),
		}
		ts.Result.setResult(piece.bp)
		ts.Details = append(ts.Details, &piece)
	}
}

// FUNCTION: 詳細情報の出力
func (ts *ProductTransfer) ShowDetails() string {
	if len(ts.Details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", ts.Result.TableName())

	msg += "  | # | product_id | … | RESULT | CHANGE | MESSAGE |\n"
	msg += "  |--:|---|---|:-:|:-:|---|\n"
	for i, piece := range ts.Details {
		msg += fmt.Sprintf("  | %d | %s | … | %s | %+d | %s |\n",
			i+1,
			piece.ProductId,
			piece.bp.status,
			piece.bp.count,
			piece.bp.msg,
		)
	}

	return msg
}
