# Decentralized Rate Limiter

A production-grade **decentralized rate limiter** built in **Go**, using **CRDTs** and **libp2p** for conflict-free state sync across distributed nodes without a central authority. Designed to maintain **eventual consistency**, **fault tolerance**, and **high performance** under heavy load.

---

## Overview

Traditional rate limiters depend on centralized coordination (e.g., Redis), creating a single point of failure. This project implements a **distributed** approach where each node:

- Tracks request counts locally using **token buckets**.
- Shares updates using **delta-state CRDTs**.
- Communicates over **libp2p** to maintain a fully decentralized sparse mesh.

---

## Architecture

### Key Components:

| Component     | Role |
|---------------|------|
| `CRDT`        | Merges token buckets using delta-state conflict-free logic. |
| `TokenBucket` | Rate limits for individual users. |
| `libp2p`      | Peer-to-peer communication layer for state sync. |
| `Gin`         | HTTP server handling user requests and applying rate limiting middleware. |

### Design Diagram:

```
       ┌─────────────┐
       │   Client    │
       └─────┬───────┘
             │
       ┌─────▼──────┐
       │  Gin API   │
       └─────┬──────┘
             │
   ┌─────────▼─────────┐
   │ RateLimiter (CRDT)│
   └─────────┬─────────┘
             │
       ┌─────▼─────┐
       │  libp2p   │<─── Peers
       └───────────┘
```

---

## Getting Started

### Prerequisites

- Go 1.21+
- Docker (optional, for benchmarking)

### Run a Node

```bash
git clone https://github.com/souviks22/decentralized-rate-limiter.git
cd decentralized-rate-limiter
go run cmd/node/main.go
````
It will create a bootstrap node followed by logging its multiaddress. This is the very first server in our decentralized network. Afterwards, run every new server with an environment variable `BOOTSTRAP_PEER` equal to that predefined multiaddress. Each node will automatically attempt to join the libp2p mesh.

---

## Features

* Fully decentralized token bucket rate limiting
* CRDT-based eventual consistency
* Resilient to partial node failures
* Delta-based broadcasting to reduce network overhead
* Plug-and-play middleware for Gin-based APIs

---

## Benchmarking & Metrics

Coming Soon

---

## 📂 Project Structure

```
.
├── cmd/                # Main entrypoint
├── internal/
│   ├── limiter/        # CRDT logic, token buckets
│   ├── middleware/     # Gin rate limiting middleware
│   └── p2p/            # libp2p networking code
├── go.mod
└── README.md
```

---

## Design Decisions

* **CRDTs** were used for safe, conflict-free merges across distributed nodes.
* **Delta-based sync** to improve efficiency over full state broadcast.
* **Thread-safe** concurrent access ensured using `sync.Mutex` over shared resources.
* Modular architecture to allow horizontal scaling and easier testing.

---

## Limitations & TODO

* Memory saving strategy for too many distinct users.
* Persist buckets across restarts (disk-based storage).
* Graceful shutdown and state handoff.
