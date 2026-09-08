# Dokumentasi Proyek: Rig-to-Cloud Telemetry Pipeline

**Nama repositori:** `rig-to-cloud-telemetry`  
**Jenis:** Polyglot monorepo (Go Workspaces + Python Simulator + React Dashboard)  
**Status dokumen:** As-Is (setelah implementasi end-to-end)  
**Tanggal snapshot:** 8 September 2026  
**Fase:** Pipeline end-to-end fungsional (local/dev)

---

## 1. Ringkasan Eksekutif

Proyek ini merancang **pipeline telemetri pengeboran (drilling telemetry)** yang tangguh dari *rig* (lokasi sumur) ke *cloud*. Tujuannya:

1. Membaca data sensor/WITS dari sisi edge (rig).
2. Menyangga data lokal saat koneksi tidak stabil.
3. Mensinkronkan ke cloud melalui MQTT.
4. Menyimpan ke TimescaleDB sebagai time-series.
5. Menampilkan well-log / strip chart di dashboard React.

Saat ini pipeline **sudah terhubung end-to-end untuk lokal/dev**:

`simulator (PTY) → edge-daemon (WITS → SQLite outbox → MQTT) → cloud-backend (MQTT → TimescaleDB → REST/WS) → dashboard`

Auth produksi, CI/CD, dan hardening keamanan masih di luar scope fase ini.

---

## 2. Latar Belakang & Masalah yang Diselesaikan

### 2.1 Konteks industri

Di operasi pengeboran minyak & gas, data sensor (kedalaman bit, ROP, WOB, gamma ray, dll.) sering dikirim dalam format **WITS** (Wellsite Information Transfer Specification) melalui **serial port** atau jaringan lokal. Tantangan khas:

| Tantangan | Dampak |
|-----------|--------|
| Koneksi uplink intermittent | Data hilang jika tidak ada buffer lokal |
| Edge device constrained | Perlu agen ringan, storage embedded |
| Format legacy (WITS) | Perlu parser khusus sebelum masuk cloud |
| Kebutuhan visualisasi real-time | Butuh API/WebSocket + chart khusus well-log |

### 2.2 Solusi yang dituju (target arsitektur)

```
[WITS Generator / Serial]
        │
        ▼
┌───────────────────┐     MQTT      ┌───────────────────┐
│   edge-daemon     │──────────────▶│  cloud-backend    │
│ serial→WITS→SQLite│  (resync)     │ broker→store→API  │
└───────────────────┘               └─────────┬─────────┘
                                              │
                                    TimescaleDB + WebSocket/HTTP
                                              │
                                              ▼
                                        dashboard
```

Kata kunci desain: **resilient** — edge menyimpan ke SQLite dulu, baru sync ke cloud saat jaringan tersedia.

---

## 3. Strategi Monorepo

Proyek memakai **satu repositori Git** dengan **Go Workspaces** (`go.work`), bukan banyak repo terpisah.

### 3.1 Alasan

| Alasan | Penjelasan |
|--------|------------|
| Pemisahan edge vs cloud | Modul Go terpisah (`rigtelemetry/edge` vs `rigtelemetry/cloud`) agar boundary jelas |
| Kontrak data bersama | Modul `rigtelemetry/shared` tanpa publish paket privat |
| Co-location tooling | Simulator Python + dashboard React + Docker Compose satu tempat untuk lokal testing |
| DX monorepo | Satu `make init`, satu `docker-compose.yml` |

### 3.2 Go Workspace

File `go.work`:

```text
go 1.22

use (
	./cloud-backend
	./edge-daemon
	./shared
)
```

Setiap service Go punya `go.mod` sendiri dan `replace` ke `../shared`:

| Modul | Path | Nama modul |
|-------|------|------------|
| Shared models | `shared/` | `rigtelemetry/shared` |
| Edge agent | `edge-daemon/` | `rigtelemetry/edge` |
| Cloud ingest/API | `cloud-backend/` | `rigtelemetry/cloud` |

---

## 4. Struktur Direktori (As-Is)

