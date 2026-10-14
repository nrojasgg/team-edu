#!/usr/bin/env bash
#
# Portable MongoDB backup script.
# Attempts in order:
#  1) Use local `mongodump` if available.
#  2) Use `docker exec` against common mongo container names.
#  3) Use `docker compose exec` / `docker-compose exec` if compose project is running.
#
# Usage:
#  ./backup_mongo.sh [-u <mongo-uri>] [-o <output-dir>] [-n <name>] [-c <container-name>]
#
# Examples:
#  ./backup_mongo.sh                             # auto-detect, output to ./backups
#  ./backup_mongo.sh -u "mongodb://host:27017"  # explicit URI
#  ./backup_mongo.sh -c mongo_db                 # use container name
#
set -euo pipefail

# Configuration / defaults
DEFAULT_OUTPUT_DIR="./backups"
TIMESTAMP="$(date +%Y%m%dT%H%M%S)"
OUTPUT_DIR="${DEFAULT_OUTPUT_DIR}"
BACKUP_NAME="mongo-backup-${TIMESTAMP}"
MONGO_URI="${MONGO_URI:-}"    # user can set env var MONGO_URI
CONTAINER_NAME="${CONTAINER_NAME:-}" # optional override
QUIET=0

# Common candidate container names used in compose files
COMMON_MONGO_NAMES=( "mongo_db" "mongo-db" "mongo" "mongodb" "mongo_db_1" )

