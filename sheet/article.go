package sheet

import (
	"context"
	"fmt"

	_ "embed"

	"golang.org/x/oauth2/google"
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
	for {
		lastRow := len(sheet.Rows)
		if startRow >= lastRow {
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

func InsertNews(spreadsheetID string, item NewsItem) error {
	conf, err := google.JWTConfigFromJSON(credentials, spreadsheet.Scope)
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client := conf.Client(context.TODO())

	service := spreadsheet.NewServiceWithClient(client)
	spreadsheet, err := service.FetchSpreadsheet(spreadsheetID)
	if err != nil {
		return fmt.Errorf("unable to retrieve spreadsheet: %v", err)
	}

	sheet, err := spreadsheet.SheetByIndex(0)
	if err != nil {
		return fmt.Errorf("unable to retrieve sheet: %v", err)
	}

	updateRow(sheet, 0, map[int]string{
		1: item.Date,
		3: item.Subject,
		5: item.URL,
	})

	return nil
}
