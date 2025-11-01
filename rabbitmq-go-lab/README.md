# RabbitMQ Go Lab (DLQ/Retry/Confirms + Observability)

## 빠른 시작
```bash
# 1) 컨테이너
docker compose up -d

# 2) 토폴로지 선언
make setup-topology

# 3) 컨슈머 실행
make run-consumer

# 4) 퍼블리시 (정상)
make publish

# 5) 실패 유도 퍼블리시
go run ./cmd/publisher --count 3 --key demo.info --fail

# 6) DLQ → 메인 재게시
make republish-dlq
```

- RabbitMQ 관리 콘솔: http://localhost:15672 (guest/guest)
- Grafana: http://localhost:3000 (admin / admin)
- Prometheus: http://localhost:9090
- 간이 모니터 UI: `go run ./cmd/monitor-ui` → http://localhost:8080

## 토폴로지
- Exchange
  - `app.events`(direct): 정상 라우팅
  - `app.events.retry`(direct): 재시도 라우팅
  - `app.events.dlx`(fanout): DLQ 라우팅
- Queues
  - `app.events.main` (DLX=app.events.retry)
  - `app.events.retry` (TTL=${RETRY_TTL_MS}ms, DLX=app.events)
  - `app.events.dlq` (DLX 바인딩)

## 재시도 & DLQ 로직
- 실패 시 `Nack(requeue=false)` → main의 DLX인 `app.events.retry`로 이동
- `retry` 큐 TTL 후 `app.events`로 다시 유입(재시도)
- `x-death.count`가 `${MAX_RETRIES}` 이상이면 DLX로 직접 게시하여 `DLQ`로 격리

## SpecKit 명령어에 매핑 (예시)
- `spec task setup:topology` → `make setup-topology`
- `spec task publish` → `make publish` 또는 `go run ./cmd/publisher --count 10 --key demo.info`
- `spec task republish-dlq` → `make republish-dlq`
- `spec service consumer` → `make run-consumer`
- `spec service monitor-ui` → `go run ./cmd/monitor-ui`

> 실제 SpecKit 사용 시, 위 명령을 SpecKit의 task 정의로 연결하세요.

## 학습 포인트 체크리스트
- [x] Publisher Confirms로 영속성/확인
- [x] Qos(pull pressure) 세팅
- [x] DLX/Retry 파이프라인(x-death 기반 카운팅)
- [x] DLQ 재처리(Admin)
- [x] Prometheus/Grafana 대시보드

## Go 버전
- `go 1.25` (환경에 맞춰 변경 가능)
