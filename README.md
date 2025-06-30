# Job Data Extractor

求人サイトから自動的にデータを抽出するツール

## ディレクトリ構造

```
.
├── src/                    # ソースコード
│   ├── universal-extractor.go  # 汎用抽出ツール（推奨）
│   ├── job-extractor.go    # kirara-support専用ツール
│   ├── scraper.go          # XPathベースのスクレイパー
│   └── browser-scraper.go  # ブラウザレンダリング版（開発中）
├── format/                 # フォーマット定義
│   ├── format.json         # 空のテンプレート
│   └── sample*.json        # サンプルXPath設定
├── configs/                # 設定ファイル
│   └── sites/             # サイト別設定
│       ├── kyujiner.json  # 求人ERの設定
│       └── example-site.json  # サンプル設定
├── examples/               # テスト用HTMLファイル
├── output/                 # 出力されたJSONファイル
├── docs/                   # ドキュメント
│   └── CUSTOMIZATION.md   # カスタマイズガイド
└── go.mod                  # Go依存関係
```

## 基本的な使い方

### 1. 汎用抽出ツール（universal-extractor）

URLから直接データを抽出して標準出力に表示：
```bash
go run src/universal-extractor.go "https://example.com/job/123"
```

ファイルに保存（自動サイト検出）：
```bash
go run src/universal-extractor.go "https://example.com/job/123" output.json
```

特定のサイト設定を指定：
```bash
go run src/universal-extractor.go --config custom-site "https://example.com/job/123"
```

ヘルプを表示：
```bash
go run src/universal-extractor.go -h
```

### 2. ビルドして使用

```bash
# ビルド
go build -o job-extractor src/universal-extractor.go

# 実行
./job-extractor "https://example.com/job/123" result.json
```

## 新しいサイトへの対応方法

### ステップ1: サイトの構造を調査

ブラウザの開発者ツールで求人ページを開き、以下の要素を確認：
- 求人タイトルのCSSセレクター
- 給与情報の場所
- 各フィールドのHTML構造

### ステップ2: 設定ファイルを作成

`configs/sites/your-site.json` を作成：

```json
{
    "name": "your-site",
    "domain": "your-site.com",
    "selectors": {
        "name": "h1.job-title",
        "price": "span.salary",
        "facility_name": "div.company-name",
        "area": "div.location",
        "occupation": "span.job-type",
        "contract": "span.employment-type",
        "detail": "div.job-description",
        "required_skill": "div.requirements",
        "holiday": "dt:contains('休日') + dd",
        "working_hours": "dt:contains('勤務時間') + dd",
        "welfare_program": "dt:contains('福利厚生') + dd"
    }
}
```

### ステップ3: 実行

```bash
# 自動検出（URLのドメインから設定を自動選択）
go run src/universal-extractor.go "https://your-site.com/job/123" result.json

# 明示的に設定を指定
go run src/universal-extractor.go "https://your-site.com/job/123" result.json your-site
```

## 実例：新しいサイトの追加

### 例1: 求人ER (kyujiner.com) の場合

1. サイト構造を確認
2. `configs/sites/kyujiner.json` を作成：

```json
{
    "name": "kyujiner",
    "domain": "kango.kyujiner.com",
    "selectors": {
        "name": "p.ichiran_t_d_name",
        "price": "dt:contains('給与') + dd",
        "facility_name": "dt:contains('施設形態') + dd",
        "area": "dt:contains('勤務地') + dd",
        "contract": "dt:contains('雇用形態') + dd",
        "detail": "dt:contains('仕事内容') + dd",
        "working_hours": "dt:contains('勤務時間') + dd"
    }
}
```

3. 実行：
```bash
# 自動検出で実行
go run src/universal-extractor.go "https://kango.kyujiner.com/job/13249" result.json

# 明示的に設定を指定
go run src/universal-extractor.go --config kyujiner "https://kango.kyujiner.com/job/13249"
```

## CSSセレクターの書き方

### 基本セレクター
- `h1.job-title` - class="job-title"のh1要素
- `div#main` - id="main"のdiv要素
- `.salary span` - class="salary"内のspan要素

### jQuery風セレクター（goquery対応）
- `dt:contains('給与') + dd` - "給与"を含むdtの次のdd要素
- `tr:contains('勤務地') td` - "勤務地"を含むtr内のtd要素

## トラブルシューティング

### データが取得できない場合

1. **セレクターの確認**
   ```javascript
   // ブラウザのコンソールで確認
   document.querySelector('your-selector')
   ```

2. **設定ファイルの場所**
   - 必ず `configs/sites/` ディレクトリ内に配置
   - ファイル名は `サイト名.json` 形式

3. **デバッグ方法**
   - まず設定なしで実行してJSON-LDの取得状況を確認
   - 段階的にセレクターを追加

## 対応済みサイト

- **benesse-mcm.jp** - 看護師求人（configs/sites/benesse-mcm.json）
- **cme-pharmacist.jp** - 薬剤師求人（configs/sites/cme-pharmacist.json）
- **kango.kyujiner.com** - 看護師求人（configs/sites/kyujiner.json）
- **kirara-support.jp** - 看護師求人（configs/sites/kirara-support.json）
- **mc-nurse.net** - 看護師求人（configs/sites/mc-nurse.json）
- **nurse-pw.jp** - 看護師求人（configs/sites/nursepower.json）
- **nursejj.com** - 看護師求人（configs/sites/nursejj.json）
- **supernurse.co.jp** - 看護師求人（configs/sites/supernurse.json）
- **th-agent.jp** - 登録販売者求人（configs/sites/th-agent.json）
- **yakumatch.com** - 薬剤師求人（configs/sites/yakumatch.json）

## 出力フォーマット

```json
{
    "name": "求人タイトル",
    "price": "給与",
    "area": "エリア",
    "facility_name": "施設名",
    "dept": "診療科目",
    "occupation": "職種",
    "contract": "雇用形態",
    "detail": "仕事内容",
    "required_skill": "必要スキル",
    "holiday": "休日",
    "working_hours": "勤務時間",
    "welfare_program": "福利厚生",
    ...
}
```