```text
rig-to-cloud-telemetry/
├── .gitignore
├── README.md                 # Quick start (EN, ringkas)
├── Makefile                  # Target operasional lokal
├── docker-compose.yml        # TimescaleDB + Mosquitto
├── go.work                   # Go workspace root
├── go.work.sum
│
├── shared/                   # Kontrak data bersama
│   ├── go.mod
│   └── model/
│       └── telemetry.go      # struct TelemetryPoint
│
├── edge-daemon/              # Agen di rig (STUB)
│   ├── go.mod / go.sum
│   ├── cmd/daemon/main.go
│   └── internal/
│       ├── serial/doc.go     # blank import go.bug.st/serial
│       ├── wits/doc.go       # blank import shared/model
│       ├── storage/doc.go    # blank import go-sqlite3
│       └── sync/doc.go       # blank import paho MQTT
│
├── cloud-backend/            # Ingest + API cloud (STUB)
│   ├── go.mod / go.sum
│   ├── cmd/ingest/main.go
│   └── internal/
│       ├── broker/doc.go     # blank import paho MQTT
│       ├── store/doc.go      # blank import pgx + shared
│       └── api/doc.go        # blank import chi + websocket
│
├── simulator/                # Emulasi WITS (BERJALAN)
│   ├── requirements.txt
│   ├── wits_generator.py
│   ├── network_impairment.sh
│   └── .venv/                # virtualenv lokal (di-gitignore)
│
├── dashboard/                # UI well-log (DEMO DATA)
│   ├── package.json
│   ├── vite.config.ts
│   ├── index.html
│   └── src/
│       ├── App.tsx           # strip chart ECharts (sintetik)
│       ├── main.tsx
│       └── index.css
│
└── infrastructure/
    ├── mosquitto/
    │   └── mosquitto.conf
    └── timescale/
        └── 01_init.sql
```

---

## 5. Matriks Kematangan (As-Is vs Target)

| Komponen | Status As-Is | Keterangan |
|----------|--------------|------------|
| Struktur monorepo | ✅ Selesai | Folder, workspace, gitignore, Makefile |
| Dependency Go/Python/Node | ✅ Terpasang | `make init` sukses |
| Model shared `TelemetryPoint` | ✅ Ada | + `Validate()`, topic helpers, WITS IDs |
| Schema TimescaleDB | ✅ Ada | Hypertable `drilling_telemetry` |
| Config Mosquitto lokal | ✅ Ada | Anon + MQTT + WebSocket |
| Docker Compose | ✅ Terdefinisi | Butuh Docker terpasang di host |
| Simulator WITS | ✅ Siap | PTY/serial writer + CLI flags |
| Network impairment script | 🟡 Partial | Script `tc netem` siap; butuh root + iface |
| Edge daemon | ✅ Siap | serial → WITS → SQLite outbox → MQTT |
| Cloud backend | ✅ Siap | MQTT → TimescaleDB → REST + WebSocket |
| Dashboard | ✅ Live | REST seed + WebSocket well-log chart |
| Integrasi E2E | ✅ Siap | Alur lokal terdokumentasi di README |
| Auth / keamanan produksi | 🔴 Belum | Mosquitto anon; kredensial DB plain di compose |
| CI/CD | 🔴 Belum | Tidak ada pipeline CI |
| Tes otomatis | 🟡 Partial | Unit test WITS parser + MQTT topic helpers |

Legenda: ✅ siap · 🟡 sebagian · 🔴 belum diimplementasi

---

## 5.1 Runbook End-to-End (lokal)

```bash
make init
make infra-up

# Terminal A
make run-cloud

# Terminal B — catat SERIAL_PORT yang dicetak
make run-sim

# Terminal C
SERIAL_PORT=/dev/pts/N make run-edge

# Terminal D
make run-dashboard   # http://localhost:5173
```

Kontrak runtime:

| Layer | Kontrak |
|-------|---------|
| MQTT | `telemetry/{well_id}/points`, QoS 1, JSON `TelemetryPoint` |
| REST | `GET /api/v1/health`, `/wells`, `/wells/{id}/latest`, `/wells/{id}/telemetry` |
| WebSocket | `GET /api/v1/ws/telemetry?well_id=...` → `{type:"point", data:...}` |
| Env | lihat [`.env.example`](../.env.example) |

Struktur modul utama:

| Path | Isi |
|------|-----|
| `edge-daemon/internal/{config,serial,wits,storage,sync,pipeline}` | Pipeline edge modular |
| `cloud-backend/internal/{config,broker,store,hub,api}` | Ingest + API modular |
| `shared/{model,topic,witsids}` | Kontrak lintas service |
| `dashboard/src/{api,hooks,components}` | Klien live UI |

---

## 6. Penjelasan Detail per Komponen

### 6.1 `shared/` — Kontrak data bersama

**Modul:** `rigtelemetry/shared`  
**Tujuan:** Satu sumber kebenaran untuk struktur telemetry yang dipakai edge dan cloud.

Struct saat ini (`shared/model/telemetry.go`):

