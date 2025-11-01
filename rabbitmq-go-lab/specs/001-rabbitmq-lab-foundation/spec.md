# Feature Specification: RabbitMQ Lab Foundation

**Feature Branch**: `001-rabbitmq-lab-foundation`
**Created**: 2025-11-01
**Status**: Draft
**Input**: User description: "Build a comprehensive RabbitMQ learning lab with DLQ/Retry patterns, Publisher Confirms, and Observability using Prometheus and Grafana"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Message Publishing with Reliability (Priority: P1)

As a developer learning RabbitMQ, I want to publish messages with confirmation that they were successfully received by the broker, so that I can understand how to build reliable message producers in production systems.

**Why this priority**: Message publishing is the foundation of any messaging system. Without reliable publishing, messages can be lost silently. This is the most critical learning objective for production-ready systems.

**Independent Test**: Can be fully tested by running a publisher that sends messages with confirms enabled, observing the confirmation callbacks, and verifying that published messages are persisted in queues. Delivers immediate value by demonstrating the reliability guarantee mechanism.

**Acceptance Scenarios**:

1. **Given** RabbitMQ is running with topology configured, **When** a developer publishes messages with confirms enabled, **Then** each published message receives a confirmation (ack/nack) from the broker
2. **Given** messages are being published, **When** a message is successfully routed to a queue, **Then** the publisher receives a positive acknowledgment
3. **Given** messages are being published, **When** the broker cannot persist a message, **Then** the publisher receives a negative acknowledgment and can handle the failure
4. **Given** multiple messages are published in sequence, **When** viewing queue statistics, **Then** all confirmed messages are visible in the target queue

---

### User Story 2 - Message Consumption with Retry Logic (Priority: P2)

As a developer learning RabbitMQ, I want to consume messages with automatic retry capability when processing fails, so that I can build resilient consumer applications that handle transient failures gracefully.

**Why this priority**: Once messages can be published reliably, the next critical skill is consuming them with fault tolerance. Retry mechanisms are essential for production systems where transient failures are common.

**Independent Test**: Can be fully tested by running a consumer that processes messages, simulating failures for specific messages, and verifying that failed messages are automatically retried according to configured retry limits. Delivers value by demonstrating production-grade error handling.

**Acceptance Scenarios**:

1. **Given** messages are in the main queue, **When** a consumer processes a message successfully, **Then** the message is acknowledged and removed from the queue
2. **Given** a message processing fails, **When** the consumer nacks the message without requeue, **Then** the message is moved to the retry queue with TTL
3. **Given** a message is in the retry queue, **When** the TTL expires, **Then** the message is automatically republished to the main queue for retry
4. **Given** a message has been retried multiple times, **When** the retry count exceeds the configured maximum, **Then** the message is routed to the dead letter queue instead of being retried again
5. **Given** a consumer is processing messages, **When** using QoS (prefetch) settings, **Then** the consumer only fetches a limited number of unacknowledged messages at a time (back pressure control)

---

### User Story 3 - Dead Letter Queue Management (Priority: P3)

As a developer learning RabbitMQ, I want to inspect and reprocess messages that have exhausted all retries, so that I can understand how to handle poison messages and implement manual recovery procedures.

**Why this priority**: While retry logic handles most transient failures, some messages will always fail (poison messages, bugs in processing logic). Understanding DLQ management is essential for production operations but can be learned after the core publish/consume flow.

**Independent Test**: Can be fully tested by sending messages that will intentionally fail all retries, verifying they land in the DLQ, manually inspecting them, and republishing them back to the main queue after fixing the underlying issue. Delivers value by demonstrating operational recovery procedures.

**Acceptance Scenarios**:

1. **Given** messages have failed and landed in the DLQ, **When** an operator views the DLQ, **Then** all failed messages are visible with their failure metadata (death count, original routing key, failure reason)
2. **Given** messages are in the DLQ, **When** an operator decides to reprocess them, **Then** the operator can republish messages from DLQ back to the main exchange
3. **Given** DLQ messages are being republished, **When** the republish completes, **Then** the messages appear in the main queue and the DLQ is emptied
4. **Given** poison messages in the DLQ, **When** reviewing failure patterns, **Then** the operator can identify common failure reasons and decide on appropriate action (fix and retry, or discard)

---

### User Story 4 - Observability and Monitoring (Priority: P4)

As a developer learning RabbitMQ, I want to monitor message flow metrics and system health through visual dashboards, so that I can understand how to operate and troubleshoot messaging systems in production.

**Why this priority**: Observability is critical for production systems, but can be learned after understanding the core messaging patterns. It enhances the learning experience but isn't required for basic functionality.

**Independent Test**: Can be fully tested by publishing and consuming messages while viewing real-time metrics in Grafana dashboards and Prometheus queries. Delivers value by visualizing the entire message lifecycle and system health.

**Acceptance Scenarios**:

1. **Given** the observability stack is running, **When** messages are published and consumed, **Then** metrics for publish rate, consume rate, queue depth, and retry counts are visible in Prometheus
2. **Given** metrics are being collected, **When** viewing the Grafana dashboard, **Then** visual graphs show message flow through each queue (main, retry, DLQ) over time
3. **Given** the system is processing messages, **When** viewing the monitoring UI, **Then** real-time statistics show current queue depths, consumer counts, and message rates
4. **Given** failures occur, **When** reviewing the dashboards, **Then** failure rates and DLQ growth are clearly visible for troubleshooting
5. **Given** the RabbitMQ management console is accessed, **When** viewing topology, **Then** all exchanges, queues, and bindings are visible with their configurations

---

### Edge Cases

- What happens when RabbitMQ restarts while messages are in the retry queue (TTL persistence)?
- How does the system handle a message that cannot be serialized/deserialized during consumption?
- What happens when the retry queue is full or reaches memory limits?
- How does the system behave when a consumer crashes mid-processing (message redelivery)?
- What happens when DLQ messages are republished but the underlying bug hasn't been fixed (infinite loop prevention)?
- How does back pressure (QoS prefetch) affect throughput when consumer processing is slow?
- What happens when multiple consumers are competing for messages from the same queue?
- How does the system handle network partitions between publisher/consumer and RabbitMQ?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a topology setup that creates all required exchanges (main events, retry, DLX) and queues (main, retry, DLQ) with proper bindings
- **FR-002**: System MUST support publisher confirms to guarantee message delivery acknowledgment from the broker
- **FR-003**: Publishers MUST be able to send messages with routing keys and verify successful routing to target queues
- **FR-004**: System MUST implement consumer QoS (prefetch) settings to enable back pressure control
- **FR-005**: Consumers MUST acknowledge or nack messages based on processing success/failure
- **FR-006**: System MUST route nacked messages (requeue=false) to the retry exchange via DLX mechanism
- **FR-007**: Retry queue MUST enforce a configurable TTL (time-to-live) before republishing messages to the main queue
- **FR-008**: System MUST track retry attempts using the x-death header to count how many times a message has been retried
- **FR-009**: System MUST route messages that exceed the maximum retry count directly to the DLQ instead of retrying again
- **FR-010**: System MUST provide an administrative function to republish messages from DLQ back to the main exchange
- **FR-011**: System MUST expose metrics to Prometheus including: publish rate, consume rate, queue depths, retry counts, DLQ size, and failure rates
- **FR-012**: System MUST provide Grafana dashboards that visualize message flow through all queues over time
- **FR-013**: System MUST provide a monitoring UI that shows real-time queue statistics and message rates
- **FR-014**: System MUST provide RabbitMQ management console access for topology inspection
- **FR-015**: System MUST use Docker Compose to orchestrate all required services (RabbitMQ, Prometheus, Grafana)
- **FR-016**: System MUST provide runnable commands for common operations: topology setup, publisher execution, consumer execution, DLQ republish, monitoring UI launch

### Key Entities

- **Message**: Represents data flowing through the system with properties including body content, routing key, headers (including x-death for retry tracking), and delivery metadata
- **Exchange**: Routes messages to queues based on routing rules; includes main events exchange (direct), retry exchange (direct), and DLX exchange (fanout)
- **Queue**: Stores messages for consumption; includes main queue (with DLX to retry), retry queue (with TTL and DLX to main), and DLQ (dead letter queue for final failures)
- **Binding**: Connects exchanges to queues with routing key patterns
- **Publisher**: Component that sends messages to exchanges with confirms enabled
- **Consumer**: Component that retrieves and processes messages from queues with acknowledgment/nack capability
- **Metric**: Measurement data collected about system behavior including counters (published, consumed, retried, dead-lettered) and gauges (queue depth, consumer count)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can publish 100 messages with publisher confirms and receive acknowledgment for all successfully routed messages within seconds
- **SC-002**: When a consumer fails to process a message, the message is automatically retried according to the configured retry policy without manual intervention
- **SC-003**: Messages that fail processing are isolated in the DLQ after exhausting retries, preventing infinite retry loops
- **SC-004**: Developers can view real-time metrics showing message flow rates, queue depths, and failure counts through Grafana dashboards
- **SC-005**: The complete topology (exchanges, queues, bindings) can be set up from scratch in under 30 seconds using provided commands
- **SC-006**: All core messaging patterns (publish, consume, retry, DLQ, republish) can be demonstrated and tested using provided example commands
- **SC-007**: System handles concurrent publishing and consuming with configurable back pressure (QoS) to prevent consumer overload
- **SC-008**: Developers can observe the complete message lifecycle from publishing through retries to either successful consumption or DLQ routing
- **SC-009**: All infrastructure dependencies (RabbitMQ, Prometheus, Grafana) start successfully using a single Docker Compose command
- **SC-010**: Monitoring tools (management console, Grafana, Prometheus, custom UI) provide visibility into system state without requiring code changes
