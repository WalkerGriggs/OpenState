# ADR 001: In-Memory Storage

## Changelog

  - [2021-04-06] [WalkerGriggs](https://github.com/walkergriggs): Drafting document

## Context

Currently, Raft's FSM is storing all objects in memory with native Go datastructures. This is not a safe storage method, and is subject to all sorts of race conditions. Instead, we should use existing in-memory datastores to gain multi-version concurrency control MVCC and atomic, ACID transactions.

## Pior Art

N/A

## Requirements

### Musts

- **Transactions**: We need atomicity across reads and writes. Ideally, we'd be able to differential between read-only and writable txns.
- **Active support**: We need a datastore with a strong OSS community or investment from a stable company.

### Ideals

- **Complex objects**: We'd prefer the datastore to handle complex objects. KV stores are fine, but namespacing kv-pairs is not a pattern we're used to
- **Snapshotting**: Raft FSMs need to be snapshotted and periodically written to disc. We like to be able to snapshot / serialize the datastore.

## Considered Options

### bolt (bbolt)

[bbolt](https://github.com/etcd-io/bbolt), originally [bolt](https://github.com/boltdb/bolt), is simple key/value store maintained by CoreOS under the etcd project.

**Pros:**

  - Stable: semantically versioned api, fixed file format
  - Trusted: currently used in production at a number of companies
  - Supported: CoreOS (and the surrounding OSS community) have invested quite a bit in Bolt and show no signs of stopping

**Cons:**

  - K/V: I'd prefer not to use a key value datastore, and would prefer instead to store complex objects whole. KV is simply a storage pattern I'm not accustomed to.
  - Not in-memory: Though Bolt runs as part of the program (a forked process, presumably), it still writes to disk. Performance concerns aside, writing a db to disk feels redundant when the entire Raft FSM is serialized and written to disk.

### go-memdb

[go-memdb](https://github.com/hashicorp/go-memdb) is an in-memory database based on radix trees, written and maintained by Hashicorp.

**Pros:**

  - Supported: Hashi uses go-memdb in a number of its products and receives consistent support.
  - Easily Integrated: OpenState is already using Hashi's Raft and Serf implementations. They've established preceidence for using go-memdb in conjunction with these frameworks. Nomad is a great example.
  - Buzzwords: Atomic, ACID, with transactions and indexing

### etcd

[Etcd](https://github.com/etcd-io/etcd) is a distributed key value store developed and maintained by CoreOS

**Pros:**

  - Trusted: Etcd's flagship application is Kubernetes. Enough said.
  - Supported: So long a K8s is around, etcd will receive investment.

**Cons:**

  - K/V: I'd prefer not to use a key value datastore, and would prefer instead to store complex objects whole. KV is simply a storage pattern I'm not accustomed to.
  - Overhead: Not only is Etcd not in-memory, but it requires a separate cluster which adds operational complexity and increases cost-to-serve

## Decision

Given OpenState's usecase, go-memdb makes the most sense (for the time being). It meets all initial search requirements, can be implemented immediately, and doesn't require any lasting architectural changes (should it need to be replaced with a heavier option like Etcd).

### Consequences

- go-memdb does not receive the same support as Bolt and Etcd, nor does it have as extensive a community. Documentation isn't exactly IC-friendly, and finding answers to FAQ / common issues will be a challenge.
- go-memdb potentially adds more overhead. It's entirely in-memory and runs as a child process, so that'll consume more CPU/mem than, say, etcd.
- Simplicity means we can integrate the datastore and move on to other issues.


### Design

TODO

## Current Status

Accepted
