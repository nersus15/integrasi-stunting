#!/bin/sh
# Satu "server": postgres + redis + app dalam satu cgroup.
# Mode: siapkan (migrasi + seed lalu keluar) | layani (jalankan app) | kalibrasi
set -eu

PGDATA=/var/lib/postgresql/data/pg
export PGDATA
PSQL="psql -h 127.0.0.1 -p $PG_PORT -U postgres -v ON_ERROR_STOP=1 -q"

mulai_postgres() {
	if [ ! -s "$PGDATA/PG_VERSION" ]; then
		mkdir -p "$PGDATA" && chown postgres:postgres "$PGDATA"
		su-exec postgres initdb -D "$PGDATA" -U postgres --auth=trust >/dev/null
	fi
	su-exec postgres postgres -D "$PGDATA" \
		-c listen_addresses=127.0.0.1 -c port="$PG_PORT" \
		-c shared_buffers="$PG_SHARED_BUFFERS" \
		-c effective_cache_size="$PG_EFFECTIVE_CACHE_SIZE" \
		-c work_mem="$PG_WORK_MEM" \
		-c max_connections="$PG_MAX_CONNECTIONS" \
		-c random_page_cost="$PG_RANDOM_PAGE_COST" \
		-c effective_io_concurrency="$PG_EFFECTIVE_IO_CONCURRENCY" \
		-c synchronous_commit="$PG_SYNCHRONOUS_COMMIT" \
		-c max_wal_size="$PG_MAX_WAL_SIZE" \
		-c commit_delay="$((DISK_FSYNC_MS * 1000))" -c commit_siblings=0 \
		-c track_io_timing=on -c log_min_messages=warning \
		${PG_PRELOAD:+-c shared_preload_libraries=$PG_PRELOAD} &
	until pg_isready -h 127.0.0.1 -p "$PG_PORT" -q; do sleep 0.2; done
}

# migrasi stunting dijalankan app sendiri (goose) saat start, sama seperti di server
siapkan() {
	$PSQL -d postgres -c "DROP DATABASE IF EXISTS ildki" -c "CREATE DATABASE ildki"
	$PSQL -d ildki -f /bench/box/access.sql
	mulai_app &
	app=$!
	until wget -qO- "http://127.0.0.1:$APP_PORT/stunting/health" >/dev/null 2>&1; do sleep 0.5; done
	kill "$app"
	wait "$app" 2>/dev/null || true
	$PSQL -d ildki \
		-v n_anak="$SEED_ANAK" -v kunj_per_anak="$SEED_KUNJUNGAN_PER_ANAK" \
		-v n_pkm="$SEED_PUSKESMAS" -v n_rs="$SEED_RS" -v persen_rujukan="$SEED_PERSEN_RUJUKAN" \
		-v kunci_jakantro="$KUNCI_JAKANTRO" -v kunci_prefix="$KUNCI_FASKES_PREFIX" \
		-f /bench/box/seed.sql
	$PSQL -d ildki -c "VACUUM ANALYZE"
	su-exec postgres pg_ctl -D "$PGDATA" -m fast stop >/dev/null
}

tulis_config() {
	mkdir -p /run/app
	cat >/run/app/config.yaml <<EOF
app:
  name: integrasi-stunting
  environment: production
  features:
    recovery: true
  logging:
    level: $APP_LOG_LEVEL
    format: "\${time} | \${status} | \${latency} | \${ip} | \${method} | \${path} | \${error}\n"
    output: stdout
  cors:
    allow_origins:
      - "http://localhost:$APP_PORT"
    allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers: ["X-API-Key", "Origin", "Content-Type", "Accept", "Authorization"]
    allow_credentials: true
    expose_headers: ["Content-Length"]
    max_age: 24h
  rate_limit:
    enabled: $RATE_LIMIT_ENABLED
    max: $RATE_LIMIT_MAX
    window: 1m
    backend: memory
    fail_open: true
server:
  host: 127.0.0.1
  port: $APP_PORT
  path: "/api"
  read_timeout: 30s
  write_timeout: 30s
auth:
  type: apikey
  control: RBAC
  store: db
database:
  driver: postgres
  host: 127.0.0.1
  port: $PG_PORT
  user: postgres
  password: postgres
  name: ildki
  schema_name: stunting
  ssl_mode: disable
  max_idle_conns: $DB_MAX_IDLE_CONNS
  max_open_conns: $DB_MAX_OPEN_CONNS
  conn_max_lifetime: 60m
redis:
  host: 127.0.0.1
  port: $REDIS_PORT
  db: 0
kafka:
  enabled: true
  brokers:
    - "$KAFKA_BROKERS"
  group_id: "$KAFKA_GROUP"
  offset_reset: earliest
  topics:
    - "$KAFKA_TOPIC"
EOF
}

mulai_app() {
	redis-server --bind 127.0.0.1 --port "$REDIS_PORT" --save "" --appendonly no --daemonize yes >/dev/null
	tulis_config
	mkdir -p /run/app/migration/stunting
	ln -sfn /bench/migration /run/app/migration/stunting/postgres
	cd /run/app
	exec /bench/build/webcore stunting
}

# fsync di bawah batas IOPS; commit_delay (DISK_FSYNC_MS) tidak ikut terukur di sini
kalibrasi() {
	pg_test_fsync -f "$PGDATA/../fsync.tmp" -s 3 | grep -E "^ +(open_datasync|fdatasync|fsync) " | head -3
	rm -f "$PGDATA/../fsync.tmp"
}

case "${1:-layani}" in
siapkan) mulai_postgres; siapkan ;;
kalibrasi) kalibrasi ;;
layani) mulai_postgres; mulai_app ;;
*) echo "mode tidak dikenal: $1" >&2; exit 2 ;;
esac
