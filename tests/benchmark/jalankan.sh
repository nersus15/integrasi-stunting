#!/usr/bin/env bash
# Benchmark integrasi-stunting dalam simulasi satu server terbatas.
#
#   ./jalankan.sh build                  kompilasi app (musl) + image box
#   ./jalankan.sh siapkan                migrasi + seed volume awal (tanpa batas sumber daya)
#   ./jalankan.sh uji http|kafka|campuran|semua|profil
#   ./jalankan.sh bersihkan              hapus container, volume, dan topic benchmark
#
# Spesifikasi server: spek-server.env. Profil beban: beban.env.
set -euo pipefail

B="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$B/../.." && pwd)"
BUILD="$B/.build"

BOX=bench-stunting-box
IMAGE=bench-stunting-box:local
BUILDER=bench-stunting-builder:local
VOL=bench-stunting-pg
VOL_AWAL=bench-stunting-pg-awal

export PG_PORT=25432 REDIS_PORT=26379 APP_PORT=22025
export KAFKA_BROKERS=localhost:1092 KAFKA_TOPIC=bench-stunting-fhir KAFKA_GROUP=bench-stunting

# nilai di file adalah bawaan; variabel yang sudah diset di shell menang
SPEK=()
muat_env() {
	local k v
	while IFS='=' read -r k v; do
		[[ "$k" =~ ^[A-Z_0-9]+$ ]] || continue
		if [ -z "${!k+x}" ]; then export "$k=$v"; fi
		if [ "$2" = spek ]; then SPEK+=(-e "$k"); fi
	done <"$1"
}
muat_env "$B/spek-server.env" spek
muat_env "$B/beban.env" beban

log() { printf '\033[1m==> %s\033[0m\n' "$*"; }

build() {
	mkdir -p "$BUILD/gocache"
	log "image builder"
	docker build -q -t "$BUILDER" - >/dev/null <<-'EOF'
		FROM golang:1.26-alpine
		RUN apk add --no-cache build-base pkgconfig
	EOF
	log "kompilasi app (musl, librdkafka statis)"
	docker run --rm --user "$(id -u):$(id -g)" \
		-v "$ROOT:/src" -v "$(go env GOMODCACHE):/gomod:ro" -w /src \
		-e HOME=/tmp -e GOMODCACHE=/gomod -e GOCACHE=/src/tests/benchmark/.build/gocache \
		-e GOPROXY=off -e GOTOOLCHAIN=local -e GOFLAGS=-buildvcs=false -e CGO_ENABLED=1 \
		-e CGO_CFLAGS="-D_LARGEFILE64_SOURCE -D_GNU_SOURCE" \
		"$BUILDER" go build -tags musl -o tests/benchmark/.build/webcore ./app
	log "image box"
	docker build -q -t "$IMAGE" "$B/box" >/dev/null
}

kunci() {
	if [ ! -f "$BUILD/kunci.env" ]; then
		mkdir -p "$BUILD"
		printf 'KUNCI_JAKANTRO=bench-%s\nKUNCI_FASKES_PREFIX=bench-%s-\n' \
			"$(head -c16 /dev/urandom | xxd -p)" "$(head -c8 /dev/urandom | xxd -p)" >"$BUILD/kunci.env"
	fi
	set -a; . "$BUILD/kunci.env"; set +a
}

box_run() {
	docker rm -f "$BOX" >/dev/null 2>&1 || true
	docker run --name "$BOX" --network host --shm-size=256m \
		"${SPEK[@]}" \
		-e PG_PORT -e REDIS_PORT -e APP_PORT -e KAFKA_BROKERS -e KAFKA_TOPIC -e KAFKA_GROUP \
		-e KUNCI_JAKANTRO -e KUNCI_FASKES_PREFIX \
		-v "$ROOT/migration/stunting/postgres:/bench/migration:ro" \
		-v "$B/box:/bench/box:ro" -v "$BUILD:/bench/build:ro" \
		"$@"
}

