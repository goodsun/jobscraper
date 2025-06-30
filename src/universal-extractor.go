package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// 汎用的なフィールド定義
type JobData struct {
	Name            string `json:"name"`
	Price           string `json:"price"`
	Area            string `json:"area"`
	Access          string `json:"access"`
	Address         string `json:"address"`
	City            string `json:"city"`
	Prefecture      string `json:"prefecture"`
	Contract        string `json:"contract"`
	Dept            string `json:"dept"`
	Detail          string `json:"detail"`
	FacilityName    string `json:"facility_name"`
	FacilityType    string `json:"facility_type"`
	Holiday         string `json:"holiday"`
	License         string `json:"license"`
	Occupation      string `json:"occupation"`
	Position        string `json:"position"`
	RequiredSkill   string `json:"required_skill"`
	StaffComment    string `json:"staff_comment"`
	Station         string `json:"station"`
	WelfareProgram  string `json:"welfare_program"`
	WorkingHours    string `json:"working_hours"`
	WorkingStyle    string `json:"working_style"`
	TitleOriginal   string `json:"title_original"`
}

// サイト別の抽出ルール
type SiteConfig struct {
	Name        string                       `json:"name"`
	Domain      string                       `json:"domain"`
	Patterns    map[string]string           `json:"patterns"`
	Selectors   map[string]string           `json:"selectors"`
	Extractors  map[string]ExtractorConfig  `json:"extractors"`
}

type ExtractorConfig struct {
	Type     string `json:"type"`     // "selector", "regex", "json-ld"
	Value    string `json:"value"`    // CSS selector, regex pattern, etc.
	Attr     string `json:"attr"`     // attribute to extract (text, href, etc.)
	Index    int    `json:"index"`    // which match to use (default 0)
}

func fetchURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func detectSite(url string) string {
	if strings.Contains(url, "kirara-support.jp") {
		return "kirara-support"
	} else if strings.Contains(url, "kyujiner.com") {
		return "kyujiner"
	}
	// 他のサイトの判定を追加
	return "default"
}

func loadSiteConfig(siteName string) (*SiteConfig, error) {
	configPath := filepath.Join("configs", "sites", siteName+".json")
	
	// デフォルト設定がない場合は、内蔵の設定を使用
	if siteName == "kirara-support" {
		return getKiraraSupportConfig(), nil
	}
	
	// カスタム設定ファイルを読み込む
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	
	var config SiteConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}

func getKiraraSupportConfig() *SiteConfig {
	return &SiteConfig{
		Name:   "kirara-support",
		Domain: "kirara-support.jp",
		Selectors: map[string]string{
			"name":           "h2.bl_jobPost_title",
			"price":          "dl.bl_jobPost_table dt:contains('給与') + dd",
			"facility_name":  "dl.bl_jobPost_table dt:contains('施設名') + dd",
			"area":           "dl.bl_jobPost_table dt:contains('勤務地') + dd",
			"access":         "dl.bl_jobPost_table dt:contains('最寄り駅') + dd",
			"occupation":     "dl.bl_jobPost_table dt:contains('職種') + dd",
			"contract":       "dl.bl_jobPost_table dt:contains('雇用形態') + dd",
			"staff_comment":  "div.bl_bulletList.bl_bulletList__nobull h3",
			"dept":           "table.bl_defTable th:contains('診療科目') + td",
			"detail":         "table.bl_defTable th:contains('仕事内容') + td",
			"facility_type":  "table.bl_defTable th:contains('施設形態') + td",
			"holiday":        "table.bl_defTable th:contains('休日') + td",
			"license":        "table.bl_defTable th:contains('必要な資格') + td",
			"required_skill": "table.bl_defTable th:contains('必要な業務経験') + td",
			"station":        "table.bl_defTable th:contains('最寄駅') + td",
			"welfare_program": "table.bl_defTable th:contains('福利厚生') + td",
			"working_hours":  "table.bl_defTable th:contains('就業時間') + td",
			"working_style":  "table.bl_defTable th:contains('勤務形態') + td",
		},
	}
}

func extractWithSelector(doc *goquery.Document, selector string) string {
	return strings.TrimSpace(doc.Find(selector).First().Text())
}

