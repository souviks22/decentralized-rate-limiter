# Decentralized Rate Limiter

[![Go Report Card](https://goreportcard.com/badge/github.com/souviks22/decentralized-rate-limiter)](https://goreportcard.com/report/github.com/souviks22/decentralized-rate-limiter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Made with Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://golang.org)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/souviks22/decentralized-rate-limiter/issues)


> **A decentralized rate limiter that continues to function under network partitions, built to study why centralized Redis-based designs break at scale.**

Most production rate limiters assume a **central truth** — a single Redis instance, a leader, or a globally consistent counter.

This project assumes the opposite:
**coordination is expensive, failures are normal, and partitions happen.**

Each node enforces rate limits **locally**, never blocking on global state, and reconciles usage **eventually** using CRDTs and peer-to-peer gossip.

**Design goal**

> *Never block locally, never trust globally — and still converge.*

This design targets infra teams operating across regions where availability and latency matter more than strict global enforcement.

---

## Why This Exists (and What Breaks in Practice)

Centralized rate limiting (Redis, API gateways, single coordinators) works well — **until it doesn’t**:

* cross-region latency leaks into the hot path,
* a single dependency becomes a blast radius,
* failure handling turns into policy ambiguity (“should we allow or deny?”).

In large distributed systems, **availability often matters more than perfect precision**.

This project explores a different trade-off:

* allow **temporary divergence**,
* guarantee **eventual convergence**,
* keep the request path **fast and local**.

---

## The Core Idea (High-Level)

Every node is a full authority for rate limiting.

```
Client → Any Node → Local Decision → Async Reconciliation
```

* No global coordinator
* No synchronous cross-node calls
* No blocking on consensus

Synchronization happens **off the hot path**.

If the node is alive, it answers.

---

## How It Works (End-to-End)

```
                           ┌──────────────────────────┐
                           │      Client Request      │
                           └────────────┬─────────────┘
                                        │
                                        ▼
                         ┌────────────────────────────────┐
                         │        Peer Node (A)           │
                         │ ────────────────────────────── │
                         │  1. Receive userID request     │
                         │  2. Check in-memory LRU        │
                         │  3. If miss, load from disk    │
                         │  4. TokenBucket.consume()      │
                         │  5. Record CRDT delta          │
                         └────────────┬───────────────────┘
                                      │
        ┌─────────────────────────────┼────────────────────────────┐
        ▼                             ▼                            ▼
┌────────────────┐         ┌──────────────────────┐        ┌────────────────────┐
│ In-Memory LRU  │◄───────▶│ Disk Storage (/data) │        │  CRDT Delta Cache  │
│  Token Buckets │         └──────────────────────┘        └────────▲───────────┘
└────▲───────────┘                                           ┌──────┴────────┐
     │                                                       │ libp2p Gossip │
     │                                                       └──────┬────────┘
     │                                                              │
     │     Broadcast deltas (every 100ms or 100 entries)            │
     └──────────────────────────────────────────────────────────────┘
                                                                    │
                                                 ┌──────────────────┴─────────────────┐
                                                 ▼                                    ▼
                                     ┌─────────────────────┐                 ┌─────────────────────┐
                                     │    Peer Node B      │◄───────────────▶│     Peer Node C     │
                                     │  (Same architecture)│     P2P Sync    │  (Same architecture)│
                                     └─────────────────────┘                 └─────────────────────┘
```

### Critical invariant

> **The request path never waits for gossip.**

Local decisions are final *for that node*.

---

## Design Components & Trade-offs

### Token Bucket (Local Authority)

Each user is governed by a standard token bucket:

* burst capacity,
* steady refill rate.

Why keep this simple:

* constant-time decisions,
* predictable latency,
* easy reasoning under load.

No hidden concurrency tricks. Thread safety is explicit.

---

### CRDT Deltas (Eventual Global Convergence)

Nodes exchange **deltas**, not full state.

Why deltas:

* small payloads,
* less merge ambiguity,
* faster convergence.

CRDT properties:

* commutative
* idempotent
* monotonic

This guarantees convergence even under:

* message loss,
* duplication,
* reordering.

**Exact precision is not promised.
Bounded divergence is.**

---

### libp2p Gossip (Coordination Without Leaders)

There is:

* no leader,
* no coordinator,
* no central broker.

Nodes discover peers and exchange deltas via libp2p gossip.

Failures are treated as routine:

* if a node disappears, others continue,
* when it returns, state reconverges.

---

### LRU + Disk (Scaling Beyond Memory)

Keeping all users in memory doesn’t scale.

So the system:

* keeps hot buckets in an **in-memory LRU**,
* evicts cold buckets to **disk**,
* reloads lazily on access.

This keeps:

* memory bounded,
* hot paths fast,
* cold users cheap.

Durability is pragmatic, not transactional.

---

## Performance Snapshot (3-node mesh)

| Metric                 | Observation    |
| ---------------------- | -------------- |
| Throughput per node    | ~3,000 req/sec |
| p99 request latency    | ~2 ms          |
| p99 gossip convergence | ~2 ms          |
| Gossip payload size    | ~3 KB          |

Interpretation:

* request latency is dominated by **local execution**,
* coordination cost is **asynchronous and amortized**.

---

## Failure Semantics (Explicit by Design)

This system chooses **availability over strict correctness**.

* **Network partitions** → nodes operate independently
* **Node crashes** → local state lost, global state reconverges
* **Delayed gossip** → temporary over-allowing possible

Observed behavior:

* ~15% bounded over-acceptance at 3 nodes
* grows roughly linearly with node count

This is acceptable for:

* abuse mitigation,
* fairness control,
* soft enforcement.

It is **not acceptable** for strict accounting.

This complexity is the cost paid to remove a global coordinator from the hot path.

---

## Design Walkthrough (Optional Deep Dive)

For a longer-form architectural walkthrough and design rationale, see:
[**High-Level Design of a Decentralized Rate Limiter**](https://medium.com/@souviksarkar2k3/high-level-design-of-a-decentralized-rate-limiter-1bcc33154ce9)

---

## When *Not* to Use This

This design is **not** a good fit if:

* every request must respect a single global counter,
* over-allowing is unacceptable (e.g., billing),
* centralized infrastructure is cheap and reliable for you.

In those cases, a Redis-backed or coordinator-based design is simpler and safer.

---

## Example Usage

```go
limiter := drl.NewRateLimiter(10, 1) // capacity, refill rate

if limiter.AllowRequest("user-123") {
    // request proceeds
} else {
    // rate limited locally
}
```

The API stays intentionally boring.
The complexity lives inside.

---

## What This Project Is (and Isn’t)

This project is **not** about replacing Redis.

It’s about answering a harder question:

> *What does rate limiting look like when the system itself refuses to be centralized?*

If that question matters in your environment, this design might be useful.



