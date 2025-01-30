# データ移行用コマンドラインツール

## 前提

* 移行元のダンプデータは、月次で最新の情報を入手すること。(スキーマ定義が変わる場合は早急に情報を入手すること)
* 移行元のダンプデータを元にLoadした`legacyDB(移行元)`をコンテナ上で動かしていること。
* `productDB(移行先)`をコンテナ上で動かしていること。

## 方針

* `legacyDB(移行元)`のデータを組み合わせて`productDB(移行先)`のデータにコンバートを行います。
* アプリ開発のスプリントバージョンに当移行ツールが追随し、開発の進捗に必要なテーブルから順次移行の対応を進めます。
* コンバートは、`1.cleansing`、`2.transfer`の2つから構成します。
  1. cleansing: `legacyDB(移行元)`のデータについて、<b>現行システムの仕様上</b>のテーブルの構成としてつじつまの合わないデータを抽出し、確認の上で「データの変換」「データの削除」を行います。
  2. transfer: クレンジング後のデータをもとに`productDB(移行先)`への変換を行います。
* `1.cleansing`、`2.transfer`それぞれの処理結果は、MDで出力します。
* コンバート処理後`productDB(移行先)`のデータをもとに、各種ダンプデータを作成します。
  1. `dml-local.sql.gz`: 開発者がローカル環境で利用するダンプデータです。データのみのダンプデータで、マイグレーションにより作成される初期投入データ、DX-supportの設定データ等は含みません。
  2. `ddl-aws.sql.gz`: 本番/ステージング環境に投入するためのスキーマ情報ダンプデータです。
  3. `dml-aws.sql.gz`: 本番/ステージング環境に投入するためのデータ情報ダンプデータです。初期投入データ、DX-supportの設定データ等も含みます。

## 使い方

1. [リリースサイト](https://github.com/teru-0529/data-transfer-sandbox/releases/latest)から最新版のツールをダウンロードする。
2. `.env`ファイルに、移行元DB/移行先DBのアクセス情報を設定する。

    ``` cmd

    # BASE_INFO
    LEGACY_LOAD_FILE=L202501   # 移行元のファイル名(拡張子無し)
    APP_VERSION=v1.0.0   # アプリケーションのバージョン

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
    REM 1.クレンジング
    data-transfer.exe cleansing
    REM 2.移行変換
    data-transfer.exe transfer
    ```

4. `実行Log`を確認する。
5. 出力されたダンプファイルを活用する。

    ローカルのコンテナDBにLoadする手順

    ``` cmd
    REM 1.コンテナ内部にコピー
    docker cp ./dist/dev(L202501)/dml-local.sql.gz product-db:/tmp/dump.sql.gz
    REM 2.Load実行
    docker exec -it product-db bash -c "echo 'gzip -d -c /tmp/dump.sql.gz | psql -U postgres -d productDB'"
    ```

## クレンジング仕様

* [仕様書はこちら](docs/cleansing-spec.md)

## 移行変換仕様

* [仕様書はこちら](docs/transfer-spec.md)

## 移行対応状況

  | # | SCHEMA | TABLE | 移行対象 | ローカルダンプ対象外 | … | 対応Ver | … | 備考 |
  |--:|---|---|:--:|:--:|---|---|---|---|
  | 1. | orders | operators(担当者) | ⭕ |  | … | v0.1.0 | … |  |
  | 2. | orders | products(商品) | ⭕ |  | … | v0.1.0 | … | |
  | 3. | orders | orders(受注) | ⭕ |  | … | <span style="color:red;"><b>v0.2.0🔆</b></span> | … |  |
  | 4. | orders | order_details(受注明細) | ⭕ | ✅ | … | <span style="color:red;"><b>v0.2.0🔆</b></span> | … | ローカルダンプ対象外はテスト用の対応 |

## 変更履歴

### v0.3.0-rc.1

* リリースモジュールの漏れ対応

### v0.3.0-rc.0

* cleansing機能完成
* cleanDBへのLoad機能完成

<details><summary>(open) 過去の更新履歴</summary>

### v0.2.0

* maigration実行ログの出力
* 移行元データの件数表示

### v0.1.0

* versionコマンドの実装

</details>

-----
