# スクレイパー作業計画

指定するサイトの求人情報を巡回し、各情報を取得する XPath を収集して欲しい。
format フォルダに 3 つのサンプルとフォーマットが入っています。
それを参考に、指定するページから該当の項目を取得できる XPath を取得し json の該当の場所を埋めてください。

## 新しいサイト追加手順（重要）

### 1. サイトのHTML構造を調査
```bash
# URLを指定してHTMLをダウンロード
curl -s "対象URL" -o examples/site-name.html

# または WebFetch でHTML構造を分析
WebFetch: URL + "求人詳細ページのHTML構造を分析して、各情報（求人タイトル、給与、施設名、勤務地、職種、雇用形態など）がどのようなHTML要素・クラス名に含まれているか教えてください"
```

### 2. 設定ファイルを作成
`configs/sites/サイト名.json` を作成：

```json
{
    "name": "サイト名",
    "domain": "ドメイン名",
    "selectors": {
        "name": "求人タイトルのセレクター",
        "price": "給与のセレクター",
        "facility_name": "施設名のセレクター",
        "area": "勤務地のセレクター",
        "occupation": "職種のセレクター",
        "contract": "雇用形態のセレクター",
        "dept": "診療科目のセレクター",
        "detail": "仕事内容のセレクター",
        "required_skill": "必要スキルのセレクター",
        "holiday": "休日のセレクター",
        "working_hours": "勤務時間のセレクター",
        "welfare_program": "福利厚生のセレクター",
        "license": "必要資格のセレクター",
        "staff_comment": "スタッフコメントのセレクター",
        "station": "最寄り駅のセレクター",
        "access": "アクセスのセレクター",
        "working_style": "勤務形態のセレクター",
        "facility_type": "施設形態のセレクター",
        "position": "役職のセレクター"
    }
}
```

### 3. universal-extractor.go に自動検出を追加（オプション）
```go
func detectSite(url string) string {
    if strings.Contains(url, "新しいサイトのドメイン") {
        return "サイト名"
    }
    // 既存のコード...
}
```

### 4. テスト実行
```bash
# 自動検出で実行
go run src/universal-extractor.go "対象URL" output/test.json

# 設定を明示的に指定
go run src/universal-extractor.go "対象URL" output/test.json サイト名
```

## セレクターの書き方例

### 基本パターン
- `h1.job-title` - クラス指定
- `#job-name` - ID指定
- `div.content p` - 階層指定

### テーブル形式（dt/dd）
```json
"price": "dt:contains('給与') + dd",
"area": "dt:contains('勤務地') + dd"
```

### テーブル形式（th/td）
```json
"dept": "th:contains('診療科目') + td",
"holiday": "th:contains('休日') + td"
```

### 実例（kyujiner）
```json
{
    "name": "p.ichiran_t_d_name",
    "price": "dt:contains('給与') + dd",
    "facility_name": "dt:contains('施設形態') + dd",
    "area": "dt:contains('勤務地') + dd"
}
```

### 実例（kirara-support）- 内蔵
```json
{
    "name": "h2.bl_jobPost_title",
    "price": "dl.bl_jobPost_table dt:contains('給与') + dd",
    "dept": "table.bl_defTable th:contains('診療科目') + td"
}
```

## デバッグのコツ

1. まず設定なしで実行してJSON-LDが取れるか確認
2. ブラウザの開発者ツールでセレクターを検証
3. 最初は name と price だけでテスト
4. 段階的にフィールドを追加

## 注意点
- セレクターはgoqueryの記法（jQuery風）
- :contains() は部分一致
- + は隣接する次の要素
- 日本語のテキストマッチングも可能