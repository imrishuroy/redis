# Redis: The Complete Engineering Guide

A deep dive into why Redis is blazingly fast, how it handles concurrency with a single thread, and everything you need to know to understand its architecture.

---

## Table of Contents

1. [What is Redis?](#what-is-redis)
2. [Data Types in Redis](#data-types-in-redis)
3. [Why Redis is Fast](#why-redis-is-fast)
4. [The Single-Threaded Architecture](#the-single-threaded-architecture)
5. [IO Multiplexing Explained](#io-multiplexing-explained)
6. [The Event Loop](#the-event-loop)
7. [Atomicity and Why It Matters](#atomicity-and-why-it-matters)
8. [Persistence Options](#persistence-options)
9. [Practical Examples](#practical-examples)
10. [Common Use Cases](#common-use-cases)

---

## What is Redis?

**Redis** stands for **Re**mote **Di**ctionary **S**erver. Think of it as a super-fast notebook that lives in your computer's RAM (memory) instead of on a hard drive.

### The Simple Analogy

Imagine you have two ways to look up a phone number:

1. **Hard Drive (Traditional Database)**: Like flipping through a physical phone book - you need to turn pages, scan entries, and it takes time.

2. **RAM (Redis)**: Like having all phone numbers memorized in your brain - instant recall, no searching needed.

Redis stores data in memory, making it **10-100x faster** than traditional disk-based databases.

### What Can Redis Do?

```
┌─────────────────────────────────────────────────────────────┐
│                        REDIS                                │
├─────────────────────────────────────────────────────────────┤
│  Database      │  Store and retrieve data instantly        │
│  Cache         │  Speed up slow database queries           │
│  Message Broker│  Send messages between applications       │
│  Session Store │  Keep user login sessions                 │
│  Rate Limiter  │  Control API request rates                │
│  Leaderboard   │  Real-time gaming scoreboards             │
└─────────────────────────────────────────────────────────────┘
```

---

## Data Types in Redis

Redis isn't just a simple key-value store. It supports **rich data structures** that make it incredibly versatile.

### 1. Strings

The simplest type. Store text, numbers, or binary data (up to 512MB).

```
SET name "Alice"           → OK
GET name                   → "Alice"

SET counter 100            → OK
INCR counter               → 101
DECR counter               → 100
```

**Use cases**: Caching, counters, session tokens

### 2. Lists

Ordered collections of strings. Think of them as arrays or queues.

```
LPUSH tasks "task1"        → Push to left (front)
RPUSH tasks "task2"        → Push to right (back)
LRANGE tasks 0 -1          → ["task1", "task2"]

┌───────────────────────────────┐
│  LIST: tasks                  │
│  ┌──────┬──────┬──────┐      │
│  │task1 │task2 │task3 │      │
│  └──────┴──────┴──────┘      │
│  ← LPUSH            RPUSH →  │
└───────────────────────────────┘
```

**Use cases**: Message queues, activity feeds, recent items

### 3. Sets

Unordered collections of **unique** strings. No duplicates allowed.

```
SADD fruits "apple"        → 1 (added)
SADD fruits "banana"       → 1 (added)
SADD fruits "apple"        → 0 (already exists!)
SMEMBERS fruits            → {"apple", "banana"}

# Set operations
SINTER setA setB           → Intersection (common elements)
SUNION setA setB           → Union (all elements)
SDIFF setA setB            → Difference (in A but not B)
```

**Use cases**: Tags, unique visitors, friend lists

### 4. Sorted Sets (ZSets)

Like Sets, but each element has a **score** for ordering.

```
ZADD leaderboard 100 "Alice"
ZADD leaderboard 85 "Bob"
ZADD leaderboard 120 "Charlie"

ZRANGE leaderboard 0 -1 WITHSCORES
→ [("Bob", 85), ("Alice", 100), ("Charlie", 120)]

┌─────────────────────────────────┐
│  SORTED SET: leaderboard        │
│  ┌─────────┬───────┐           │
│  │ Member  │ Score │           │
│  ├─────────┼───────┤           │
│  │ Bob     │  85   │           │
│  │ Alice   │ 100   │           │
│  │ Charlie │ 120   │           │
│  └─────────┴───────┘           │
└─────────────────────────────────┘
```

**Use cases**: Leaderboards, priority queues, time-series data

### 5. Hashes

Store objects with fields. Like a mini-database row.

```
HSET user:1 name "Alice"
HSET user:1 email "alice@example.com"
HSET user:1 age 30

HGETALL user:1
→ {name: "Alice", email: "alice@example.com", age: "30"}

┌─────────────────────────────────┐
│  HASH: user:1                   │
│  ┌─────────┬──────────────────┐│
│  │ name    │ "Alice"          ││
│  │ email   │ "alice@..."      ││
│  │ age     │ "30"             ││
│  └─────────┴──────────────────┘│
└─────────────────────────────────┘
```

**Use cases**: User profiles, product details, configuration

### 6. Streams

Append-only log data structure for event streaming.

```
XADD mystream * field1 value1
→ "1609459200000-0" (auto-generated ID)

# Like a message queue but with history
┌─────────────────────────────────────────┐
│  STREAM: mystream                       │
│  ┌────────────────┬────────────────┐   │
│  │ 1609459200000-0│ field1=value1  │   │
│  │ 1609459200001-0│ field2=value2  │   │
│  │ 1609459200002-0│ field3=value3  │   │
│  └────────────────┴────────────────┘   │
│              ↓ time flows down          │
└─────────────────────────────────────────┘
```

**Use cases**: Event sourcing, activity logs, real-time analytics

### 7. Other Types

- **Bitmaps**: Efficient storage for binary flags
- **HyperLogLog**: Estimate unique counts (uses minimal memory)
- **Geospatial**: Store and query location data

---

## Why Redis is Fast

Redis achieves its incredible speed through three fundamental design decisions:

### 1. Everything Lives in RAM

```
┌────────────────────────────────────────────────────────────┐
│                    ACCESS TIME COMPARISON                  │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  RAM Access:        ~100 nanoseconds     ████              │
│                                                            │
│  SSD Access:        ~100 microseconds    ████████████████  │
│                     (1,000x slower)      ████████████████  │
│                                          ████████████████  │
│                                                            │
│  HDD Access:        ~10 milliseconds     ████████████████  │
│                     (100,000x slower)    ████████████████  │
│                                          ████████████████  │
│                                          ... (way more)    │
└────────────────────────────────────────────────────────────┘
```

When you run `GET user:1`, Redis doesn't need to:
- Seek to a position on disk
- Wait for disk platters to spin
- Transfer data through slow I/O channels

It just reads directly from memory - the same memory your CPU uses for everything.

### 2. Optimized Data Structures

Redis doesn't use generic data structures. Each type is carefully optimized:

```
┌─────────────────────────────────────────────────────────────┐
│  Data Type    │  Internal Implementation                   │
├───────────────┼─────────────────────────────────────────────┤
│  Strings      │  Simple Dynamic Strings (SDS)              │
│  Lists        │  Quicklist (linked list of ziplists)       │
│  Sets         │  Hashtable or Intset (for small sets)      │
│  Sorted Sets  │  Skip List + Hashtable                     │
│  Hashes       │  Hashtable or Ziplist (for small hashes)   │
└─────────────────────────────────────────────────────────────┘
```

For example, a small hash with few fields uses a **ziplist** (compact, sequential memory) instead of a full hashtable, saving memory and improving cache locality.

### 3. Single-Threaded = No Locks

This is the counterintuitive genius of Redis. Let's explore this deeply.

---

## The Single-Threaded Architecture

### The Problem with Multi-Threading

Imagine a bank account with $100. Two people try to withdraw $80 at the same time:

```
WITHOUT PROPER LOCKING (Bug!):

Thread 1                    Thread 2
────────                    ────────
Read balance: $100          
                            Read balance: $100
Check: 100 >= 80? YES       
                            Check: 100 >= 80? YES
Withdraw $80                
                            Withdraw $80
New balance: $20            
                            New balance: $20

Result: Bank loses $60! Both withdrawals succeeded.
```

To fix this, databases use **locks** (mutexes):

```
WITH LOCKING:

Thread 1                    Thread 2
────────                    ────────
Acquire lock ✓              
Read balance: $100          Waiting for lock... ⏳
Check: 100 >= 80? YES       Waiting for lock... ⏳
Withdraw $80                Waiting for lock... ⏳
New balance: $20            Waiting for lock... ⏳
Release lock                
                            Acquire lock ✓
                            Read balance: $20
                            Check: 20 >= 80? NO
                            Reject withdrawal ✗
                            Release lock
```

**The problem**: Locks are expensive!

```
┌────────────────────────────────────────────────────────────┐
│                 MULTI-THREADING OVERHEAD                   │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  1. Lock Acquisition     - CPU cycles to obtain lock       │
│  2. Lock Contention      - Threads waiting, doing nothing  │
│  3. Context Switching    - OS switching between threads    │
│  4. Cache Invalidation   - CPU caches become stale         │
│  5. Memory Barriers      - Ensuring memory consistency     │
│                                                            │
│  All this adds up to significant overhead!                 │
└────────────────────────────────────────────────────────────┘
```

### Redis Solution: Just Don't Use Threads

Redis says: "What if we just process one command at a time?"

```
SINGLE-THREADED REDIS:

Command Queue              Redis Server
─────────────              ────────────
┌──────────┐               
│ GET x    │ ──────────→  Process GET x
├──────────┤               Return result
│ SET y    │ ──────────→  Process SET y
├──────────┤               Return result
│ INCR z   │ ──────────→  Process INCR z
└──────────┘               Return result

No locks needed! Each command has exclusive access.
```

**But wait** - doesn't this mean only one client can connect?

**No!** This is where the magic happens: **IO Multiplexing**.

---

## IO Multiplexing Explained

### The Naive Approach (Thread Per Connection)

```
Traditional Server:

Client 1  ────→  Thread 1  ────→  Database
Client 2  ────→  Thread 2  ────→  Database
Client 3  ────→  Thread 3  ────→  Database
   ...           ...
Client N  ────→  Thread N  ────→  Database

Problem: 10,000 clients = 10,000 threads = CHAOS
- Each thread uses ~1MB stack memory
- Context switching nightmare
- Thread creation/destruction overhead
```

### The Smart Approach (IO Multiplexing)

```
Redis with IO Multiplexing:

                    ┌─────────────────┐
Client 1  ────┐     │                 │
Client 2  ────┼────→│  Event Loop     │────→  Single Thread
Client 3  ────┤     │  (epoll/kqueue) │       Processing
   ...        │     │                 │
Client N  ────┘     └─────────────────┘

One thread handles ALL clients!
```

### How Does It Work?

Think of a **waiter at a restaurant**:

**Bad Waiter (Blocking I/O)**:
1. Go to Table 1, stand there waiting for them to decide
2. They're still reading the menu... keep waiting...
3. 10 minutes later, take their order
4. Now go to Table 2, wait again...

**Good Waiter (IO Multiplexing)**:
1. Scan all tables
2. Table 3 has raised their hand → take their order
3. Scan all tables again
4. Table 1 is ready → take their order
5. Table 5 has raised their hand → take their order
6. Never waste time waiting!

```
┌────────────────────────────────────────────────────────────┐
│                    IO MULTIPLEXING                         │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  System Call: epoll_wait() / kqueue() / select()          │
│                                                            │
│  "Hey OS, watch these 10,000 sockets for me.              │
│   Wake me up ONLY when one of them has data ready."       │
│                                                            │
│  ┌─────────┐                                              │
│  │ Socket 1│ ─── waiting ───                              │
│  │ Socket 2│ ─── waiting ───                              │
│  │ Socket 3│ ─── DATA READY! ←── Process this one!        │
│  │ Socket 4│ ─── waiting ───                              │
│  │   ...   │                                              │
│  └─────────┘                                              │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### The Key Insight

**Network I/O is SLOW compared to CPU operations.**

```
Time to send 1KB over network:  ~10 microseconds
Time for Redis to process SET:  ~1 microsecond

While waiting for network data, CPU could have done
10 operations! Why waste that time blocking?
```

IO Multiplexing ensures the CPU is **always doing useful work**, never waiting.

---

## The Event Loop

The heart of Redis is its **Event Loop**. Here's how it works:

```
┌─────────────────────────────────────────────────────────────┐
│                    REDIS EVENT LOOP                         │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │     START     │
                    └───────┬───────┘
                            │
                            ▼
              ┌─────────────────────────────┐
              │  Wait for events            │
              │  (epoll_wait / kqueue)      │◄─────────────┐
              └─────────────┬───────────────┘              │
                            │                              │
                            ▼                              │
              ┌─────────────────────────────┐              │
              │  New connection?            │              │
              │  → Accept it, add to watch  │              │
              └─────────────┬───────────────┘              │
                            │                              │
                            ▼                              │
              ┌─────────────────────────────┐              │
              │  Data ready to read?        │              │
              │  → Read command from client │              │
              └─────────────┬───────────────┘              │
                            │                              │
                            ▼                              │
              ┌─────────────────────────────┐              │
              │  Parse & Execute command    │              │
              │  (This is the fast part!)   │              │
              └─────────────┬───────────────┘              │
                            │                              │
                            ▼                              │
              ┌─────────────────────────────┐              │
              │  Send response to client    │              │
              └─────────────┬───────────────┘              │
                            │                              │
                            └──────────────────────────────┘
                                    Loop forever
```

### Code Representation (Simplified)

```go
// Simplified Redis event loop concept
func eventLoop() {
    for {
        // Wait for something to happen on any socket
        readyEvents := epoll.Wait()
        
        for _, event := range readyEvents {
            switch event.Type {
            case NEW_CONNECTION:
                // Accept new client
                client := acceptConnection(event.Socket)
                epoll.Add(client.Socket)
                
            case DATA_READY:
                // Read and parse command
                command := readCommand(event.Socket)
                
                // Execute command (THE FAST PART)
                result := executeCommand(command)
                
                // Send response
                sendResponse(event.Socket, result)
            }
        }
    }
}
```

### Why This Works So Well

```
┌─────────────────────────────────────────────────────────────┐
│                   TIME BREAKDOWN                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Waiting for network data:     ~1-10 ms   (SLOW)           │
│  Reading data from socket:     ~10 µs     (fast)           │
│  Executing Redis command:      ~1 µs      (BLAZING FAST)   │
│  Sending response:             ~10 µs     (fast)           │
│                                                             │
│  The Event Loop ensures we only spend time on the          │
│  fast parts, never blocking on the slow network wait!      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Atomicity and Why It Matters

Every Redis command is **atomic**. This means:

### What Atomic Means

```
ATOMIC = Indivisible

When you run: INCR counter

This happens as ONE unit:
1. Read current value
2. Add 1
3. Write new value
4. Return result

NOTHING can happen in between these steps.
```

### Why This Matters

```
Two clients running INCR counter simultaneously:

┌─────────────────────────────────────────────────────────────┐
│  NON-ATOMIC (Broken):                                      │
│                                                             │
│  Client A: Read 5 ──┐                                      │
│  Client B: Read 5 ──┼── Both read 5                        │
│  Client A: Write 6 ─┤                                      │
│  Client B: Write 6 ─┘── Both write 6. Lost an increment!   │
│                                                             │
│  Expected: 7, Got: 6  ✗                                    │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  ATOMIC (Redis):                                           │
│                                                             │
│  Client A: INCR ──→ [Read 5, Write 6] ──→ Returns 6       │
│  Client B: INCR ──→ [Read 6, Write 7] ──→ Returns 7       │
│                                                             │
│  Commands execute one at a time. Always correct! ✓         │
└─────────────────────────────────────────────────────────────┘
```

### Atomic Operations in Redis

```
INCR key          # Atomic increment
DECR key          # Atomic decrement
LPUSH + LTRIM     # Use MULTI/EXEC for compound atomicity
GETSET key value  # Atomic get-and-set
SETNX key value   # Set if not exists (atomic)
```

### Transactions with MULTI/EXEC

For multiple commands that need to be atomic together:

```
MULTI                    # Start transaction
SET user:1:name "Alice"  # Queued
SET user:1:email "a@b"   # Queued
INCR user:count          # Queued
EXEC                     # Execute all at once

# All three commands execute atomically
# Either all succeed or none do
```

---

## Persistence Options

"But if Redis is in-memory, don't I lose everything on restart?"

No! Redis offers two persistence mechanisms:

### 1. RDB (Redis Database Backup)

**Periodic snapshots** of your entire dataset.

```
┌─────────────────────────────────────────────────────────────┐
│                    RDB SNAPSHOTS                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Time ──────────────────────────────────────────────────→  │
│                                                             │
│       [Snapshot]         [Snapshot]         [Snapshot]     │
│          │                  │                  │           │
│          ▼                  ▼                  ▼           │
│       dump.rdb           dump.rdb           dump.rdb       │
│                                                             │
│  Configuration:                                            │
│  save 900 1      # Snapshot if 1 key changed in 15 min    │
│  save 300 10     # Snapshot if 10 keys changed in 5 min   │
│  save 60 10000   # Snapshot if 10000 keys changed in 1 min│
│                                                             │
└─────────────────────────────────────────────────────────────┘

Pros: Compact, fast restart, good for backups
Cons: Can lose data between snapshots
```

### 2. AOF (Append-Only File)

**Log every write operation** to a file.

```
┌─────────────────────────────────────────────────────────────┐
│                    AOF LOGGING                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Every write command gets appended to appendonly.aof:      │
│                                                             │
│  SET user:1 "Alice"                                        │
│  INCR pageviews                                            │
│  LPUSH queue "job1"                                        │
│  SET user:2 "Bob"                                          │
│  INCR pageviews                                            │
│  ...                                                       │
│                                                             │
│  On restart: Replay all commands to rebuild state          │
│                                                             │
│  fsync options:                                            │
│  - always: Every command (safest, slowest)                 │
│  - everysec: Every second (good balance) ← Recommended     │
│  - no: Let OS decide (fastest, riskiest)                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘

Pros: Minimal data loss (at most 1 second)
Cons: Larger files, slower restart
```

### Best Practice: Use Both

```
┌─────────────────────────────────────────────────────────────┐
│                    HYBRID APPROACH                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  RDB for:                                                  │
│  - Fast restarts                                           │
│  - Point-in-time backups                                   │
│  - Disaster recovery                                       │
│                                                             │
│  AOF for:                                                  │
│  - Minimal data loss                                       │
│  - Recovery from corruption                                │
│                                                             │
│  On restart: Load RDB first, then replay AOF               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Practical Examples

### Example 1: Rate Limiting

Limit API requests to 100 per minute per user:

```
function isRateLimited(userId) {
    key = "ratelimit:" + userId
    
    current = INCR(key)           // Atomic increment
    
    if current == 1 {
        EXPIRE(key, 60)           // First request, set 60s TTL
    }
    
    return current > 100          // true = blocked
}

┌─────────────────────────────────────────────────────────────┐
│  Request 1:   INCR → 1, set TTL 60s    → Allowed          │
│  Request 2:   INCR → 2                  → Allowed          │
│  ...                                                        │
│  Request 100: INCR → 100                → Allowed          │
│  Request 101: INCR → 101                → BLOCKED!         │
│  ...                                                        │
│  [60 seconds pass, key expires]                            │
│  Request N:   INCR → 1, set TTL 60s    → Allowed          │
└─────────────────────────────────────────────────────────────┘
```

### Example 2: Session Storage

```
// Store session
HSET session:abc123 userId "42"
HSET session:abc123 email "user@example.com"
HSET session:abc123 role "admin"
EXPIRE session:abc123 3600    // 1 hour TTL

// Retrieve session
HGETALL session:abc123
→ {userId: "42", email: "user@example.com", role: "admin"}

// Check if session exists
EXISTS session:abc123
→ 1 (yes) or 0 (no/expired)
```

### Example 3: Real-time Leaderboard

```
// Add/update scores
ZADD leaderboard 1500 "player:alice"
ZADD leaderboard 2300 "player:bob"
ZADD leaderboard 1800 "player:charlie"

// Get top 10 players
ZREVRANGE leaderboard 0 9 WITHSCORES
→ [("player:bob", 2300), ("player:charlie", 1800), ...]

// Get player rank (0-indexed)
ZREVRANK leaderboard "player:alice"
→ 2 (third place)

// Increment score
ZINCRBY leaderboard 500 "player:alice"
→ 2000 (new score)
```

### Example 4: Pub/Sub Messaging

```
// Publisher (sends messages)
PUBLISH notifications "New order received!"

// Subscriber (receives messages)
SUBSCRIBE notifications
→ Waiting for messages...
→ Received: "New order received!"

┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  Publisher ─── PUBLISH "channel" "msg" ───→ Redis         │
│                                              │             │
│                                              ▼             │
│  Subscriber 1 ←─────────────────────── "channel"          │
│  Subscriber 2 ←─────────────────────── "channel"          │
│  Subscriber 3 ←─────────────────────── "channel"          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Common Use Cases

### 1. Caching

```
function getUser(userId) {
    // Try cache first
    cached = GET("user:" + userId)
    if cached != nil {
        return JSON.parse(cached)
    }
    
    // Cache miss - fetch from database
    user = database.query("SELECT * FROM users WHERE id = ?", userId)
    
    // Store in cache for 1 hour
    SET("user:" + userId, JSON.stringify(user), EX=3600)
    
    return user
}
```

### 2. Distributed Locking

```
// Acquire lock
SET lock:resource123 "owner:me" NX EX 30
// NX = only if not exists
// EX 30 = expires in 30 seconds (prevents deadlock)

// Do critical work...

// Release lock (only if we own it)
if GET("lock:resource123") == "owner:me" {
    DEL("lock:resource123")
}
```

### 3. Queue Processing

```
// Producer adds jobs
LPUSH job_queue '{"type": "email", "to": "user@example.com"}'

// Consumer processes jobs (blocking pop)
while true {
    job = BRPOP job_queue 0    // Block until job available
    process(job)
}
```

---

## Summary: Why Redis Wins

```
┌─────────────────────────────────────────────────────────────┐
│                    THE REDIS FORMULA                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐  │
│   │  In-Memory   │ + │ Single Thread│ + │ IO Multiplex │  │
│   │  Storage     │   │ (No Locks)   │   │ (epoll/kqueue│  │
│   └──────────────┘   └──────────────┘   └──────────────┘  │
│          │                  │                  │           │
│          ▼                  ▼                  ▼           │
│   Microsecond        Zero lock           CPU never         │
│   data access        overhead            waits idle        │
│                                                             │
│                         ═══                                │
│                          │                                 │
│                          ▼                                 │
│              ┌────────────────────┐                        │
│              │ 100,000+ ops/sec   │                        │
│              │ per single thread! │                        │
│              └────────────────────┘                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Key Takeaways

1. **In-Memory = Fast**: No disk I/O for data operations
2. **Single-Threaded = Simple**: No locks, no race conditions, no complexity
3. **IO Multiplexing = Scalable**: One thread handles thousands of connections
4. **Atomic Commands = Safe**: No data corruption, no lost updates
5. **Rich Data Types = Versatile**: Not just key-value, but lists, sets, sorted sets, and more
6. **Persistence Options = Durable**: RDB and AOF ensure data survives restarts

Redis proves that sometimes the simplest architecture is the most powerful. By removing complexity (threads, locks, disk I/O), Redis achieves performance that more complex systems struggle to match.

---

## Further Reading

- [Redis Official Documentation](https://redis.io/docs/)
- [Redis University](https://university.redis.com/)
- [Redis Best Practices](https://redis.io/docs/management/optimization/)
