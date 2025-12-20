# 🌐 Decentralized Rate Limiter

[![Go Report Card](https://goreportcard.com/badge/github.com/souviks22/decentralized-rate-limiter)](https://goreportcard.com/report/github.com/souviks22/decentralized-rate-limiter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Made with Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://golang.org)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/souviks22/decentralized-rate-limiter/issues)

> Most rate limiters assume a central truth.
> This one assumes failure.

This project explores how rate limiting behaves when **central coordination is expensive, unreliable, or undesirable** — across regions, failures, and partitions. Instead of enforcing a single global counter, each node makes **local decisions** and reconciles state **eventually**, using CRDTs and peer-to-peer gossip.

The goal is simple but strict:
**never block locally, never trust globally, and still converge.**

---

## Why This Exists

Traditional rate limiters (Redis, centralized gateways) work well — until:

* cross-region latency dominates the hot path,
* a single dependency becomes a blast radius,
* or failure handling turns into policy ambiguity.

In large distributed systems, **availability often matters more than precision**.
This rate limiter is designed for those environments.

Each node:

* limits requests **locally**,
* survives **network partitions**,
* and synchronizes state **without a leader**.

Precision is relaxed. Safety and continuity are not.

---

## High-Level Design

At a high level, every node is fully capable of enforcing limits on its own.

```
Client → Any Node → Local Decision → Eventual Reconciliation
```

No node blocks waiting for global state. Synchronization happens **off the hot path**.

---

## How It Works (End-to-End)

```
                           ┌──────────────────────────┐
                           │      Client Request      │
                           └────────────┬─────────────┘
                                        │
                                        ▼
                         ┌────────────────────────────────┐
                         │    Peer Node (e.g., Node A)    │
                         │ ────────────────────────────── │
                         │  1. Receive userID request     │
                         │  2. Check in-memory LRU cache  │
                         │  3. If miss, load from disk    │
                         │  4. Call TokenBucket.consume() │
                         │  5. Add to CRDT delta cache    │
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

### The important detail

**The request path never waits for gossip.**
If the node is alive, it answers.

---

## Core Components & Why They Exist

### 🪣 Token Bucket (Local Authority)

Each user is governed by a standard token bucket:

* capacity for bursts,
* refill rate for sustained traffic.

This is deliberately simple:

* predictable latency,
* constant-time decisions,
* easy to reason about under load.

Thread safety is explicit — no hidden concurrency tricks.

---

### 🧠 CRDT Synchronization (Global Convergence)

Local decisions generate **deltas**, not full state.

Why deltas?

* smaller payloads,
* less merge ambiguity,
* faster convergence.

CRDT merges are:

* commutative,
* idempotent,
* monotonic.

This guarantees that:

> even if messages are duplicated, delayed, or reordered, nodes eventually agree.

Exact precision is not promised.
**Bounded divergence is.**

---

### 📡 libp2p Gossip (Decentralization Without Orchestration)

There is:

* no leader,
* no coordinator,
* no central broker.

Nodes discover peers and exchange updates via libp2p gossip.
Failures are treated as routine, not exceptional.

If a node disappears:

* others continue,
* state reconverges when it returns.

---

### 💾 LRU + Disk (Scaling Beyond Memory)

Keeping billions of users in memory is unrealistic.

So the system:

* keeps **hot buckets in an in-memory LRU**,
* evicts cold buckets to **disk**,
* reloads lazily on demand.

This keeps:

* memory bounded,
* hot paths fast,
* cold users cheap.

Durability is pragmatic, not transactional.

---

## Performance Characteristics

Measured on a small libp2p mesh (3 nodes):

| Metric                 | Observation             |
| ---------------------- | ----------------------- |
| Throughput per node    | ~3,000 req/sec          |
| p99 request latency    | ~2 ms                   |
| p99 gossip convergence | ~2 ms                   |
| Gossip payload size    | ~3 KB                   |

These numbers matter less than *where latency lives*:

* request path → local only,
* synchronization → async.

---

## Failure Semantics (Explicit)

This system chooses availability over strict correctness.

* **Network partition** → nodes continue independently
* **Node crash** → local state lost, global state recovers
* **Delayed gossip** → temporary over-allowing possible

This is intentional.

If your use case requires **strict global enforcement**, this is not the right tool.

---

## Example Usage

```go
limiter := drl.NewRateLimiter(100, 10) // capacity, refill rate

if limiter.AllowRequest("user-123") {
    // request proceeds
} else {
    // rate limited locally
}
```

The API stays boring on purpose.
The complexity lives inside.

---

## When You Should (and Shouldn’t) Use This

**Good fit if:**

* low latency matters more than perfect precision,
* regions must operate independently,
* failures are common, not exceptional.

**Not a good fit if:**

* every request must respect a single global counter,
* over-allowing is unacceptable,
* centralized infrastructure is cheap and reliable for you.

---

## Closing Thought

This project is not about replacing Redis.

It’s about asking a harder question:

> *What does rate limiting look like when the system itself refuses to be centralized?*

If that question matters to you, this project might too.

