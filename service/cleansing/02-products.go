/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cleansing

import (
	"fmt"
	"log"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/clean"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/legacy"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// STRUCT: FIXME:
type ProductPiece struct {
	ProductName string
	bp          *Piece
	status      Status
	approved    bool
	message     string
}

// FUNCTION: ステータスの変更 FIXME:
func (p *ProductPiece) setStatus(status Status, approved bool) {
	p.status = judgeStatus(p.status, status)
	p.approved = p.approved && approved
}

// FUNCTION: メッセージの追加 FIXME:
func (p *ProductPiece) addMessage(msg string, id string) {
	p.message = genMessage(p.message, msg, id)
}

// STRUCT:
type ProductsClensing struct {
	conns   infra.DbConnection
	refData *RefData
	Result  Result
	Details []*ProductPiece
}

// FUNCTION:
func NewProducts(conns infra.DbConnection, refData *RefData) ProductsClensing {
	s := time.Now()

	// INFO: 固定値設定
	cs := ProductsClensing{
		conns:   conns,
		refData: refData,
		Result:  Result{TableNameJp: "商品", TableNameEn: "products"},
	}
	log.Printf("[%s] table cleansing ...", cs.Result.TableNameEn)

	// PROCESS: 入力データ量
	cs.setEntryCount()

	// PROCESS: クレンジング先のtruncate
	_, err := queries.Raw(cs.Result.truncateSql()).ExecContext(ctx, cs.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}

	// INFO: クレンジング先のシーケンスリセット(EXT)
	sql := "SELECT SETVAL ('clean.product_id_seed', 1, false);"
	_, err = queries.Raw(sql).ExecContext(ctx, cs.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}
	// PROCESS: 1行ごと処理
	cs.iterate()

	duration := time.Since(s).Seconds()
	cs.Result.duration = duration
	log.Printf("cleansing completed … %3.2fs\n", duration)
	return cs
}

// FUNCTION: 入力データ量
func (cs *ProductsClensing) setEntryCount() {
	// INFO: Legacyテーブル名
	num, err := legacy.Products().Count(ctx, cs.conns.LegacyDB)
	if err != nil {
		log.Fatalln(err)
	}
	cs.Result.EntryCount = int(num)
}

// FUNCTION: データ繰り返し取得(1000件単位で分割)
func (cs *ProductsClensing) iterate() {
	bar := pb.Default.Start(cs.Result.EntryCount)
	bar.SetMaxWidth(80)

	// PROCESS: 1000件単位でのSQL実行に分割する
	for section := 0; section < cs.Result.sectionCount(); section++ {
		// INFO: Legacyテーブル名
		// INFO: 商品名:昇順
		records, err := legacy.Products(
			qm.Limit(LIMIT), qm.Offset(section*LIMIT), qm.OrderBy("product_name ASC")).All(ctx, cs.conns.LegacyDB)
		if err != nil {
			log.Fatalln(err)
		}

		for _, record := range records {
			// PROCESS: レコード毎のチェック
			piece := cs.checkAndClensing(record)
			// PROCESS: レコード毎のクレンジング後データ登録
			cs.saveData(record, piece)

			// FIXME:
			// PROCESS: 処理結果の登録
			cs.Result.setResult(piece.bp)
			if piece.bp.isWarn() {
				cs.Details = append(cs.Details, piece)
			}

			bar.Increment()
		}
	}
	bar.Finish()
}

// FUNCTION: レコード毎のチェック
func (cs *ProductsClensing) checkAndClensing(record *legacy.Product) *ProductPiece {
	// INFO: piece FIXME:
	piece := ProductPiece{
		ProductName: record.ProductName,
		bp:          NewPiece(),
		status:      NO_CHANGE,
		approved:    true,
	}

	// PROCESS: #2-01: cost_priceが負の数字の場合は、0にクレンジングする。
	costPrice := record.CostPrice

	if costPrice < 0 {
		record.CostPrice = 0

		piece.bp.approveStay() //TODO: 承認確認中
		piece.bp.modified().addMessage(
			fmt.Sprintf("cost_price(商品原価) が負の数。`%d` → `0`(固定値) にクレンジング。", costPrice), "#2-01")

		// FIXME:
		piece.setStatus(MODIFY, false)
		piece.addMessage(fmt.Sprintf("cost_price(商品原価) が負の数。`%d` → `0`(固定値) にクレンジング。", costPrice), "#2-01")
	}

	return &piece
}

// FUNCTION: レコード毎のクレンジング後データ登録
func (cs *ProductsClensing) saveData(record *legacy.Product, piece *ProductPiece) {
	// PROCESS: REMOVEDの場合はDBに登録しない
	// if piece.status == REMOVE { FIXME:
	if piece.bp.isRemove() {
		// cs.setResult(piece)
		return
	}

	// PROCESS: データ登録
	// INFO: cleanテーブル
	rec := clean.Product{
		ProductName: record.ProductName,
		CostPrice:   record.CostPrice,
		CreatedBy:   OPERATION_USER,
		UpdatedBy:   OPERATION_USER,
	}
	err := rec.Insert(ctx, cs.conns.WorkDB, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納、未承認扱い)
	if err != nil {
		piece.bp.dbError(err)
		// FIXME:
		// piece.bp.setStatus(REMOVE).setApprove(NOT_FINDED).addMessage(redFont(fmt.Sprintf("%v", err)), "")

		// FIXME:
		piece.setStatus(REMOVE, false)
		piece.addMessage(fmt.Sprintf("%v", err), "")
	} else {
		// INFO: [商品名]登録
		cs.refData.ProductNameSet[record.ProductName] = struct{}{}
	}
	// FIXME:
	// cs.setResult(piece)
}

// FUNCTION: クレンジング結果の登録 FIXME:
func (cs *ProductsClensing) setResult(piece *ProductPiece) {
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
			piece.bp.status,
			piece.bp.approve,
			piece.bp.msg,

			// FIXME:
			// piece.status,
			// approveStr(piece.approved),
			// piece.message,
		)
	}

	return msg
}
