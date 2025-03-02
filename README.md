# 뉴스 봇 사용법

네이버 뉴스 API로 최신 뉴스를 받아 Telegram 채널에 전송하는 봇임.

## 환경 설정

### .env 파일  
민감 정보 저장. 예시:
```dotenv
BOT_TOKEN=your_telegram_bot_token
NAVER_CLIENT_ID=your_naver_client_id
NAVER_CLIENT_SECRET=your_naver_client_secret
```
※ `.env` 파일은 버전 관리에서 제외할 것.

### config.yaml 파일  
민감하지 않은 운영 설정 저장. 예시:
```yaml
telegram:
  chat_ids:
    - "186**" # telegram chat id 
  delay_seconds: 1

news:
  query_list: # 검색어 리스트
    - "컴투스 | 엑스플라 | XPLA | 컴투스홀딩스 | 소울스트라이크 | 컴투스플랫폼"
  time_window_hours: 24 # 최근 몇 시간 동안의 뉴스를 가져올지
  pull_interval_seconds: 120 # 뉴스를 주기적으로 가져오는 간격
```

## 실행 방법

1. 의존성 설치
   ```bash
   go mod tidy
   ```

2. 환경 설정 파일 작성  
   프로젝트 루트에 `.env`와 `config.yaml` 파일을 위 예시대로 생성.

3. 빌드 및 실행
   ```bash
   go build -o newsbot
   ./newsbot
   ```

프로젝트 실행되면 주기적으로 네이버 뉴스 API 호출, 새 뉴스 발견 시 Telegram으로 전송함.