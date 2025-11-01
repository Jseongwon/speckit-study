# Implementation Plan: RabbitMQ Lab Foundation

**Branch**: `001-rabbitmq-lab-foundation` | **Date**: 2025-11-01 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-rabbitmq-lab-foundation/spec.md`

## Summary

This learning lab provides hands-on experience with production-grade RabbitMQ patterns including Dead Letter Queues (DLQ), automatic retry mechanisms, publisher confirms, and comprehensive observability. The system demonstrates reliable message publishing with confirmation, fault-tolerant consumption with configurable retry logic, dead letter queue management for poison messages, and real-time monitoring through Prometheus and Grafana dashboards.

**Primary Goal**: Enable developers to learn and practice essential RabbitMQ patterns for building resilient, observable messaging systems.

**Technical Approach**: Go-based publisher/consumer applications using the official AMQP 0-9-1 client library, containerized infrastructure with Docker Compose, and Prometheus/Grafana stack for observability.

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**:
- `github.com/rabbitmq/amqp091-go` v1.10.0 (official AMQP 0-9-1 client)
- `github.com/joho/godotenv` v1.5.1 (environment configuration)

**Storage**: RabbitMQ 3.13 with persistent message storage (durable queues)

**Testing**: Manual integration testing through executable commands demonstrating each pattern (publish, consume, retry, DLQ recovery)

**Target Platform**:
- Development: macOS/Linux/Windows with Docker support
- Runtime: Docker Compose orchestration (RabbitMQ, Prometheus, Grafana containers)

**Project Type**: Learning lab with multiple CLI tools (publisher, consumer, admin, monitor-ui)

**Performance Goals**:
- Topology setup: < 30 seconds
- Publisher confirms: Acknowledgment within seconds for 100 messages
- Consumer throughput: Configurable via QoS prefetch (default: 10 concurrent messages)
- Retry latency: Configurable TTL (default: 10 seconds)
- Metrics refresh: Real-time updates in Grafana (< 15 second scrape interval)

**Constraints**:
- Local development environment (not production-hardened)
- Single RabbitMQ node (no clustering)
- Simple authentication (guest/guest for learning purposes)
- Prefetch limit: 10 messages (back pressure control)
- Max retries: 3 attempts before DLQ routing
- Retry TTL: 10 seconds (configurable)

**Scale/Scope**:
- 4 user stories (P1-P4: Publish → Consume → DLQ → Observability)
- 16 functional requirements
- 4 CLI commands (publisher, consumer, admin, monitor-ui)
- 3 exchanges, 3 queues, multiple bindings
- 3 Docker services (RabbitMQ, Prometheus, Grafana)
- Learning lab scope (not production scale)

## Constitution Check

*No constitution file found - skipping constitutional gates*

## Project Structure

### Documentation (this feature)

```text
specs/001-rabbitmq-lab-foundation/
├── plan.md              # This file (/speckit.plan command output)
├── spec.md              # Feature specification (created)
├── research.md          # Phase 0 output (to be created by /speckit.plan)
├── data-model.md        # Phase 1 output (to be created by /speckit.plan)
├── quickstart.md        # Phase 1 output (to be created by /speckit.plan)
├── contracts/           # Phase 1 output (to be created by /speckit.plan)
├── checklists/          # Quality checklists
│   └── requirements.md  # Spec quality validation (created)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/                        # Executable commands
├── publisher/
│   └── main.go            # Message publisher with confirms
├── consumer/
│   └── main.go            # Message consumer with retry logic
├── admin/
│   └── main.go            # Admin operations (topology setup, DLQ republish)
└── monitor-ui/
    └── main.go            # Simple monitoring web UI

internal/                   # Internal packages (not importable externally)
├── config/
│   └── config.go          # Environment configuration loading
├── rabbit/
│   ├── connect.go         # RabbitMQ connection management
│   └── setup.go           # Topology declaration (exchanges, queues, bindings)
└── types/
    └── event.go           # Message type definitions

