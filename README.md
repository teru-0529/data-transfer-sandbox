# データ移行用コマンドラインツール

## 方針

* 移行元のデータを事前に入手し、ローカルで立ち上げたDBにLoadする。
* 登録済みのデータを組み合わせて、移行先のデータにコンバートする。
* 移行先のデータは、ローカルで立ち上げた新システムのDBに登録する。(必要に応じてダンプを取得して本番機等へ投入する)
* ソースコードのversionに追随してタグ管理を行うことで、開発の進捗に合わせて必要なテーブルから順次移行する。
* ツール実行により、実行Log、エラーデータの一覧をマークダウンで出力する。

## 前提

* 移行元のデータスキーマ/移行先のデータスキーマがDB定義され、postgresqlとして起動していること。
* 移行元のDBにコンバート対象データが登録されていること。
* 移行先のDBが空であること。

## 使い方

1. リリースサイトから最新版のツールをダウンロードする。
2. `.env`ファイルを作成し、移行元DB/移行先DBのアクセス情報を設定する。

    ``` powershell
    # SOURCE_DB
    SOURCE_POSTGRES_USER=postgres
    SOURCE_POSTGRES_PASSWORD=password
    SOURCE_POSTGRES_HOST_NAME=localhost
    SOURCE_POSTGRES_PORT=9901
    SOURCE_POSTGRES_DB=sourceDB

    # DEST_DB
    DEST_POSTGRES_USER=postgres
    DEST_POSTGRES_PASSWORD=password
    DEST_POSTGRES_HOST_NAME=localhost
    DEST_POSTGRES_PORT=9902
    DEST_POSTGRES_DB=destDB
    ```

3. `exe`ファイルを実行する。

    ``` powershell
    data-transfer.exe migrate
    ```

4. `実行Log`、および必要な場合は`エラーデータの一覧`を確認する。