# Helpers
log() {
  if [ "$QUIET" -eq 0 ]; then
    printf '%s\n' "$*"
  fi
}
err() {
  printf 'ERROR: %s\n' "$*" >&2
}
usage() {
  cat <<EOF
Usage: $0 [options]

Options:
  -u <mongo-uri>        MongoDB URI (e.g. mongodb://user:pass@host:27017/db)
  -o <output-dir>       Output directory for backups (default: ${DEFAULT_OUTPUT_DIR})
  -n <name>             Base name for backup file (default: ${BACKUP_NAME})
  -c <container-name>   Docker container name to use for docker exec
  -q                    Quiet
  -h                    Show this help
EOF
}

check_cmd() {
  command -v "$1" >/dev/null 2>&1
}

ensure_output_dir() {
  mkdir -p "${OUTPUT_DIR}"
}

compress_archive() {
  local src="$1"
  local dest="$2"
  # prefer tar+gz
  if check_cmd tar; then
    tar -C "$(dirname "$src")" -czf "${dest}" "$(basename "$src")"
  else
    # fallback: gzip single file if it's an archive file
    if check_cmd gzip; then
      gzip -c "${src}" > "${dest}"
    else
      # no compressor available; just move
      cp -a "${src}" "${dest}"
    fi
  fi
}

try_mongodump() {
  # Attempt to use local mongodump binary
  if ! check_cmd mongodump; then
    return 1
  fi

  log "Found local 'mongodump'. Creating backup..."

  ensure_output_dir

  # prefer --archive to produce a single file
  local out_file="${OUTPUT_DIR}/${BACKUP_NAME}.archive.gz"

  if [ -n "${MONGO_URI}" ]; then
    # use URI if provided
    if mongodump --version >/dev/null 2>&1; then
      # try archive + gzip
      if mongodump --uri="${MONGO_URI}" --archive="${out_file}" --gzip >/dev/null 2>&1; then
        log "Backup created: ${out_file}"
        return 0
      else
        err "mongodump failed with --archive; will try directory dump fallback"
      fi
    fi
  fi

  # fallback to directory dump (if no URI given or archive failed)
  local tmp_dir="${OUTPUT_DIR}/${BACKUP_NAME}.d"
  rm -rf "${tmp_dir}"
  mkdir -p "${tmp_dir}"

  if [ -n "${MONGO_URI}" ]; then
    mongodump --uri="${MONGO_URI}" --out="${tmp_dir}" || return 1
  else
    mongodump --out="${tmp_dir}" || return 1
  fi

  # compress directory
  local tarfile="${OUTPUT_DIR}/${BACKUP_NAME}.tar.gz"
  if check_cmd tar; then
    tar -C "${OUTPUT_DIR}" -czf "${tarfile}" "$(basename "${tmp_dir}")"
    rm -rf "${tmp_dir}"
    log "Backup created: ${tarfile}"
    return 0
  else
    log "Backup created as directory: ${tmp_dir}"
    return 0
  fi
}

_try_docker_exec_once() {
  # parameters: container, dest_dir
  local container="$1"
  local dest_dir="$2"

  # We will create an archive inside the container using mongodump --archive and stream it out
  # If mongodump exists inside container, use it. Otherwise try running mongodump from host into container's mongodb socket (not covered here).
  if docker exec "${container}" sh -c 'command -v mongodump >/dev/null 2>&1' >/dev/null 2>&1; then
    log "Using 'docker exec ${container} mongodump --archive --gzip'..."
    # stream archive to host
    docker exec "${container}" mongodump --archive --gzip 2>/dev/null | cat > "${dest_dir}/${BACKUP_NAME}.archive.gz"
    return 0
  fi

  # try mongodump inside container by full path (some images install in /usr/bin or /usr/local/bin)
  if docker exec "${container}" bash -lc 'command -v mongodump >/dev/null 2>&1' >/dev/null 2>&1; then
    docker exec "${container}" bash -lc 'mongodump --archive --gzip' 2>/dev/null | cat > "${dest_dir}/${BACKUP_NAME}.archive.gz"
    return 0
  fi

  return 1
}

try_docker_exec() {
  # Try to find a running container and run mongodump inside it, streaming archive to host.
  if ! check_cmd docker; then
    return 1
  fi

  log "docker present. Looking for a mongo container..."

  local candidates=()
  if [ -n "${CONTAINER_NAME}" ]; then
    candidates=( "${CONTAINER_NAME}" )
  else
    candidates=( "${COMMON_MONGO_NAMES[@]}" )
  fi

  ensure_output_dir

  for c in "${candidates[@]}"; do
    # check container exists and is running
    if docker ps --format '{{.Names}}' | grep -xq "${c}"; then
      log "Found running container '${c}', attempting backup..."
      if _try_docker_exec_once "${c}" "${OUTPUT_DIR}"; then
        log "Backup created: ${OUTPUT_DIR}/${BACKUP_NAME}.archive.gz"
        return 0
      else
        err "Failed to run mongodump inside container '${c}'"
      fi
    fi
  done

  # No running container found by name; attempt to use docker ps to find images with 'mongo' in name
  local candidate_from_ps
  candidate_from_ps="$(docker ps --format '{{.Names}} {{.Image}}' | awk '/mongo/ {print $1; exit}') || true"
  if [ -n "${candidate_from_ps}" ]; then
    log "Using detected container '${candidate_from_ps}'"
    if _try_docker_exec_once "${candidate_from_ps}" "${OUTPUT_DIR}"; then
      log "Backup created: ${OUTPUT_DIR}/${BACKUP_NAME}.archive.gz"
      return 0
    fi
  fi

  return 1
}

try_docker_compose_exec() {
  # Try both `docker compose exec` and `docker-compose exec` against common service names
  local dc_cmd=""
  if check_cmd docker && docker compose version >/dev/null 2>&1; then
    dc_cmd="docker compose exec -T"
  elif check_cmd docker-compose; then
    dc_cmd="docker-compose exec -T"
  else
    return 1
  fi

  log "Trying compose exec via: ${dc_cmd}"

  local candidates=()
  if [ -n "${CONTAINER_NAME}" ]; then
    candidates=( "${CONTAINER_NAME}" )
  else
    candidates=( "${COMMON_MONGO_NAMES[@]}" )
  fi

  ensure_output_dir

  for svc in "${candidates[@]}"; do
    # attempt exec; some compose setups require service name, not container name
    if ${dc_cmd} "${svc}" sh -c 'command -v mongodump >/dev/null 2>&1' >/dev/null 2>&1; then
      log "Using compose service '${svc}', running mongodump..."
      ${dc_cmd} "${svc}" mongodump --archive --gzip 2>/dev/null | cat > "${OUTPUT_DIR}/${BACKUP_NAME}.archive.gz"
      log "Backup created: ${OUTPUT_DIR}/${BACKUP_NAME}.archive.gz"
      return 0
    fi
  done

  return 1
}

main() {
  # parse args
  while getopts ":u:o:n:c:qh" opt; do
    case "${opt}" in
      u) MONGO_URI="${OPTARG}" ;;
      o) OUTPUT_DIR="${OPTARG}" ;;
      n) BACKUP_NAME="${OPTARG}" ;;
      c) CONTAINER_NAME="${OPTARG}" ;;
      q) QUIET=1 ;;
      h) usage; exit 0 ;;
      \?) err "Invalid option -${OPTARG}" ; usage ; exit 2 ;;
    esac
  done
  shift $((OPTIND-1))

  # Short-circuit: if user provided a full archive filename via BACKUP_NAME including extension, keep it
  TIMESTAMP="$(date +%Y%m%dT%H%M%S)"
  if [ -z "${BACKUP_NAME}" ]; then
    BACKUP_NAME="mongo-backup-${TIMESTAMP}"
  fi

  # Try order of operations
  log "Starting MongoDB backup procedure..."
  log "Output directory: ${OUTPUT_DIR}"
  [ -n "${MONGO_URI}" ] && log "Using MONGO_URI from param/env"

  # 1) Try local mongodump
  if try_mongodump; then
    log "Backup completed using local mongodump."
    return 0
  fi
  log "Local mongodump not available or failed; trying docker exec..."

  # 2) Try docker exec against mongo container
  if try_docker_exec; then
    log "Backup completed using docker exec (mongodump inside container)."
    return 0
  fi
  log "docker exec approach failed; trying docker compose exec..."

  # 3) Try docker-compose / docker compose exec
  if try_docker_compose_exec; then
    log "Backup completed using docker compose exec."
    return 0
  fi

  err "All backup methods failed. Please ensure you have either 'mongodump' installed locally or a running mongo container accessible via Docker / Docker Compose."
  return 3
}

# Run
main "$@"
