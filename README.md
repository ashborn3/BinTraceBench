# BinTraceBench

**BinTraceBench** is a systems-level backend tool written in Go that performs **static and dynamic analysis of binaries**, **live process inspection**, and **sandboxed benchmarking** - all accessible through a clean **RESTful API**.

> Not for developers, security researchers, reverse engineers, and performance analysts.

---

## Features

### 🔍 Binary Analyzer
- **Static Analysis**: ELF/PE parsing - symbols, headers, sections, strings
- **Dynamic Tracing**: `ptrace`-based syscall tracing with full argument dumps

### Sandboxed Benchmarking
- Run binaries inside Linux namespaces (`unshare`)
- Collect execution time, exit code, success status
- Optional dynamic tracing during benchmark

### Live Process Inspector
- Inspect `/proc/:pid` to view:
  - Command-line args
  - Open file descriptors
  - Network connections

---

## REST API Endpoints

### Binary Analysis
| Method | Endpoint                        | Description                      |
|--------|----------------------------------|----------------------------------|
| `POST` | `/analyze`                       | Static binary analysis           |
| `POST` | `/analyze?dynamic=true`          | Dynamic tracing using `ptrace`   |
| `GET`  | `/analyze/:id/logs` *(optional)* | Retrieve previous trace logs     |

### Benchmarking
| Method | Endpoint        | Description                                |
|--------|------------------|--------------------------------------------|
| `POST` | `/bench`         | Run binary in sandbox and benchmark        |
| `GET`  | `/bench/:id`     | *(optional)* Fetch previous benchmark data |

### Process Inspection
| Method | Endpoint               | Description                    |
|--------|-------------------------|--------------------------------|
| `GET`  | `/proc/:pid`            | Basic process info             |
| `GET`  | `/proc/:pid/files`      | Open file descriptors          |
| `GET`  | `/proc/:pid/net`        | Network connections            |

---

## Getting Started

### Requirements
- Go 1.20+
- Linux OS (for `ptrace`, `/proc`, `unshare`)

### Installation

```bash
git clone https://github.com/yourname/BinTraceBench.git
cd BinTraceBench
go build -o bintracer ./internal/tools/bintracer.go
go build -o bintracebench ./cmd/bintracebench
````

### Running the Server

```bash
./bintracebench
# Server runs on http://localhost:8080
```

---

## Testing with Python

Use the provided [test script](./scripts/endpointTester.py):

```bash
python3 ./scripts/endpointTester.py
```

---

## Sample Output

```json
{
  "exit_code": 0,
  "runtime_ms": 42,
  "success": true,
  "syscalls": [
    {
      "name": "execve",
      "args": [
        "arg0=0x55f8aa23b000",
        "arg1=0x7ffd6c28b6a8",
        "arg2=0x7ffd6c28b6b0",
        "arg3=0x0",
        "arg4=0x0",
        "arg5=0x0"
      ]
    }
  ]
}
```

---

## Roadmap

* [x] Static binary parsing
* [x] Dynamic syscall tracing
* [x] Sandbox + benchmarking
* [x] Live process inspection
* [ ] SQLite-based persistent storage
* [ ] More detailed binary analysis
* [ ] LLM-based insights on binary analysis
* [ ] Return value and errno analysis
* [ ] Export results to JSON/HTML
* [ ] Syscall statistics & timeline view
* [ ] Web dashboard UI (optional) [Ahh Frontend man, my old nemesis...]

---

## Credits

Built in pain by [ashborn3](https://github.com/ashborn3)
Inspired by strace, perf, and devtools used in systems programming.

```