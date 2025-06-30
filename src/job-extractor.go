package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

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

func extractJobData(htmlContent string) (*JobData, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	data := &JobData{}

	// Extract from structured data (JSON-LD)
	structuredDataRegex := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*>(.*?)</script>`)
	matches := structuredDataRegex.FindAllStringSubmatch(htmlContent, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			var jsonData []map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &jsonData); err == nil {
				for _, item := range jsonData {
					if item["@type"] == "JobPosting" {
						// Extract job title from description
						if desc, ok := item["description"].(string); ok {
							lines := strings.Split(desc, "<br>")
							if len(lines) > 0 {
								data.Name = strings.TrimSpace(lines[0])
								data.TitleOriginal = strings.TrimSpace(lines[0])
							}
						}
						
						// Extract salary
						if baseSalary, ok := item["baseSalary"].(map[string]interface{}); ok {
							if value, ok := baseSalary["value"].(map[string]interface{}); ok {
								if min, ok := value["minValue"].(string); ok {
									if max, ok := value["maxValue"].(string); ok {
										data.Price = fmt.Sprintf("年収 %s〜%s円", min, max)
									}
								}
							}
						}
						
						// Extract location
						if jobLocation, ok := item["jobLocation"].(map[string]interface{}); ok {
							if address, ok := jobLocation["address"].(map[string]interface{}); ok {
								if region, ok := address["addressRegion"].(string); ok {
									data.Prefecture = region
								}
								if locality, ok := address["addressLocality"].(string); ok {
									data.City = locality
								}
								if street, ok := address["streetAddress"].(string); ok {
									data.Address = fmt.Sprintf("%s%s%s", data.Prefecture, data.City, street)
								}
								data.Area = data.Prefecture + data.City
							}
						}
						
						// Extract organization name
						if org, ok := item["hiringOrganization"].(map[string]interface{}); ok {
							if name, ok := org["name"].(string); ok {
								data.FacilityName = name
							}
						}
						
						// Extract employment type
						if empType, ok := item["employmentType"].(string); ok {
							if empType == "FULL_TIME" {
								data.Contract = "正社員(常勤)"
							}
						}
						
						// Extract job title (this should be dept, not occupation)
						if title, ok := item["title"].(string); ok {
							// title is "看護師" which is actually the occupation, not dept
							if data.Occupation == "" {
								data.Occupation = title
							}
						}
					}
				}
			}
		}
	}

	// Extract additional data from DOM
	// Extract facility type from title
	titleText := doc.Find("title").Text()
	if strings.Contains(titleText, "クリニック") {
		data.FacilityType = "クリニック"
	} else if strings.Contains(titleText, "病院") {
		data.FacilityType = "病院"
	}

	// Extract detailed info from tables
	doc.Find("table.bl_defTable tr").Each(func(i int, s *goquery.Selection) {
		th := strings.TrimSpace(s.Find("th").Text())
		td := strings.TrimSpace(s.Find("td").Text())
		
		switch th {
		case "必要な資格":
			data.License = td
		case "必要な業務経験":
			data.RequiredSkill = td
		case "仕事内容":
			data.Detail = td
		case "福利厚生":
			data.WelfareProgram = td
		case "就業時間":
			data.WorkingHours = td
		case "勤務形態":
			data.WorkingStyle = td
		case "休日":
			data.Holiday = td
		case "最寄駅":
			data.Station = td
		case "施設形態":
			data.FacilityType = td
		case "診療科目":
			data.Dept = td
		}
	})

	// Extract from basic info section (dl.bl_jobPost_table)
	doc.Find("dl.bl_jobPost_table").Each(func(i int, s *goquery.Selection) {
		dt := strings.TrimSpace(s.Find("dt").Text())
		dd := strings.TrimSpace(s.Find("dd").Text())
		
		switch dt {
		case "給与":
			if data.Price == "" {
				data.Price = dd
			}
		case "施設名":
			if data.FacilityName == "" {
				data.FacilityName = dd
			}
		case "勤務地":
			if data.Address == "" {
				data.Address = dd
			}
		case "最寄り駅":
			if data.Access == "" {
				data.Access = dd
			}
		case "職種":
			if data.Occupation == "" {
				data.Occupation = dd
			}
		case "雇用形態":
			if data.Contract == "" {
				data.Contract = dd
			}
		}
	})

	// Extract staff comment
	staffComment := doc.Find("div.bl_bulletList.bl_bulletList__nobull h3").Text()
	if staffComment != "" {
		data.StaffComment = strings.TrimSpace(staffComment)
	}

	// Extract position from job title
	if strings.Contains(data.Name, "管理職") {
		data.Position = "管理職候補"
	}

	return data, nil
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

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <url_or_file> <output_file>")
		fmt.Println("Example: go run main.go https://example.com/job result.json")
		fmt.Println("Example: go run main.go job.html result.json")
		os.Exit(1)
	}

	input := os.Args[1]
	outputFile := os.Args[2]

	var htmlContent string
	var err error

	// Check if input is URL or file
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		// It's a URL
		fmt.Printf("Fetching data from URL: %s\n", input)
		htmlContent, err = fetchURL(input)
		if err != nil {
			log.Fatal("Error fetching URL:", err)
		}
	} else {
		// It's a file
		fmt.Printf("Reading from file: %s\n", input)
		content, err := ioutil.ReadFile(input)
		if err != nil {
			log.Fatal("Error reading file:", err)
		}
		htmlContent = string(content)
	}

	// Extract job data
	jobData, err := extractJobData(htmlContent)
	if err != nil {
		log.Fatal("Error extracting job data:", err)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(jobData, "", "    ")
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
	}

	// Write to output file
	err = ioutil.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		log.Fatal("Error writing output file:", err)
	}

	fmt.Printf("Job data extracted successfully and saved to %s\n", outputFile)
}