func extractData(htmlContent string, config *SiteConfig) (*JobData, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	data := &JobData{}

	// JSON-LD extraction (共通)
	extractJSONLD(htmlContent, data)

	// セレクターベースの抽出
	if config.Selectors != nil {
		if selector, ok := config.Selectors["name"]; ok {
			data.Name = extractWithSelector(doc, selector)
			data.TitleOriginal = data.Name
		}
		if selector, ok := config.Selectors["price"]; ok {
			data.Price = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["facility_name"]; ok {
			data.FacilityName = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["area"]; ok {
			data.Area = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["access"]; ok {
			data.Access = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["occupation"]; ok {
			data.Occupation = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["contract"]; ok {
			data.Contract = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["staff_comment"]; ok {
			data.StaffComment = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["dept"]; ok {
			data.Dept = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["detail"]; ok {
			data.Detail = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["facility_type"]; ok {
			data.FacilityType = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["holiday"]; ok {
			data.Holiday = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["license"]; ok {
			data.License = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["required_skill"]; ok {
			data.RequiredSkill = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["station"]; ok {
			data.Station = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["welfare_program"]; ok {
			data.WelfareProgram = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["working_hours"]; ok {
			data.WorkingHours = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["working_style"]; ok {
			data.WorkingStyle = extractWithSelector(doc, selector)
		}
	}

	// 住所から都道府県と市区町村を抽出
	if data.Address != "" {
		extractLocationInfo(data)
	} else if data.Area != "" {
		data.Address = data.Area
		extractLocationInfo(data)
	}

	return data, nil
}

func extractJSONLD(htmlContent string, data *JobData) {
	// JSON-LD スキーマの抽出
	structuredDataRegex := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*>(.*?)</script>`)
	matches := structuredDataRegex.FindAllStringSubmatch(htmlContent, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			var jsonData []map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &jsonData); err == nil {
				for _, item := range jsonData {
					if item["@type"] == "JobPosting" {
						extractFromJobPosting(item, data)
					}
				}
			}
		}
	}
}

func extractFromJobPosting(item map[string]interface{}, data *JobData) {
	// タイトル
	if desc, ok := item["description"].(string); ok && data.Name == "" {
		lines := strings.Split(desc, "<br>")
		if len(lines) > 0 {
			data.Name = strings.TrimSpace(lines[0])
			data.TitleOriginal = data.Name
		}
	}
	
	// 給与
	if baseSalary, ok := item["baseSalary"].(map[string]interface{}); ok && data.Price == "" {
		if value, ok := baseSalary["value"].(map[string]interface{}); ok {
			if min, ok := value["minValue"].(string); ok {
				if max, ok := value["maxValue"].(string); ok {
					data.Price = fmt.Sprintf("年収 %s〜%s円", min, max)
				}
			}
		}
	}
	
	// 勤務地
	if jobLocation, ok := item["jobLocation"].(map[string]interface{}); ok {
		if address, ok := jobLocation["address"].(map[string]interface{}); ok {
			if region, ok := address["addressRegion"].(string); ok && data.Prefecture == "" {
				data.Prefecture = region
			}
			if locality, ok := address["addressLocality"].(string); ok && data.City == "" {
				data.City = locality
			}
			if street, ok := address["streetAddress"].(string); ok && data.Address == "" {
				data.Address = fmt.Sprintf("%s%s%s", data.Prefecture, data.City, street)
			}
			if data.Area == "" {
				data.Area = data.Prefecture + data.City
			}
		}
	}
	
	// 施設名
	if org, ok := item["hiringOrganization"].(map[string]interface{}); ok {
		if name, ok := org["name"].(string); ok && data.FacilityName == "" {
			data.FacilityName = name
		}
	}
	
	// 雇用形態
	if empType, ok := item["employmentType"].(string); ok && data.Contract == "" {
		if empType == "FULL_TIME" {
			data.Contract = "正社員(常勤)"
		}
	}
	
	// 職種
	if title, ok := item["title"].(string); ok && data.Occupation == "" {
		data.Occupation = title
	}
}

func extractLocationInfo(data *JobData) {
	// 都道府県の抽出
	prefectureRegex := regexp.MustCompile(`^([^都道府県]+[都道府県])`)
	if matches := prefectureRegex.FindStringSubmatch(data.Address); len(matches) > 1 {
		data.Prefecture = matches[1]
	}
	
	// 市区町村の抽出
	cityRegex := regexp.MustCompile(`[都道府県]([^区市町村]+[区市町村])`)
	if matches := cityRegex.FindStringSubmatch(data.Address); len(matches) > 1 {
		data.City = matches[1]
	}
	
	if data.Area == "" && data.Prefecture != "" && data.City != "" {
		data.Area = data.Prefecture + data.City
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run universal-extractor.go <url> <output_file> [site_config]")
		fmt.Println("Example: go run universal-extractor.go https://example.com/job result.json")
		fmt.Println("Example: go run universal-extractor.go https://example.com/job result.json custom-site")
		os.Exit(1)
	}

	url := os.Args[1]
	outputFile := os.Args[2]
	
	// サイト設定の自動検出または指定
	siteName := ""
	if len(os.Args) >= 4 {
		siteName = os.Args[3]
	} else {
		siteName = detectSite(url)
	}
	
	fmt.Printf("Using site configuration: %s\n", siteName)
	
	// サイト設定を読み込む
	config, err := loadSiteConfig(siteName)
	if err != nil {
		fmt.Printf("Warning: Could not load site config for %s, using generic extraction\n", siteName)
		config = &SiteConfig{Name: "default"}
	}

	// URLからHTMLを取得
	fmt.Printf("Fetching data from URL: %s\n", url)
	htmlContent, err := fetchURL(url)
	if err != nil {
		log.Fatal("Error fetching URL:", err)
	}

	// データを抽出
	jobData, err := extractData(htmlContent, config)
	if err != nil {
		log.Fatal("Error extracting data:", err)
	}

	// JSONに変換
	jsonData, err := json.MarshalIndent(jobData, "", "    ")
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
	}

	// ファイルに保存
	err = ioutil.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		log.Fatal("Error writing output file:", err)
	}

	fmt.Printf("Data extracted successfully and saved to %s\n", outputFile)
}