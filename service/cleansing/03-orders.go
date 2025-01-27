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
type OrderPiece struct {
	OrderNo  int
	bp       *Piece
	status   Status
	approved bool
	message  string
}

// FUNCTION: ステータスの変更 FIXME:
func (p *OrderPiece) setStatus(status Status, approved bool) {
	p.status = judgeStatus(p.status, status)
	p.approved = p.approved && approved
}

// FUNCTION: メッセージの追加 FIXME:
func (p *OrderPiece) addMessage(msg string, id string) {
	p.message = genMessage(p.message, msg, id)
}

// STRUCT:
type OrdersClensing struct {
	conns   infra.DbConnection
	refData *RefData
	Result  Result
	Details []*OrderPiece
}

// FUNCTION:
func NewOrders(conns infra.DbConnection, refData *RefData) OrdersClensing {
	s := time.Now()

	// INFO: 固定値設定
	cs := OrdersClensing{
		conns:   conns,
		refData: refData,
		Result:  Result{TableNameJp: "受注", TableNameEn: "orders"},
	}
	log.Printf("[%s] table cleansing ...", cs.Result.TableNameEn)

	// PROCESS: 入力データ量
	cs.setEntryCount()

	// PROCESS: クレンジング先のtruncate
	_, err := queries.Raw(cs.Result.truncateSql()).ExecContext(ctx, cs.conns.WorkDB)
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
func (cs *OrdersClensing) setEntryCount() {
	// INFO: Legacyテーブル名
	num, err := legacy.Orders().Count(ctx, cs.conns.LegacyDB)
	if err != nil {
		log.Fatalln(err)
	}
	cs.Result.EntryCount = int(num)
}

// FUNCTION: データ繰り返し取得(1000件単位で分割)
func (cs *OrdersClensing) iterate() {
	bar := pb.Default.Start(cs.Result.EntryCount)
	bar.SetMaxWidth(80)

	// PROCESS: 1000件単位でのSQL実行に分割する
	for section := 0; section < cs.Result.sectionCount(); section++ {
		// INFO: Legacyテーブル名
		records, err := legacy.Orders(qm.Limit(LIMIT), qm.Offset(section*LIMIT)).All(ctx, cs.conns.LegacyDB)
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
func (cs *OrdersClensing) checkAndClensing(record *legacy.Order) *OrderPiece {
	// INFO: piece FIXME:
	piece := OrderPiece{
		OrderNo:  record.OrderNo,
		bp:       NewPiece(),
		status:   NO_CHANGE,
		approved: true,
	}

	// PROCESS: #3-01: order_dateが日付のフォーマットに合致しない場合は、"20250101"にクレンジングする。
	orderDate := record.OrderDate
	var defOrderDate = "20250101"

	_, err := time.Parse(DATE_LAYOUT, orderDate)
	if err != nil {
		record.OrderDate = defOrderDate

		piece.bp.modified().addMessage(
			fmt.Sprintf("order_date(受注日付) が日付フォーマットではない。`%s` → `%s`(固定値) にクレンジング。", orderDate, defOrderDate), "#3-01")

		// FIXME:
		piece.setStatus(MODIFY, true)
		piece.addMessage(
			fmt.Sprintf("order_date(受注日付) が日付フォーマットではない。`%s` → `%s`(固定値) にクレンジング。", orderDate, defOrderDate), "#3-01")
	}

	// PROCESS: #3-02: order_picが[担当者]に存在しない場合は、"N/A"にクレンジングする。
	orderPic := record.OrderPic
	var defOrderPic = "N/A"

	_, exist := cs.refData.OperatorNameSet[orderPic]
	if !exist {
		record.OrderPic = defOrderPic

		piece.bp.modified().addMessage(
			fmt.Sprintf("order_pic(受注担当者名) が[担当者]に存在しない。`%s` → `%s`(固定値) にクレンジング。", orderPic, defOrderPic), "#3-02")

		// FIXME:
		piece.setStatus(MODIFY, true)
		piece.addMessage(fmt.Sprintf("order_pic(受注担当者名) が[担当者]に存在しない。`%s` → `%s`(固定値) にクレンジング。", orderPic, defOrderPic), "#3-02")
	}

	return &piece
}

// FUNCTION: レコード毎のクレンジング後データ登録
func (cs *OrdersClensing) saveData(record *legacy.Order, piece *OrderPiece) {
	// PROCESS: REMOVEDの場合はDBに登録しない
	// if piece.status == REMOVE { FIXME:
	if piece.bp.isRemove() {
		// cs.setResult(piece)
		return
	}

	orderDate, _ := time.Parse(DATE_LAYOUT, record.OrderDate)

	// PROCESS: データ登録
	// INFO: cleanテーブル
	rec := clean.Order{
		OrderNo:      record.OrderNo,
		OrderDate:    orderDate,
		OrderPic:     record.OrderPic,
		CustomerName: record.CustomerName,
		CreatedBy:    OPERATION_USER,
		UpdatedBy:    OPERATION_USER,
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
		// INFO: [受注番号]登録
		cs.refData.OrderNoSet[record.OrderNo] = struct{}{}
	}
	// FIXME:
	// cs.setResult(piece)
}

// FUNCTION: クレンジング結果の登録 FIXME:
func (cs *OrdersClensing) setResult(piece *OrderPiece) {
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
func (cs *OrdersClensing) ShowDetails() string {
	if len(cs.Details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", cs.Result.TableName())

	msg += "  | # | order_no | … | RESULT | APPROVED | MESSAGE |\n"
	msg += "  |--:|--:|---|:-:|:-:|---|\n"
	for i, piece := range cs.Details {
		msg += fmt.Sprintf("  | %d | %d | … | %s | %s | %s |\n",
			i+1,
			piece.OrderNo,
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