| Field | Tipe Go | JSON | Makna lapangan |
|-------|---------|------|----------------|
| `Time` | `time.Time` | `time` | Timestamp sampel |
| `WellID` | `string` | `well_id` | Identitas sumur |
| `BitDepth` | `float64` | `bit_depth` | Kedalaman bit (m) |
| `ROP` | `float64` | `rop` | Rate of Penetration |
| `WOB` | `float64` | `wob` | Weight on Bit |
| `GammaRay` | `float64` | `gamma_ray` | Gamma ray (API units) |

**Catatan as-is:** Belum ada skema Protobuf/Avro, validasi, atau versioning message. Serialisasi yang diharapkan ke depan kemungkinan JSON di MQTT (belum dikunci).

---

### 6.2 `edge-daemon/` — Agen edge di rig

**Modul:** `rigtelemetry/edge`  
**Entry point:** `cmd/daemon/main.go`

#### Target pipeline (belum diimplementasi)

```text
Serial reader → WITS parser → SQLite buffer → MQTT publisher (sync)
```

#### Package internal (kerangka saja)

| Package | Peran yang direncanakan | As-is |
|---------|-------------------------|-------|
| `internal/serial` | Baca byte stream dari serial/PTY | Blank import `go.bug.st/serial` |
| `internal/wits` | Parse frame WITS → `TelemetryPoint` | Blank import `shared/model` |
| `internal/storage` | Buffer lokal SQLite (store-and-forward) | Blank import `go-sqlite3` |
| `internal/sync` | Publish MQTT + retry/resync | Blank import `paho.mqtt.golang` |

#### Dependensi langsung

- `go.bug.st/serial` — komunikasi serial
- `github.com/mattn/go-sqlite3` — buffer embedded (**membutuhkan CGO**)
- `github.com/eclipse/paho.mqtt.golang` — klien MQTT
- `github.com/caarlos0/env/v11` — konfigurasi dari environment
- `rigtelemetry/shared` — model bersama

#### Perilaku saat ini

Menjalankan proses hanya mencetak:

```text
edge-daemon: starting (stub)
```

Lalu exit. Belum ada loop baca, parse, persist, atau publish.

---

### 6.3 `cloud-backend/` — Ingest & API cloud

**Modul:** `rigtelemetry/cloud`  
**Entry point:** `cmd/ingest/main.go`

#### Target pipeline (belum diimplementasi)

```text
MQTT subscriber → TimescaleDB writer → HTTP/WebSocket API → Dashboard
```

#### Package internal (kerangka saja)

| Package | Peran yang direncanakan | As-is |
|---------|-------------------------|-------|
| `internal/broker` | Subscribe topik MQTT dari edge | Blank import MQTT |
| `internal/store` | Insert ke hypertable TimescaleDB | Blank import `pgx/v5` + shared |
| `internal/api` | REST + WebSocket untuk dashboard | Blank import `chi` + `gorilla/websocket` |

#### Dependensi langsung

- `github.com/eclipse/paho.mqtt.golang`
- `github.com/jackc/pgx/v5`
- `github.com/go-chi/chi/v5`
- `github.com/gorilla/websocket`
- `rigtelemetry/shared`

#### Perilaku saat ini

```text
cloud-backend ingest: starting (stub)
```

Lalu exit.

---

### 6.4 `simulator/` — Emulasi sumber data WITS

#### `wits_generator.py` (berjalan)

Generator sintetik yang:

1. Menghasilkan sampel tiap 1 detik untuk well `WELL-DEMO-01`.
2. Menghitung `bit_depth`, `rop`, `wob`, `gamma_ray` dengan noise sinusoidal.
3. Membungkus ke frame ASCII gaya WITS Level 0:

```text
&&
0108<bit_depth>
0113<rop>
010A<wob>
0122<gamma_ray>
!!
```

4. Menampilkan live table di terminal memakai library `rich`.

**Batasan as-is:**

- Belum membuka serial port / PTY (`pyserial` sudah di-`requirements.txt` tapi belum dipakai).
- Belum mengirim ke edge-daemon.
- Mapping item ID WITS bersifat ilustratif (bukan dictionary WITS resmi yang lengkap).

#### `network_impairment.sh`

Helper Linux `tc netem` untuk mensimulasikan uplink buruk:

- `apply <iface>` → delay 100ms±20ms + loss 2%
- `clear <iface>` → hapus qdisc

Membutuhkan privilege jaringan (`root` / `CAP_NET_ADMIN`).

#### Dependensi Python

```text
pyserial>=3.5
rich>=13.7.0
```

Virtualenv default: `simulator/.venv` (dibuat oleh `make init`).

---

