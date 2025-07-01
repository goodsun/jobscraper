#!/bin/sh

# ビルド
echo "Building job-extractor..."
go build -o job-extractor src/universal-extractor.go || exit 1

# 出力ディレクトリ作成
mkdir -p output/test

# テスト関数
test_site() {
    site_name=$1
    url=$2
    output_color=$3  # 出力色指定（オプション）
    
    # デフォルトは白色
    if [ -z "$output_color" ]; then
        output_color="37"
    fi
    
    echo ""
    echo "===== Testing $site_name ====="
    printf "URL: \033[37m%s\033[0m\n" "$url"  # 白色でURL表示
    ./job-extractor "$url" "output/test/${site_name}.json"
    
    if [ $? -eq 0 ]; then
        echo "✓ $site_name: Success"
        # 結果を全て表示（指定色）
        if [ -f "output/test/${site_name}.json" ]; then
            echo "Full output:"
            printf "\033[${output_color}m"  # 指定色に設定
            cat "output/test/${site_name}.json"
            printf "\033[0m"   # リセット
        fi
    else
        echo "✗ $site_name: Failed"
    fi
}

# 各サイトをテスト
test_site "mc-nurse" "https://mc-nurse.net/jobs/detail/25-FYSK8/"
test_site "kyujiner" "https://kango.kyujiner.com/job/13249/"
test_site "kango-oshigoto" "https://kango-oshigoto.jp/offer/150102/"
test_site "kirara-support" "https://job.kiracare.jp/offer/1326251/"
test_site "nurse-step" "https://www.nurse-step.com/tokyo/12605/employment_1/jobcategory_1/facilityform_5/id_567641/"
test_site "benesse-mcm" "https://kango.benesse-mcm.jp/p13/c0399/fac008783/jobN136124/"
test_site "cme-pharmacist" "https://www.cme-pharmacist.jp/job/job-405528/?i_num=10"
test_site "kiracare" "https://job.kiracare.jp/offer/1326251/"
test_site "nursejj" "https://www.nursejj.com/bs/tokyo/148.html"
test_site "nursepower" "https://www.nursepower.co.jp/result/detail/17270/"
test_site "supernurse" "https://www.supernurse.co.jp/%E7%9F%B3%E5%B7%9D%E7%9C%8C/%E9%87%91%E6%B2%A2%E5%B8%82/JO0000152729.html"
test_site "th-agent" "https://www.th-agent.jp/job/job-405062/?i_num=10"
test_site "yakumatch" "https://kangoshi.yakumatch.com/jobs/198966"

echo ""
echo ""
echo "pharmacarrer は 動的生成のサイトのため正しく取得できません"
test_site "pharmacareer" "https://pharmacareer.jp/job/j-1067786/" "31"  # 31は赤色

echo ""
echo "All tests completed. Results saved in output/test/"