observability/              # Monitoring infrastructure configuration
├── prometheus/
│   └── prometheus.yml     # Prometheus scrape configuration
└── grafana/
    ├── dashboards/
    │   └── rabbitmq_lab.json  # RabbitMQ metrics dashboard
    └── provisioning/
        ├── datasources/
        │   └── datasource.yml  # Prometheus datasource
        └── dashboards/
            └── dashboards.yml  # Dashboard provisioning

# Root configuration
docker-compose.yml          # Service orchestration (RabbitMQ, Prometheus, Grafana)
.env                        # Environment variables (topology config, credentials)
Makefile                    # Command shortcuts (compose-up, setup-topology, etc.)
go.mod                      # Go module definition
go.sum                      # Go dependency checksums
README.md                   # Quick start guide
```

**Structure Decision**: Single Go module with multiple command-line tools (cmd/) and shared internal packages (internal/). This structure supports:
- **Separation of concerns**: Each CLI tool has independent entry point
- **Code reuse**: Common functionality (connection, config, types) in internal/
- **Observability infrastructure**: Declarative configuration files separate from application code
- **Development workflow**: Makefile provides simple command shortcuts for common operations

## Architecture Overview

### Topology Design

The RabbitMQ topology implements a DLQ/Retry pattern with three exchanges and three queues:

**Exchanges**:
1. **`app.events`** (direct): Main routing exchange for incoming messages
2. **`app.events.retry`** (direct): Retry routing exchange for failed messages
3. **`app.events.dlx`** (fanout): Dead Letter Exchange for messages exceeding retry limit

**Queues**:
1. **`app.events.main`**: Main processing queue
   - DLX: `app.events.retry` (failed messages route here)
   - Bound to: `app.events` exchange

2. **`app.events.retry`**: Temporary retry queue with TTL
   - TTL: Configurable (default 10 seconds)
   - DLX: `app.events` (messages re-route to main after TTL)
   - Bound to: `app.events.retry` exchange

3. **`app.events.dlq`**: Dead Letter Queue (poison message isolation)
   - No DLX (terminal state)
   - Bound to: `app.events.dlx` exchange

**Message Flow**:
```
1. Publish → app.events exchange → app.events.main queue
2. Consumer processes:
   - Success: Ack → message removed
   - Failure: Nack(requeue=false) → DLX routing
3. Failed message → app.events.retry exchange → app.events.retry queue
4. After TTL expires → app.events exchange → app.events.main queue (retry)
5. Retry count check (x-death header):
   - count < MAX_RETRIES: Continue retry cycle (step 2-4)
   - count >= MAX_RETRIES: Direct publish to app.events.dlx → app.events.dlq