batas() {
	echo --cpus "$SERVER_CPUS" --memory "$SERVER_MEMORY" --memory-swap "$SERVER_MEMORY" \
		--device-read-iops "$DISK_DEVICE:$DISK_READ_IOPS" --device-write-iops "$DISK_DEVICE:$DISK_WRITE_IOPS" \
		--device-read-bps "$DISK_DEVICE:${DISK_READ_MBPS}mb" --device-write-bps "$DISK_DEVICE:${DISK_WRITE_MBPS}mb"
}

siapkan() {
	kunci
	docker volume rm -f "$VOL_AWAL" >/dev/null
	docker volume create "$VOL_AWAL" >/dev/null
	log "migrasi + seed $SEED_ANAK anak x $SEED_KUNJUNGAN_PER_ANAK kunjungan (tanpa batas)"
	time box_run -v "$VOL_AWAL:/var/lib/postgresql/data" "$IMAGE" siapkan
	docker rm -f "$BOX" >/dev/null
}

volume_segar() {
	docker rm -f "$BOX" >/dev/null 2>&1 || true
	docker volume rm -f "$VOL" >/dev/null
	docker volume create "$VOL" >/dev/null
	docker run --rm -v "$VOL_AWAL:/dari:ro" -v "$VOL:/ke" --entrypoint sh "$IMAGE" -c 'cp -a /dari/. /ke/'
}

nyalakan() {
	log "box: $SERVER_CPUS CPU, RAM $SERVER_MEMORY, disk ${DISK_READ_IOPS}/${DISK_WRITE_IOPS} IOPS, fsync ${DISK_FSYNC_MS} ms"
	# shellcheck disable=SC2046
	box_run -d $(batas) -v "$VOL:/var/lib/postgresql/data" "$IMAGE" layani >/dev/null
	for _ in $(seq 1 120); do
		curl -sf -o /dev/null "http://127.0.0.1:$APP_PORT/stunting/health" && return 0
		docker inspect "$BOX" >/dev/null 2>&1 || { log "box berhenti"; return 1; }
		sleep 0.5
	done
	docker logs --tail 30 "$BOX"
	return 1
}

kalibrasi() {
	log "kalibrasi disk di dalam batas"
	docker exec "$BOX" /bench/box/entrypoint.sh kalibrasi | tee "$1/kalibrasi.txt"
}

uji() {
	local mode="$1"
	kunci
	hasil="$B/hasil/$(date +%Y%m%d-%H%M%S)-$mode"
	mkdir -p "$hasil"
	env | grep -E '^(SERVER|DISK|PG|DB|APP|RATE|SEED|HTTP|KAFKA|DURASI|MAKS|BATAS|CAMPURAN|BOBOT)_' | sort >"$hasil/env-dipakai.txt"

	if [ "$mode" = profil ]; then export PG_PRELOAD=pg_stat_statements; fi
	volume_segar
	log "topic $KAFKA_TOPIC ($KAFKA_PARTISI partisi)"
	(cd "$ROOT/tests" && GOWORK=off go run ./benchmark/beban -siapkan-topic)
	nyalakan
	trap 'docker logs "$BOX" >"$hasil/box.log" 2>&1 || true; docker rm -f "$BOX" >/dev/null 2>&1 || true' EXIT
	kalibrasi "$hasil"

	log "beban: $mode"
	(cd "$ROOT/tests" && ROOT="$ROOT" GOWORK=off go run ./benchmark/beban \
		-mode "$mode" -hasil "$hasil" -box "$(docker inspect -f '{{.Id}}' "$BOX")" \
		-base "http://127.0.0.1:$APP_PORT/api" \
		-pg "postgres://postgres:postgres@127.0.0.1:$PG_PORT/ildki?sslmode=disable")
	log "hasil: $hasil"
}

bersihkan() {
	docker rm -f "$BOX" >/dev/null 2>&1 || true
	docker volume rm -f "$VOL" "$VOL_AWAL" >/dev/null
	(cd "$ROOT/tests" && GOWORK=off go run ./benchmark/beban -hapus-topic) || true
}

case "${1:-}" in
build) build ;;
siapkan) siapkan ;;
uji) uji "${2:?mode: http|kafka|campuran|semua|profil}" ;;
bersihkan) bersihkan ;;
*) sed -n '2,9p' "$0"; exit 2 ;;
esac
