# サイト設定ファイル作成手順書

## 概要
この手順書は、新しい求人サイトに対応するための設定ファイル（`configs/sites/`配下のJSONファイル）を作成する方法を説明します。

## 作成手順

### 1. 対象サイトのHTML構造を調査

まず、対象となる求人詳細ページのHTML構造を分析します。

```bash
# 方法1: curlでHTMLをダウンロード
curl -s "https://example.com/job/12345" -o examples/example-site.html

# 方法2: WebFetchで構造分析
# URLと以下のプロンプトを使用：
# "求人詳細ページのHTML構造を分析して、各情報（求人タイトル、給与、施設名、勤務地、職種、雇用形態など）がどのようなHTML要素・クラス名に含まれているか教えてください"
```

### 2. ブラウザの開発者ツールで要素を確認

1. Chrome/Firefoxで求人詳細ページを開く
2. F12キーで開発者ツールを開く
3. 要素選択ツール（矢印アイコン）で各項目をクリック
4. 以下の情報を収集：
   - 求人タイトル
   - 給与
   - 施設名
   - 勤務地
   - 職種
   - 雇用形態
   - その他の項目

### 3. セレクターのパターンを特定

#### 基本的なセレクターパターン

```css
/* クラス指定 */
h1.job-title
div.salary-info

/* ID指定 */
#job-name
#facility-info

/* 階層指定 */
div.content p
section.job-detail > h2

/* 属性指定 */
[data-field="salary"]
```

#### jQuery風セレクター（goquery対応）

```css
/* テキストを含む要素の隣接要素 */
dt:contains('給与') + dd
th:contains('勤務地') + td

/* テキストを含む要素の子要素 */
tr:contains('施設形態') td
div:contains('仕事内容') p
```

### 4. 設定ファイルを作成

`configs/sites/サイト名.json`を作成します。

#### 基本テンプレート

```json
{
    "name": "サイト名",
    "domain": "example.com",
    "selectors": {
        "name": "",
        "price": "",
        "facility_name": "",
        "area": "",
        "occupation": "",
        "contract": "",
        "dept": "",
        "detail": "",
        "required_skill": "",
        "holiday": "",
        "working_hours": "",
        "welfare_program": "",
        "license": "",
        "staff_comment": "",
        "station": "",
        "access": "",
        "working_style": "",
        "facility_type": "",
        "position": ""
    }
}
```

#### フィールド説明

| フィールド名 | 説明 | 例 |
|------------|------|-----|
| name | 求人タイトル | 看護師/准看護師 |
| price | 給与・年収 | 月給25万円〜35万円 |
| facility_name | 施設・会社名 | ○○病院 |
| area | 勤務地エリア | 東京都渋谷区 |
| occupation | 職種 | 看護師 |
| contract | 雇用形態 | 正社員(常勤) |
| dept | 診療科目 | 内科、外科 |
| detail | 仕事内容詳細 | 病棟での看護業務... |
| required_skill | 必要なスキル・経験 | 看護師経験3年以上 |
| holiday | 休日・休暇 | 土日祝休み |
| working_hours | 勤務時間 | 9:00〜18:00 |
| welfare_program | 福利厚生 | 社会保険完備 |
| license | 必要な資格 | 看護師免許 |
| staff_comment | スタッフコメント | アットホームな職場... |
| station | 最寄り駅 | JR渋谷駅 |
| access | アクセス方法 | 渋谷駅から徒歩5分 |
| working_style | 勤務形態 | 日勤のみ |
| facility_type | 施設形態 | 一般病院 |
| position | 役職・ポジション | 主任看護師 |

### 5. 実例：テーブル形式のデータ抽出

#### dt/dd形式の場合
```html
<dl>
  <dt>給与</dt>
  <dd>月給25万円〜35万円</dd>
  <dt>勤務地</dt>
  <dd>東京都渋谷区</dd>
</dl>
```

設定：
```json
{
    "price": "dt:contains('給与') + dd",
    "area": "dt:contains('勤務地') + dd"
}
```

#### th/td形式の場合
```html
<table>
  <tr>
    <th>診療科目</th>
    <td>内科、外科、整形外科</td>
  </tr>
  <tr>
    <th>休日</th>
    <td>土日祝休み、年間休日120日</td>
  </tr>
</table>
```

設定：
```json
{
    "dept": "th:contains('診療科目') + td",
    "holiday": "th:contains('休日') + td"
}
```

### 6. universal-extractor.goへの追加

`src/universal-extractor.go`の`detectSite`関数に新しいサイトの判定を追加：

```go
func detectSite(url string) string {
    // 既存のサイト判定...
    
    if strings.Contains(url, "example.com") {
        return "example-site"
    }
    
    return "default"
}
```

### 7. テスト実行

```bash
# 設定ファイルが正しく動作するかテスト
go run src/universal-extractor.go "https://example.com/job/12345" output/test.json

# 出力されたJSONを確認
cat output/test.json
```

### 8. デバッグのヒント

1. **段階的にテスト**
   - まず`name`と`price`だけでテスト
   - 動作確認後、他のフィールドを追加

2. **セレクターの検証**
   ```javascript
   // ブラウザのコンソールで確認
   document.querySelector('your-selector')
   jQuery('your-selector').text()
   ```

3. **一般的な問題と解決策**
   - 空白文字：セレクターで`.trim()`相当の処理は自動実行
   - 文字化け：`encoding`フィールドで文字コードを指定
   - 動的コンテンツ：現状は静的HTMLのみ対応

### 9. 完成例

```json
{
    "name": "nursejj",
    "domain": "nursejj.com",
    "selectors": {
        "name": "h1.job-ttl",
        "price": "dl.info-list dt:contains('給与') + dd",
        "facility_name": "dl.info-list dt:contains('施設名') + dd",
        "area": "dl.info-list dt:contains('所在地') + dd",
        "occupation": "dl.info-list dt:contains('募集職種') + dd",
        "contract": "dl.info-list dt:contains('雇用形態') + dd",
        "dept": "table.detail-tbl th:contains('診療科目') + td",
        "detail": "div.job-detail-box h3:contains('仕事内容') + div",
        "holiday": "table.detail-tbl th:contains('休日・休暇') + td",
        "working_hours": "table.detail-tbl th:contains('勤務時間') + td",
        "welfare_program": "table.detail-tbl th:contains('待遇・福利厚生') + td"
    }
}
```

## 注意事項

- セレクターは大文字小文字を区別します
- 日本語のテキストマッチングも可能です
- `:contains()`は部分一致で動作します
- 複数の要素がマッチする場合、最初の要素のテキストが取得されます
- 必須フィールドは`name`のみで、他は省略可能です