6. Manual recovery: Admin republishes DLQ → app.events → app.events.main
```

### Component Architecture

**Publisher** (`cmd/publisher/main.go`):
- Establishes connection with publisher confirms enabled
- Publishes messages to `app.events` exchange with routing key
- Waits for confirmation (ack/nack) from broker
- Supports forced failure flag for testing retry/DLQ flow

**Consumer** (`cmd/consumer/main.go`):
- Establishes connection with QoS prefetch=10 (back pressure)
- Consumes from `app.events.main` queue
- Processes messages with simulated failure logic
- Checks `x-death` header to track retry count
- Routes to DLQ if retry count exceeds MAX_RETRIES
- Otherwise: Nack(requeue=false) to trigger retry via DLX

**Admin** (`cmd/admin/main.go`):
- **Topology Setup**: Declares all exchanges, queues, and bindings
- **DLQ Republish**: Consumes messages from DLQ and republishes to main exchange

**Monitor UI** (`cmd/monitor-ui/main.go`):
- Simple HTTP server with `/publish` endpoint
- Web interface for triggering message publishing
- Real-time queue statistics display

### Observability Stack

**RabbitMQ Prometheus Plugin** (port 15692):
- Exposes metrics: queue depth, message rates, consumer count
- Scraped by Prometheus every 15 seconds

**Prometheus** (port 9090):
- Scrapes RabbitMQ metrics
- Time-series database for historical analysis
- Query interface for ad-hoc metric exploration

**Grafana** (port 3000):
- Pre-configured dashboard: `rabbitmq_lab.json`
- Visualizations: Queue depths over time, message rates, retry/DLQ trends
- Datasource: Prometheus (auto-provisioned)

**RabbitMQ Management Console** (port 15672):
- Web UI for topology inspection
- Queue/exchange browsing
- Message statistics and monitoring

## Reliability Patterns

### Publisher Confirms

**Pattern**: Synchronous confirmation of message persistence
- Publisher enables confirm mode on channel
- Broker sends ack when message is persisted to disk/quorum
- Publisher blocks waiting for confirm before considering publish successful
- Guarantees: Message durably stored or publisher knows it failed

**Implementation**:
- `channel.Confirm(false)` - enable confirms
- `channel.PublishWithDeferredConfirm()` - publish and get confirmation
- Handle ack/nack to detect failures

### Consumer Acknowledgment & QoS

**Pattern**: Manual acknowledgment with prefetch limit
- QoS prefetch=10: Limit unacknowledged messages to 10 (back pressure)
- Manual ack on successful processing
- Manual nack(requeue=false) on failure to trigger DLX routing
- Guarantees: Messages not lost if consumer crashes (redelivered)

**Implementation**:
- `channel.Qos(10, 0, false)` - set prefetch
- `delivery.Ack(false)` - successful processing
- `delivery.Nack(false, false)` - failed processing, trigger DLX

### Retry with Exponential Backoff (via TTL)

**Pattern**: Dead Letter Exchange with TTL for delayed retry
- Failed messages nacked without requeue → routed to retry exchange via DLX
- Retry queue has TTL (10 seconds default)
- After TTL, message expires → DLX routes back to main exchange
- Retry count tracked via `x-death` header (RabbitMQ automatic)

**Guarantees**: Transient failures handled automatically, permanent failures isolated

**Implementation**:
- Main queue: `x-dead-letter-exchange: app.events.retry`
- Retry queue: `x-message-ttl: 10000`, `x-dead-letter-exchange: app.events`
- Consumer checks `x-death[0].count >= MAX_RETRIES` → publish to DLX

### Dead Letter Queue Management

**Pattern**: Terminal queue for poison messages
- Messages exceeding MAX_RETRIES published directly to DLX exchange
- DLQ queue has no DLX (terminal state)
- Manual intervention required: Admin inspects and republishes after fix

**Guarantees**: Poison messages don't block queue, can be recovered manually

**Implementation**:
- Consumer: If retry count >= MAX_RETRIES, publish to `app.events.dlx`
- Admin tool: Consume from DLQ, republish to `app.events` exchange

## Execution Flow & Commands

### Initial Setup

```bash
# 1. Start infrastructure
make compose-up
# → Starts RabbitMQ (5672, 15672, 15692), Prometheus (9090), Grafana (3000)

# 2. Declare topology (exchanges, queues, bindings)
make setup-topology
# → Runs: go run ./cmd/admin --setup
# → Creates: app.events, app.events.retry, app.events.dlx exchanges
# → Creates: app.events.main, app.events.retry, app.events.dlq queues
# → Configures: DLX routing, TTL, bindings
```

### User Story 1: Reliable Publishing (P1)

```bash
# Publish 10 messages with confirms
make publish
# → Runs: go run ./cmd/publisher --count 10 --key demo.info
# → Publishes to app.events exchange
# → Waits for broker confirmation (ack/nack)
# → Logs confirmation status for each message

