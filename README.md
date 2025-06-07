# 🌐 Decentralized Rate Limiter

[![Go Report Card](https://goreportcard.com/badge/github.com/souviks22/decentralized-rate-limiter)](https://goreportcard.com/report/github.com/souviks22/decentralized-rate-limiter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Made with Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://golang.org)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](#contributing)

> A fault-tolerant, decentralized, and CRDT-synced rate limiter for large-scale distributed systems — designed to scale to billions of users with local failover and eventual consistency.

---

## 🚀 Features

- ⏳ **Token Bucket Algorithm** for burst-friendly traffic control.
- 🧠 **CRDT-powered synchronization** — conflict-free, peer-to-peer.
- 💾 **LRU cache with disk persistence** — supports both active and inactive users efficiently.
- 📡 **libp2p gossip** — decentralized and self-healing.
- 🔒 **Resilience to partitions and node failure**.

---

## 📸 Architecture

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

- **Each node** locally limits requests and syncs token deltas with others via gossip.
- **State merging** is done using CRDT-style max-based reconciliation.


---

## 🔧 Installation

```bash
git clone https://github.com/souviks22/decentralized-rate-limiter.git
cd decentralized-rate-limiter
go mod tidy
go run main.go
````

> Requires Go 1.20+ and a writable `/data/<node_id>` directory for disk persistence. Optionally, `Docker` configurations for bootstrap node and network nodes are also available.

---

## 🧪 Example Usage

```go
limiter := limiter.New(100.0, 10.0) // capacity = 100, refillRate = 10 tokens/sec

if limiter.AllowRequest("user-123") {
    // ✅ Proceed with request
} else {
    // ❌ Rate limited
}
```

---

## 🧠 Internals

### 🪣 TokenBucket

* Uses capacity, refill rate, and timestamps to refill tokens.
* Thread-safe with mutex locks.
* Supports delta-based `merge()` for CRDT sync.

### 🧠 CRDT

* Batched updates pushed via libp2p `Broadcast()`.
* Incoming deltas merged every `100ms`.
* Cold buckets are periodically flushed to disk.

### 🧱 Disk + LRU

* Evicted buckets go to disk for durability.
* Reloaded lazily when requested again.
* Guarantees hot-path speed and cold-path persistence.

---

## 📊 Benchmark & Testing

Testing in Progress

---

## 📎 Links

* [libp2p Docs](https://libp2p.io)
* [CRDTs Explained](https://crdt.tech/)
* [Go LRU Cache](https://github.com/hashicorp/golang-lru)]