### 6.5 `dashboard/` — Visualisasi well-log

**Stack:** Vite 6 + React 19 + TypeScript 5.7 + Tailwind CSS 4 + ECharts + lucide-react

#### As-is UI

- Halaman tunggal dengan judul “Rig Telemetry Dashboard”.
- Strip chart vertikal: **Gamma Ray (X) vs Bit Depth (Y, inverted)** — pola khas well-log.
- Data dari array demo di memori (40 titik), **bukan** dari API/WebSocket cloud.

#### Scripts npm

| Script | Fungsi |
|--------|--------|
| `npm run dev` | Dev server Vite (port 5173) |
| `npm run build` | Typecheck + production build |
| `npm run preview` | Preview hasil build |

#### Belum ada

- Koneksi ke cloud API / WebSocket
- Multi-curve (ROP, WOB, dll.)
- Pemilihan well / rentang waktu
- Auth

---

### 6.6 Infrastruktur lokal

#### Docker Compose (`docker-compose.yml`)

| Service | Image | Port | Catatan |
|---------|-------|------|---------|
| `timescaledb` | `timescale/timescaledb:latest-pg16` | `5433→5432` | User/DB/password: `telemetry` (host 5433 avoids local Postgres on 5432) |
| `mqtt-broker` | `eclipse-mosquitto:2` | `1883`, `9001` | Config mount dari repo |

Volume persistent: `timescale_data`.  
Init SQL di-mount ke `/docker-entrypoint-initdb.d/01_init.sql` (hanya dijalankan saat volume DB **baru**).

#### Mosquitto (`infrastructure/mosquitto/mosquitto.conf`)

- Listener MQTT plain `:1883`
- Listener WebSockets `:9001`
- `allow_anonymous true`
- `persistence false`

**Hanya untuk development lokal — tidak aman untuk produksi.**

#### TimescaleDB schema (`infrastructure/timescale/01_init.sql`)

Tabel `drilling_telemetry`:

| Kolom | Tipe |
|-------|------|
| `time` | `TIMESTAMPTZ NOT NULL` |
| `well_id` | `TEXT NOT NULL` |
| `bit_depth` | `DOUBLE PRECISION` |
| `rop` | `DOUBLE PRECISION` |
| `wob` | `DOUBLE PRECISION` |
| `gamma_ray` | `DOUBLE PRECISION` |

Ditambah:

- `create_hypertable('drilling_telemetry', 'time')`
- Index `(well_id, time DESC)`

---

## 7. Model Data End-to-End (Kontrak yang Sudah Ada)

Ketiga lapisan sudah **selaras secara konseptual** untuk field inti:

| Lapisan | Representasi |
|---------|--------------|
| Shared Go | `model.TelemetryPoint` |
| TimescaleDB | kolom `drilling_telemetry` |
| Simulator | dict + frame WITS item `0108/0113/010A/0122` |
| Dashboard (demo) | hanya Gamma Ray + Depth |

Ini memudahkan implementasi berikutnya: parse WITS → struct → JSON MQTT → INSERT SQL → chart series.

---

## 8. Tooling Operasional

### 8.1 Makefile

| Target | Fungsi |
|--------|--------|
| `make init` | `go mod tidy` (3 modul) + venv Python + `npm install` |
| `make infra-up` | `docker compose up -d` |
| `make infra-down` | `docker compose down` |
| `make run-sim` | Jalankan WITS generator via venv |
| `make run-edge` | `go run ./cmd/daemon` |
| `make run-cloud` | `go run ./cmd/ingest` |

### 8.2 Prerequisites yang diharapkan

- Go **1.22+**
- Python **3.10+** (lingkungan bootstrap memakai 3.12)
- Node.js **20+** / npm (bootstrap memakai Node 22)
- Docker & Docker Compose (untuk infra)
- Compiler C / CGO (untuk `go-sqlite3` saat build edge)

### 8.3 `.gitignore` (cakupan utama)

- Binary Go, `node_modules`, venv Python
- SQLite: `*.db`, `*.db-wal`, `*.db-shm`
- `.env`, log, artefak IDE/OS
- `go.work.sum` (diabaikan di VCS)

---

## 9. Alur Pengembangan Lokal yang Disarankan (Saat Ini)

```bash
# 1. Pasang dependency
make init

# 2. Nyalakan DB + MQTT (butuh Docker)
make infra-up

# 3. Di terminal terpisah
make run-sim      # lihat frame WITS di terminal
make run-edge     # stub
make run-cloud    # stub
cd dashboard && npm run dev   # UI demo di :5173
```