# Verify in RabbitMQ console
# → Open: http://localhost:15672 (guest/guest)
# → Check: app.events.main queue should show 10 messages
```

### User Story 2: Consumption with Retry (P2)

```bash
# Start consumer (separate terminal)
make run-consumer
# → Runs: go run ./cmd/consumer
# → Consumes from app.events.main with prefetch=10
# → Processes messages (logs success/failure)
# → Nacks failures → triggers retry flow

# Publish messages with forced failures
go run ./cmd/publisher --count 3 --key demo.info --fail
# → Messages intentionally fail processing
# → Consumer nacks → messages route to retry queue
# → After TTL (10s), messages return to main queue
# → Retry up to MAX_RETRIES (3) times
# → After 3 retries → route to DLQ

# Observe retry flow
# → RabbitMQ console: Watch message counts in retry queue
# → Consumer logs: See retry attempts with x-death count
```

### User Story 3: DLQ Management (P3)

```bash
# Inspect DLQ
# → RabbitMQ console: Check app.events.dlq queue
# → Should contain messages that failed 3+ times
# → Click "Get Messages" to view message details
# → x-death header shows retry history

# Republish from DLQ (after fixing issue)
make republish-dlq
# → Runs: go run ./cmd/admin --republish-dlq --limit 100
# → Consumes up to 100 messages from DLQ
# → Republishes to app.events exchange
# → Messages re-enter normal processing flow
# → DLQ should be empty after successful republish
```

### User Story 4: Observability (P4)

```bash
# View Prometheus metrics
# → Open: http://localhost:9090
# → Query examples:
#   - rabbitmq_queue_messages{queue="app.events.main"}
#   - rate(rabbitmq_queue_messages_published_total[1m])
#   - rabbitmq_queue_consumers

# View Grafana dashboard
# → Open: http://localhost:3000 (admin/admin)
# → Navigate: Dashboards → RabbitMQ Lab
# → Panels show:
#   - Queue depths over time (main, retry, DLQ)
#   - Message publish/consume rates
#   - Consumer count
#   - Retry and DLQ accumulation trends

