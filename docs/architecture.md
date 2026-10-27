# Architecture and operations

`mesh.Worker` is a composable processing pipeline. It rejects already-observed IDs, executes the handler behind a circuit breaker, and retries transient failures within a caller-controlled context. Policies are values, so applications can tune them without coupling business code to a broker SDK.

## Production topology

Attach Kafka, NATS JetStream, or SQS at the ingress boundary. Store idempotency keys in Redis with an expiry longer than the broker's redelivery horizon. Acknowledge only after the handler commits its durable side effect. Exhausted messages move to a dead-letter stream with attempt count and correlation ID.

Measure delivery lag, handler latency, retry count, open-circuit time, duplicate rate, and dead-letter volume. Alert on consumer lag and sustained circuit-open states rather than individual retries.

## Twelve-factor mapping

| Factor | Implementation |
|---|---|
| Codebase | One deployable worker codebase |
| Dependencies | Go modules; broker adapters are explicit packages |
| Config | Retry, broker, and port settings are environment-owned |
| Backing services | Broker, Redis, and handlers are attached resources |
| Build, release, run | CI test -> immutable image -> environment-configured process |
| Processes | Workers are disposable; durable offsets and dedupe keys live externally |
| Port binding | Operational HTTP endpoint binds `$PORT` |
| Concurrency | Add partitions and worker processes without changing code |
| Disposability | Context cancellation, bounded retries, fast startup |
| Dev/prod parity | Use the same worker image; swap resource URLs only |
| Logs | Structured events stream to stdout |
| Admin processes | Replay and dead-letter drains run from the same release image |

The memory adapter is deliberately marked local-only; it is not represented as distributed deduplication.
