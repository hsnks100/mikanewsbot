package sheet

import (
	"context"
	"fmt"
	"time"

	_ "embed"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	spreadsheet "gopkg.in/Iwark/spreadsheet.v2"
)

//go:embed credentials.json
var credentials []byte

// 시트에 특정 행을 업데이트하는 함수
// colValues: (컬럼 인덱스 → 문자열 값) 맵
func updateRow(
	sheet *spreadsheet.Sheet,
	startRow int,
	colValues map[int]string,
) error {
	if len(colValues) == 0 {
		return nil // 업데이트할 내용이 없다면 그냥 리턴
	}

	// map에서 '어떤 컬럼'을 기준으로 비어있는 행을 판단할지 정해야 함
	// 대개는 "가장 중요한 컬럼 하나"를 기준으로 비어있는지 확인하는 경우가 많습니다.
	// 여기서는 map 중 첫번째 key를 골라서 그 컬럼이 비어있는지 확인한다고 가정.
	// (실무에서는 주로 '날짜'나 '제목' 컬럼 등, 꼭 들어가야 할 데이터를 기준으로 하곤 함.)
	var firstColumn int
	for colIndex := range colValues {
		firstColumn = colIndex
		fmt.Println("first col: ", firstColumn)
		break
	}

	// 비어있는 행 탐색
	// 총 행의 길이, 어떤 이유로 어떤 포지션이 선택되었는지 로그 자세히...
	// 해당 위치가 반드시 비어있어야 사용가능한 상태로 해야함
	for {
		lastRow := len(sheet.Rows)
		fmt.Println("last row: ", lastRow)
		if startRow >= lastRow {
			fmt.Println("start row is greater than last row")
			break
		}
		if sheet.Rows[startRow][firstColumn].Value == "" {
			break
		}

		startRow++
	}

	// map을 순회하며 해당 컬럼을 업데이트
	for colIndex, val := range colValues {
		sheet.Update(startRow, colIndex, val)
	}

	// 구글 스프레드시트에 반영
	if err := sheet.Synchronize(); err != nil {
		return fmt.Errorf("unable to synchronize sheet: %v", err)
	}

	return nil
}

type NewsItem struct {
	Date         string
	MediaCompany string
	Subject      string
	Keyword      string
	URL          string
}

// InsertNews: 함수 원형 유지
func InsertNewsAtomic(spreadsheetID string, item NewsItem) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 1) 서비스‑계정 자격으로 Sheets 서비스 생성
	conf, err := google.JWTConfigFromJSON(credentials, sheets.SpreadsheetsScope)
	if err != nil {
		return fmt.Errorf("parse credentials: %w", err)
	}
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(conf.Client(ctx)))
	if err != nil {
		return fmt.Errorf("create sheets service: %w", err)
	}

	// 2) “첫 번째 시트 탭” 이름 알아오기 (Sheet1 고정이 아닐 수 있으므로)
	ss, err := srv.Spreadsheets.Get(spreadsheetID).
		Fields("sheets.properties").
		Do()
	if err != nil {
		return fmt.Errorf("get spreadsheet: %w", err)
	}
	fmt.Println("ss: ", ss.Sheets)
	// if len(ss.Sheets) == 0 {
	// 	return fmt.Errorf("spreadsheet has no sheets")
	// }
	sheetName := "News" // ss.Sheets[0].Properties.Title // 예: "Sheet1"

	// 현재 시트의 행 개수 구하기 (A열 기준)
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, fmt.Sprintf("%s!A:A", sheetName)).Do()
	if err != nil {
		return fmt.Errorf("get row count: %w", err)
	}
	rowCount := len(resp.Values)
	rowNumber := rowCount + 1 // 새로 추가될 행 번호 (1-based)

	// C열에 수식 삽입
	formula := fmt.Sprintf("=INDEX('매체 리스트'!$B$3:$B$3000, MATCH(TRUE,ISNUMBER(SEARCH('매체 리스트'!$C$3:$C$3000, G%d)), 0))", rowNumber)
	// =HYPERLINK(G1389,E1389)
	hyper := fmt.Sprintf("=HYPERLINK(G%d,E%d)", rowNumber, rowNumber)
	// =TEXTJOIN(", ", TRUE, FILTER('키워드 목록'!$B$3:$B$999, ISNUMBER(SEARCH('키워드 목록'!$B$3:$B$999, E1389))))
	keywordFormular := fmt.Sprintf("=TEXTJOIN(\", \", TRUE, FILTER('키워드 목록'!$B$3:$B$999, ISNUMBER(SEARCH('키워드 목록'!$B$3:$B$999, E%d))))", rowNumber)
	row := []interface{}{rowNumber - 1, item.Date, formula, hyper, item.Subject, keywordFormular, item.URL}
	vr := &sheets.ValueRange{Values: [][]interface{}{row}}

	// 4) values.append 호출 – 맨 아래 새 행 삽입
	_, err = srv.Spreadsheets.Values.Append(
		spreadsheetID,
		fmt.Sprintf("%s!A1", sheetName), // A1 아무 셀 지정
		vr).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("append row: %w", err)
	}
	return nil
}
