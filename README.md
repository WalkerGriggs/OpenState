
<p align=center>
  <img alt="OpenState" src="./assets/black_alpha.png" width="50%">
</p>

OpenState is a language agnostic task runner focusesd on a low code, declarative approach to
task definition and lifecycle management.


## Features

**Note: OpenState is pre-alpha and should not be used in production environments
(yet). Its API and core schema is subject to change.**

### Present

The current feature set includes:

- Raft and Serf instrumentation for strong consistency
- Task definitions and an event-driven Finite State Machine framework

### Future

A number of these items are aspirational, but might form some semblance of a
roadmap. Pull requests are always welcome, of course!

- Typed task callbacks and pluggable runtime drivers.
  - Docker support for various environments (locally & Kubernetes)
  - Pre-defined callbacks for HTTP, SMTP, Kafka etc.
- Integration with Vault for secure key management.

## Getting Started

### Building

Assuming you have a working Go environment, setup is easy and all required
dependencies are vendored.

To build the project:

```bash
make dev
```

To build for a specific environment:

```bash
make pkg/linux_amd64/openstate
make pkg/linux_386/openstate
make pkg/darwin_amd64/openstate

# Windows and ARM architectures are not supported.
```

Of course, if you'd prefer to run the project directly, `main.go` in the top level
directory is an easy entrypoint.

### Running

OpenState is configured with sensible defaults, but you'll need to make some
adjustments if you plan to run a local cluster (ie. address ports).

Config files are written in YAML, and all keys can be set through CLI flags as
well. Below is a sample config that might serve as a good jumping off point.

```yaml
# log_level is the verbosity of OpenState's logger. Both the Raft implementation
# and the Serf protocol inherit this log level.
log_level: INFO

# dev_mode indicates if OpenState is running in a development environment. It will
# disable all persistence, and opt for in-memory stores instead.
dev_mode: false

# data_dir is the path to the direct where OpenState will store persisted objects
# ie. snapshots, logs, and stable stores.
data_dir: $HOME/.openstate/

# advertise are the addresses for Raft, Serf, and HTTP endpoints. They must
# be unique (in both this config, and across running servers on the same host)
advertise:
    http: 127.0.0.1:8080
    raft: 127.0.0.1:4648
    serf: 127.0.0.1:8080

server:
    # node_name is the advertised name of the server. It's most commonly used in
    # Serf's gossip protocol.
    node_name: node-1

    # bootstrap_expect is the number of peers the new server should expect. If 1,
    # the new server will create a single node cluster and elect itself leader.
    bootstrap_expect: 1

    # Join is the initial, comma separated list of peer serf addresses. This option
    # is a hack to bypass the need for service discovery (TODO). This list only
    # needs to contain contain ONE valid peer; the gossip layer will propogate the
    # new peer across all nodes.
    # join:
```
