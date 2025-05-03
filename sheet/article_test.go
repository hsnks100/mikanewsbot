package sheet

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestInsertNewsAtomic(t *testing.T) {
	// .env 파일을 이용해 환경변수 로드 (테스트 환경에서도 동일하게 사용하기 위함)
	_ = godotenv.Load()
	fmt.Println("NEWS_SPREADSHEET_ID: ", os.Getenv("NEWS_SPREADSHEET_ID"))

	spreadsheetID := os.Getenv("NEWS_SPREADSHEET_ID")
	if spreadsheetID == "" {
		t.Skip("NEWS_SPREADSHEET_ID 환경변수가 설정되어 있지 않아 테스트를 건너뜁니다.")
	}

	item := NewsItem{
		Date:    time.Now().In(time.Local).Format("2006/01/02 15:04"),
		Subject: "&quot;블록체인도 외부 검증&quot;...XPLA, ISAE 3000 인증 2년 연속 획득",
		URL:     "https://www.digitaltoday.co.kr/news/articleView.html?idxno=561052",
	}

	if err := InsertNewsAtomic(spreadsheetID, item); err != nil {
		t.Fatalf("InsertNewsAtomic 실패: %v", err)
	}
}