# Run monitoring UI
go run ./cmd/monitor-ui
# → Starts web server: http://localhost:8080
# → UI provides:
#   - Publish button with forceFail option
#   - Real-time queue statistics
#   - Quick visual feedback on message flow
```

## Risk Assessment & Mitigation

### Risk 1: Large Message Volumes Overwhelming Retry Queue

**Scenario**: Sudden failure causes thousands of messages to accumulate in retry queue
**Impact**: Memory pressure on RabbitMQ, delayed processing, potential node instability

**Mitigation**:
- Set `x-max-length` on retry queue to cap size (e.g., 10,000 messages)
- Implement retry queue overflow policy: Drop oldest or route to DLQ
- Monitor retry queue depth with Grafana alerts (threshold: > 5,000)
- Circuit breaker pattern: Stop publishing if retry queue exceeds threshold
- Learning opportunity: Add optional task to implement queue length limits

### Risk 2: Retry Storm (Cascading Failures)

**Scenario**: Bug in consumer causes all messages to fail → retry → fail → retry (amplification)
**Impact**: RabbitMQ CPU/network saturation, exponential message growth, service degradation

**Mitigation**:
- Enforce strict MAX_RETRIES limit (3 attempts → DLQ)
- TTL provides natural rate limiting (10 second delay between retries)
- DLQ isolates poison messages preventing infinite loops
- Monitor DLQ growth rate → alerts for sudden spikes
- Consumer idempotency: Log message IDs to detect duplicate processing
- Learning opportunity: Demonstrate retry storm scenario and recovery

### Risk 3: Duplicate Message Processing

**Scenario**: Consumer crashes after processing but before ack → message redelivered
**Impact**: Duplicate effects (e.g., double-charging, duplicate records)

**Mitigation**:
- Consumer implements idempotency checks (log message ID before processing)
- Use `delivery.Redelivered` flag to detect redeliveries
- Application-level deduplication (message ID tracking with TTL)
- Learning opportunity: Demonstrate redelivery scenario with consumer crash
- Document that exactly-once processing requires application-level tracking

### Risk 4: Connection Pool Exhaustion

**Scenario**: Too many publishers/consumers creating connections simultaneously
**Impact**: RabbitMQ connection limit reached, new connections rejected, service unavailable

**Mitigation**:
- Learning lab scope: Single publisher, single consumer (no pooling needed)
- Document connection best practices: Reuse connections, share channels
- RabbitMQ default connection limit: 1,000+ (sufficient for lab)
- If scaling: Implement connection pooling library
- Monitor connection count via RabbitMQ management console
- Learning opportunity: Optional advanced task to implement connection pooling

### Risk 5: Message Loss on RabbitMQ Restart

**Scenario**: RabbitMQ container restarts, messages in retry queue with TTL are lost
**Impact**: In-flight retries lost, potential data loss for unacknowledged messages

**Mitigation**:
- Durable queues: Messages persisted to disk (configured in setup)
- Publisher confirms: Ensure message persisted before considering sent
- Consumer manual ack: Message stays in queue until successfully processed
- Docker volume for RabbitMQ data persistence (optional enhancement)
- Learning opportunity: Demonstrate restart scenario and durability guarantees
- Trade-off discussion: TTL persistence vs. performance

## Definition of Done

### Topology Verification

- [ ] All exchanges exist: `app.events`, `app.events.retry`, `app.events.dlx`
- [ ] All queues exist: `app.events.main`, `app.events.retry`, `app.events.dlq`
- [ ] Main queue DLX correctly points to retry exchange
- [ ] Retry queue has TTL configured (10 seconds) and DLX points to main exchange
- [ ] DLQ has no DLX (terminal state)
- [ ] Bindings correct: Exchanges → Queues with proper routing keys
- [ ] Verification: `make setup-topology` completes without errors
- [ ] Verification: RabbitMQ console shows all topology elements

### Publish & Confirm Flow

- [ ] Publisher establishes connection successfully
- [ ] Publisher enables confirm mode on channel
- [ ] Messages published to `app.events` exchange with routing key
- [ ] Publisher receives ack/nack for each message
- [ ] Messages appear in `app.events.main` queue
- [ ] Verification: `make publish` logs confirmation for all 10 messages
- [ ] Verification: RabbitMQ console shows message count in main queue

### Consume & Retry Flow

- [ ] Consumer establishes connection with QoS prefetch=10
- [ ] Consumer processes messages successfully (ack)
- [ ] Consumer detects failures and nacks without requeue
- [ ] Failed messages route to retry queue via DLX
- [ ] Messages re-appear in main queue after TTL expires
- [ ] Consumer checks `x-death` header for retry count
- [ ] Messages exceeding MAX_RETRIES route to DLQ
- [ ] Verification: `go run ./cmd/publisher --fail` → messages retry → land in DLQ
- [ ] Verification: Consumer logs show retry count increasing

### DLQ Management

- [ ] Failed messages (3+ retries) appear in DLQ
- [ ] DLQ messages contain `x-death` header with retry history
- [ ] Admin tool can consume from DLQ
- [ ] Admin tool republishes messages to main exchange
- [ ] Republished messages re-enter processing flow
- [ ] DLQ empties after successful republish
- [ ] Verification: `make republish-dlq` moves messages from DLQ to main queue
- [ ] Verification: RabbitMQ console shows DLQ count = 0 after republish

### Observability & Monitoring

- [ ] RabbitMQ Prometheus plugin exposes metrics on port 15692
- [ ] Prometheus scrapes RabbitMQ metrics successfully
- [ ] Grafana datasource configured and connected to Prometheus
- [ ] Grafana dashboard displays queue depths, message rates
- [ ] RabbitMQ management console accessible (port 15672)
- [ ] Monitor UI runs and displays real-time statistics
- [ ] Verification: Grafana dashboard shows live metrics while publishing/consuming
- [ ] Verification: Prometheus query returns queue depth metrics
- [ ] Verification: Monitor UI `/publish` button triggers message publishing

### Infrastructure & Commands

- [ ] `make compose-up` starts all services (RabbitMQ, Prometheus, Grafana)
- [ ] All services healthy and accessible on configured ports
- [ ] `make setup-topology` declares topology successfully
- [ ] `make publish` publishes messages with confirms
- [ ] `make run-consumer` starts consumer successfully
- [ ] `make republish-dlq` republishes DLQ messages
- [ ] `go run ./cmd/monitor-ui` starts monitoring web UI
- [ ] Verification: All Makefile targets execute without errors
- [ ] Verification: README quickstart sequence completes successfully

### Code Quality & Documentation

- [ ] All Go code follows standard formatting (`go fmt`)
- [ ] Environment variables documented in `.env` with comments
- [ ] README provides quickstart with command examples
- [ ] Code includes comments explaining retry/DLQ logic
- [ ] Error handling implemented (connection failures, publish failures)
- [ ] Graceful shutdown for consumer (handle SIGTERM/SIGINT)
- [ ] Verification: `go build ./...` succeeds without warnings
- [ ] Verification: README instructions lead to successful demonstration

## Next Steps & Future Enhancements

### Phase 1: Core Patterns (Current Feature)

Implement the foundational DLQ/Retry patterns as specified in the current plan.

### Phase 2: Advanced Reliability (Future)

**Quorum Queues** (RabbitMQ 3.8+):
- Replace classic queues with quorum queues for better durability
- Benefits: Built-in replication, better data safety guarantees
- Trade-offs: Higher resource usage, learning cluster concepts
- Learning opportunity: Compare classic vs. quorum queue behavior

**Priority Queues**:
- Add message priority support (0-9 scale)
- High-priority messages processed before low-priority
- Use case: Critical alerts bypass normal processing queue
- Implementation: `x-max-priority` queue argument

**Delayed Message Exchange Plugin**:
- Alternative to TTL-based retry (more flexible)
- Specify exact delay per message (vs. queue-level TTL)
- Use case: Variable retry delays (exponential backoff)

### Phase 3: Advanced Observability (Future)

**OpenTelemetry Integration**:
- Instrument publisher/consumer with OTEL tracing
- Distributed tracing: Track message flow across components
- Span attributes: Message ID, routing key, retry count
- Trace backends: Jaeger, Zipkin, or Grafana Tempo
- Learning opportunity: Visualize message lifecycle end-to-end

**Custom Metrics**:
- Application-level metrics beyond RabbitMQ metrics
- Examples: Processing duration, business event counts, error types
- Prometheus client library integration
- Custom Grafana dashboard panels

**Alerting**:
- Prometheus alerting rules (DLQ growth, consumer lag)
- Alert routing: Email, Slack, PagerDuty
- Runbooks: Document response procedures

### Phase 4: Production Patterns (Future)

**Connection Pooling**:
- Reuse connections across multiple publishers
- Channel pooling for high-throughput scenarios
- Connection recovery on transient failures

**Consumer Scaling**:
- Horizontal scaling: Multiple consumer instances
- Work distribution: RabbitMQ round-robin delivery
- Consistent processing with competing consumers

**Message Schemas & Validation**:
- JSON Schema or Protocol Buffers for message format
- Schema validation before processing
- Backward compatibility strategies

**Security Hardening**:
- TLS encryption for connections
- User permissions and virtual hosts
- Secret management (Vault, AWS Secrets Manager)

## Implementation Roadmap

### Tasks Organization

Tasks will be generated using `/speckit.tasks` command and organized by user story priority:

**Phase 1: Setup** (Foundation)
- Docker Compose configuration
- Environment variable setup
- Makefile command definitions

**Phase 2: Foundational** (Blocking prerequisites)
- Go module initialization
- Internal packages (config, rabbit, types)
- Connection management
- Topology setup (admin tool)

**Phase 3: User Story 1 - Reliable Publishing (P1)**
- Publisher implementation with confirms
- Message type definitions
- Publish command integration
- Verification: Messages appear in queue with confirmations

**Phase 4: User Story 2 - Retry Logic (P2)**
- Consumer implementation with QoS
- Retry count tracking (x-death header)
- DLQ routing logic
- Consumer command integration
- Verification: Retry flow with TTL and DLQ routing

**Phase 5: User Story 3 - DLQ Management (P3)**
- Admin DLQ republish functionality
- Message inspection utilities
- Republish command integration
- Verification: DLQ recovery workflow

**Phase 6: User Story 4 - Observability (P4)**
- Prometheus configuration
- Grafana dashboard creation
- Monitor UI implementation
- Metrics validation
- Verification: End-to-end observability

**Phase 7: Polish**
- Documentation (README, code comments)
- Error handling improvements
- Graceful shutdown
- Final integration testing

### Execution Sequence

Refer to `tasks.md` (generated by `/speckit.tasks`) for detailed task breakdown with dependencies:

```bash
# After tasks.md generation:
# 1. Phase 1: Setup → make compose-up
# 2. Phase 2: Foundational → make setup-topology
# 3. Phase 3: Publishing → make publish (verify confirmations)
# 4. Phase 4: Consumption → make run-consumer (verify retry flow)
# 5. Phase 5: DLQ → make republish-dlq (verify recovery)
# 6. Phase 6: Observability → Grafana verification
# 7. Phase 7: Polish → Documentation and final testing
```

### Parallel Execution Opportunities

Within each phase, some tasks can be executed in parallel:
- **Setup phase**: Docker Compose, .env, Makefile are independent
- **Foundational phase**: config, types packages can be developed concurrently
- **Publishing/Consumer**: Publisher and consumer can be built in parallel (both depend on foundational)
- **Observability**: Prometheus config, Grafana dashboard, Monitor UI are independent

Detailed parallelization strategy will be provided in `tasks.md`.

## Assumptions & Constraints

### Assumptions

1. **Local Development Environment**: Users have Docker and Go 1.25+ installed
2. **Learning Context**: This is a lab environment, not production deployment
3. **Simple Authentication**: guest/guest credentials acceptable for local learning
4. **Single Node**: RabbitMQ clustering not required for learning objectives
5. **Manual Testing**: Integration testing via command execution (no automated test suite)
6. **Existing Implementation**: Project already has baseline code (plan documents current state)

### Constraints

1. **RabbitMQ Version**: 3.13-management (for Prometheus plugin support)
2. **Go Version**: 1.25 (as specified in go.mod)
3. **Network Ports**: 5672 (AMQP), 15672 (management), 15692 (Prometheus), 9090 (Prometheus), 3000 (Grafana)
4. **Retry Limit**: MAX_RETRIES = 3 (configurable via .env)
5. **TTL**: RETRY_TTL_MS = 10000 (10 seconds, configurable)
6. **Prefetch**: QoS prefetch = 10 (back pressure control)
7. **Message Persistence**: Durable queues only (no transient queues)
8. **Routing**: Direct exchanges for main/retry, fanout for DLX

### Out of Scope

- Production deployment (Kubernetes, cloud environments)
- RabbitMQ clustering and high availability
- Advanced security (TLS, user permissions, vhosts)
- Automated test suites (unit, integration, contract tests)
- Message schema validation (JSON Schema, Protobuf)
- Connection pooling and advanced performance optimization
- Consumer autoscaling
- Alert notification integrations (Slack, PagerDuty)

These may be addressed in future enhancements (see Next Steps section).
