# データ移行用コマンドラインツール

## 方針

* legacyDB(移行元)のダンプデータを事前に入手し、ローカルで立ち上げたDBにLoadする。
* legacyDBのデータを組み合わせて、productDB(移行先)のデータにコンバートする。
* ローカルで立ち上げたproductDBに登録する。
* ソースコードのversionに追随してタグ管理を行うことで、開発の進捗に合わせて必要なテーブルから順次移行する。
* ツール実行により、実行Log、エラーデータの一覧をマークダウンで出力する。
* productDBから、local用、production用の、ダンプデータを作成する。
  * local用・・・ローカル環境に、データ移行作業で作成したデータを投入することを想定したダンプデータ。マイグレーションによって作成される初期投入データ、およびバッチ処理で登録するDX-supportのデータについては含まない。
  * production用・・・本番/ステージング環境(AWS)に投入するためのダンプデータ。初期投入データ、DX-supportのデータも含む。

## 使い方

1. リリースサイトから最新版のツールをダウンロードする。
2. `.env`ファイルを作成し、移行元DB/移行先DBのアクセス情報を設定する。

    ``` cmd
    # LEGACY_DB
    LEGACY_MARIADB_USER=maria
    LEGACY_MARIADB_PASSWORD=password
    LEGACY_MARIADB_HOST=localhost
    LEGACY_MARIADB_PORT=6001
    LEGACY_MARIADB_DB=legacyDB

    # WORK_DB
    WORK_POSTGRES_USER=postgres
    WORK_POSTGRES_PASSWORD=password
    WORK_POSTGRES_HOST=localhost
    WORK_POSTGRES_PORT=6101
    WORK_POSTGRES_DB=workDB

    # PRODUCT_DB
    PRODUCT_POSTGRES_USER=postgres
    PRODUCT_POSTGRES_PASSWORD=password
    PRODUCT_POSTGRES_HOST=localhost
    PRODUCT_POSTGRES_PORT=6201
    PRODUCT_POSTGRES_DB=productDB
    ```

3. `exe`ファイルを実行する。

    ``` cmd
    data-transfer.exe migrate
    ```

4. `実行Log`、および必要な場合は`エラーデータの一覧`を確認する。

## 移行元情報

* [移行元DBレイアウト](docs/source-db.md)

## 移行設計

> [!NOTE]
> TBD.

## 変更履歴

### v0.1.0

* versionコマンドの実装

### v0.2.0

* maigration実行ログの出力
* 移行元データの件数表示

-----

> [!IMPORTANT]
> go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-mysql@latest

-----

> [!IMPORTANT]
> go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest

-----

> [!TIP]
> Helpful advice for doing things better or more easily.

-----

-----

> [!WARNING]
> Urgent info that needs immediate user attention to avoid problems.

-----

> [!CAUTION]
> Advises about risks or negative outcomes of certain actions.
