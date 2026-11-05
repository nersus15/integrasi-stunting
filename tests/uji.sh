#!/usr/bin/env bash
# Menjalankan test per bagian, dengan env dari .test.env. Key API test dibuat otomatis (testutil/keys.go).
#
#   ./uji.sh daftar                          daftar bagian beserta kebutuhannya
#   ./uji.sh <bagian>... [argumen go test]   mis. ./uji.sh update_endpoint -run PUTFaskes -v
#   ./uji.sh unit | integration | functional | semua
#
# Keluaran diringkas: daftar lulus/gagal dan pesan kegagalan. Log aplikasi ditulis ke
# tests/.log/<kelompok>-<paket>.log; UJI_LOG=stdout menampilkannya, -v memberi keluaran mentah.
#
# <bagian> boleh nama folder (update_endpoint) atau path (functional/update_endpoint).
# Nama yang ada di lebih dari satu kelompok (find_orangtua) harus memakai path.
set -euo pipefail

T="$(cd "$(dirname "$0")" && pwd)"
cd "$T"

# env dari shell tetap menang atas .test.env
muat_env() {
	[ -f "$1" ] || return 0
	local baris k v
	while IFS= read -r baris || [ -n "$baris" ]; do
		baris="${baris%$'\r'}"
		[[ "$baris" =~ ^[[:space:]]*(#|$) ]] && continue
		[[ "$baris" == *=* ]] || continue
		baris="${baris#export }"
		k="${baris%%=*}"
		v="${baris#*=}"
		k="${k//[[:space:]]/}"
		[ -n "${!k+x}" ] && continue
		v="${v%[\"\']}"
		v="${v#[\"\']}"
		export "$k=$v"
	done <"$1"
}
muat_env .test.env

daftar() {
	printf '%-36s %s\n' "BAGIAN" "BUTUH"
	for d in unit/* integration/* functional/*; do
		[ -d "$d" ] || continue
		local butuh="-"
		if grep -qsE 'DatabaseAvailable|OpenDB|StartApp|PG_DSN' "$d"/*_test.go; then butuh="database"; fi
		printf '%-36s %s\n' "$d" "$butuh"
	done
}

cari() {
	case "$1" in
	unit | integration | functional) echo "./$1/..." ;;
	semua) echo "./..." ;;
	*/*) [ -d "$1" ] && echo "./$1" || { echo "bagian tidak ada: $1" >&2; return 1; } ;;
	*)
		local hasil
		hasil=$(ls -d unit/"$1" integration/"$1" functional/"$1" 2>/dev/null || true)
		case "$(echo "$hasil" | grep -c .)" in
		1) echo "./$hasil" ;;
		0) echo "bagian tidak ada: $1 (lihat ./uji.sh daftar)" >&2; return 1 ;;
		*) echo "bagian $1 ambigu, pakai salah satu: $(echo $hasil)" >&2; return 1 ;;
		esac
		;;
	esac
}

case "${1:-}" in
"" | -h | --help) sed -n '2,10p' "$0"; exit 0 ;;
daftar) daftar; exit 0 ;;
esac

paket=()
while [ $# -gt 0 ] && [[ "$1" != -* ]]; do
	paket+=("$(cari "$1")")
	shift
done

# -p 1 wajib: semua paket berbagi satu database dev
if [[ " $* " == *" -v "* ]]; then
	exec env GOWORK=off go test -p 1 -count=1 "$@" "${paket[@]}"  # keluaran mentah go test
fi

# -json lalu diringkas: hasil per test dan pesan kegagalan saja, log aplikasi ada di tests/.log/
rinci=()
if [ "${#paket[@]}" -eq 1 ] && [[ "${paket[0]}" != *"..." ]]; then rinci=(-daftar); fi
GOWORK=off go build -o .log/ringkas ./alat/ringkas
set +e
GOWORK=off go test -json -p 1 -count=1 "$@" "${paket[@]}" 2>&1 | .log/ringkas "${rinci[@]}"
status=${PIPESTATUS[0]}
find .log -name '*.log' -empty -delete 2>/dev/null
exit "$status"