**Harapan realistis as-is:** Anda bisa melihat simulator dan dashboard demo secara mandiri. Edge/cloud belum saling bicara dan belum menyentuh DB/MQTT.

---

## 10. Gap Analisis: Apa yang Belum Ada

Pipeline inti sudah selesai. Yang masih terbuka:

### Prioritas menengah (kualitas)

1. Idempotency insert (hindari duplikat saat retry MQTT).
2. Health/metrics yang lebih kaya + structured logging.
3. Integration test Compose (butuh Docker di CI/host).
4. Multi-curve dashboard (ROP/WOB tracks) selain gamma strip.

### Prioritas produksi

5. Auth MQTT, TLS, secret management.
6. Hardening Mosquitto & DB credentials.
7. Observability (log terstruktur, tracing).
8. CI (build Go/Node, lint, test).
9. Deployment artifacts (Dockerfile per service).

---

## 11. Keputusan Desain yang Sudah Terkunci

| Keputusan | Pilihan as-is |
|-----------|----------------|
| Bentuk repo | Monorepo polyglot |
| Bahasa edge/cloud | Go 1.22 + workspaces |
| Shared contracts | Modul Go internal (`rigtelemetry/shared`) |
| Buffer edge | SQLite (`go-sqlite3`) |
| Transport cloud sync | MQTT (Eclipse Paho + Mosquitto) |
| TSDB | TimescaleDB di atas PostgreSQL 16 |
| HTTP router | go-chi |
| Realtime UI transport (rencana) | WebSocket (gorilla) |
| Simulator | Python + rich (+ pyserial nanti) |
| Frontend | React + Vite + Tailwind + ECharts |
| Orkestrasi lokal | Docker Compose + Makefile |

---

## 12. Risiko & Catatan Teknis As-Is

| Risiko / catatan | Dampak |
|------------------|--------|
| Stub blank-import | Dependensi “langsung” ada, tapi belum ada logika bisnis |
| `go-sqlite3` + CGO | Build edge gagal tanpa toolchain C |
| Init SQL hanya sekali | Ubah schema setelah volume ada → perlu migrate manual / reset volume |
| Mosquitto anonymous | Aman hanya di localhost |
| Image `latest-pg16` | Tag `latest` bisa berubah; pin versi untuk reproducibility |
| Docker belum tersedia di semua mesin | `make infra-up` tidak bisa diverifikasi tanpa instalasi Docker |
| WITS item ID ilustratif | Perlu validasi terhadap dictionary WITS aktual sebelum produksi |
| Tidak ada topik MQTT / API contract tertulis | Perlu disepakati sebelum wiring |

---

## 13. Glosarium Singkat

| Istilah | Arti |
|---------|------|
| **WITS** | Wellsite Information Transfer Specification — format transfer data wellsite |
| **ROP** | Rate of Penetration — laju pengeboran |
| **WOB** | Weight on Bit — beban pada bit |
| **Gamma Ray** | Log radiasi alami formasi; umum pada strip chart |
| **Bit Depth** | Kedalaman bit bor |
| **Edge** | Komputasi di dekat sumber data (rig), bukan di cloud |
| **Store-and-forward** | Simpan lokal dulu, kirim kemudian saat jaringan tersedia |
| **Hypertable** | Tabel TimescaleDB yang dipartisi otomatis berdasarkan waktu |
| **PTY** | Pseudo-terminal — sering dipakai untuk emulasi serial di Linux |

---

## 14. Referensi Cepat File Penting

| Keperluan | File |
|-----------|------|
| Quick start | `README.md` |
| Workspace Go | `go.work` |
| Model telemetry | `shared/model/telemetry.go` |
| Schema DB | `infrastructure/timescale/01_init.sql` |
| Config MQTT | `infrastructure/mosquitto/mosquitto.conf` |
| Compose lokal | `docker-compose.yml` |
| Target make | `Makefile` |
| Generator WITS | `simulator/wits_generator.py` |
| Stub edge | `edge-daemon/cmd/daemon/main.go` |
| Stub cloud | `cloud-backend/cmd/ingest/main.go` |
| UI demo | `dashboard/src/App.tsx` |

---

## 15. Kesimpulan As-Is

Proyek sudah melewati fase bootstrap: **pipeline end-to-end lokal berfungsi** dengan arsitektur modular (edge outbox store-and-forward, cloud ingest + hub, dashboard live).

Yang masih terbuka untuk fase berikutnya: auth/TLS, CI, idempotency/dedup yang lebih ketat, multi-well UI, dan hardening produksi.

Dokumen ini mencerminkan kondisi aktual repositori setelah wiring E2E.
