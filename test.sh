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
    
    echo ""
    echo "===== Testing $site_name ====="
    ./job-extractor "$url" "output/test/${site_name}.json"
    
    if [ $? -eq 0 ]; then
        echo "✓ $site_name: Success"
        # 結果を全て表示
        if [ -f "output/test/${site_name}.json" ]; then
            echo "Full output:"
            cat "output/test/${site_name}.json"
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
test_site "pharmacareer" "https://pharmacareer.jp/job/j-1067786/"
test_site "benesse-mcm" "https://kango.benesse-mcm.jp/p13/c0399/fac008783/jobN136124/"

# TODO: 以下のサイトは実際の求人URLに置き換えてコメントを外す
# test_site "cme-pharmacist" "https://www.cme-pharmacist.jp/job/[job-id]/"
# test_site "kiracare" "https://job.kiracare.jp/offer/[offer-id]/"
# test_site "nursejj" "https://www.nursejj.com/job/[job-id]/"
# test_site "nursepower" "https://www.nursepower.co.jp/job/[job-id]/"
# test_site "supernurse" "https://www.supernurse.co.jp/job/[job-id]/"
# test_site "th-agent" "https://www.th-agent.jp/job/[job-id]/"
# test_site "yakumatch" "https://kangoshi.yakumatch.com/job/[job-id]/"

echo ""
echo "All tests completed. Results saved in output/test/"
