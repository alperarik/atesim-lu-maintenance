<a name="readme-top"></a>

<h3 align="center">ATESIM-Pluto</h3>

  <p align="center">
    Laser Unit Maintenance Tracker Application
  </p>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#build-and-run">Build and Run</a></li>
        <li><a href="#cli-flags">CLI Flags</a></li>
        <li><a href="#key-features">Key Features</a></li>
      </ul>
    </li>
    <li><a href="#test">Test</a></li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#after-maintenance">After Maintenance</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->

## About The Project

Pluto is a UDP-based laser unit monitoring service that:

- Tracks device activity, trigger counts, and maintenance status
- Maintains device state in memory
- Persists data to a SQLite database
- Features configurable thresholds for maintenance operations

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

- [![golang][golang]][golang-url]
- [![go-sqlite3][go-sqlite3]][go-sqlite3-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->

## Getting Started

#### Prerequisites

- Go 1.24.4 or later
- SQLite3

#### Build and Run

```bash
./pluto [flags]
```

```bash
go build -o pluto . && ./pluto -maintenance-threshold=5000 -udp-port=8080 -http-port=8081
```

### CLI Flags

- udp-port: UDP port to listen on (default: 8080)
- http-port: HTTP port for reload API (default: 8081)
- maintenance-threshold: Trigger count threshold for maintenance (default: 5000)

### Key Features

- Device Tracking:
    - Monitors device IP addresses and online status
    - Tracks current trigger counts (user must perform manuel reset, after maintenance)
    - Maintains total lifetime trigger counts
    - Records first registration and last seen timestamps
- Configuration:
    - Configurable UDP listen port
    - Separate HTTP port for reload operations
    - Adjustable maintenance threshold (default: 5000 triggers)
- Persistence:
    - SQLite database storage ("pluto.db")
    - Automatic device state loading on startup
- Interfaces:
    - UDP server for device communications
    - HTTP server for administrative reload operations

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Test

Run unit tests with:

```bash
go test -v ./test
```

```bash
go test ./...
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Usage

- On initial startup, the server will:
    - Create a new SQLite database file (pluto.db)
    - Initialize the device tracking system
    - Start listening on configured ports
    - Note: The warning about failing to load devices is expected on first run.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## After Maintenance

- To reload the running application's database mirror after manual manipulation (*reset trigger count after
  maintenance):

```bash
curl -X POST http://localhost:8081/reload
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->

[golang]: https://img.shields.io/badge/go-1.24.4-blue

[golang-url]: https://go.dev

[go-sqlite3]: https://img.shields.io/badge/go--sqlite3-1.14.28-orange

[go-sqlite3-url]: https://github.com/mattn/go-sqlite3