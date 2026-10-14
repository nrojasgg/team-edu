#!/usr/bin/env sh
# Portable Postgres backup script (auth and courses DBs)
# - Preferred method: run pg_dump inside the postgres containers (docker exec)
# - Fallback: use local pg_dump connecting to exposed ports
#
# Defaults are tailored to this repository's docker-compose.yml:
#  - auth container name: auth_db  (port 5432)
#  - courses container name: courses_db (port 5433)
#  - DB names: auth_db, courses_db
#
# Output: ./backups/postgres-<service>-YYYYmmdd_HHMMSS.sql.gz
#
# Usage:
#  ./scripts/backup_postgres.sh          # backups both DBs to ./backups
#  ./scripts/backup_postgres.sh -o out   # custom output dir
#  ./scripts/backup_postgres.sh -p pass  # override postgres password used for connection
#  ./scripts/backup_postgres.sh -h       # help
set -eu

# Defaults
OUT_DIR=${OUT_DIR:-"./backups"}
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
AUTH_CONTAINER=${AUTH_CONTAINER:-"auth_db"}
COURSES_CONTAINER=${COURSES_CONTAINER:-"courses_db"}
AUTH_DB=${AUTH_DB:-"auth_db"}
COURSES_DB=${COURSES_DB:-"courses_db"}
# Default password used by docker-compose in this project; override via -p or env POSTGRES_PASSWORD
PGPASSWORD_ENV=${POSTGRES_PASSWORD:-"1234"}
QUIET=0

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
  -o <dir>    Output directory (default: ${OUT_DIR})
  -p <pwd>    Postgres password to use when required (overrides POSTGRES_PASSWORD)
  -A <name>   Auth container name (default: ${AUTH_CONTAINER})
  -C <name>   Courses container name (default: ${COURSES_CONTAINER})
  -h          Show this help
  -q          Quiet
Notes:
  - Script prefers docker exec into containers. If docker or containers are not available,
    it will try local pg_dump connecting to ports 5432 (auth) and 5433 (courses).
EOF
}

# parse args
while [ $# -gt 0 ]; do
  case "$1" in
    -o)
      shift
      OUT_DIR=$1
      shift
      ;;
    -p)
      shift
      PGPASSWORD_ENV=$1
      shift
      ;;
    -A)
      shift
      AUTH_CONTAINER=$1
      shift
      ;;
    -C)
      shift
      COURSES_CONTAINER=$1
      shift
      ;;
    -q)
      QUIET=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      err "Unknown argument: $1"
      usage
      exit 2
      ;;
  esac
done

mkdir -p "$OUT_DIR"

# Helpers to check commands
has_cmd() {
  command -v "$1" >/dev/null 2>&1
}

# Try to create a backup by running pg_dump inside a running container and streaming to host
# args: <container> <dbname> <outfile>
backup_via_docker_exec() {
  container="$1"
  db="$2"
  outfile="$3"

  if ! has_cmd docker; then
    return 1
  fi

  # check container running
  if docker ps --format '{{.Names}}' | grep -xq "$container"; then
    log "Using docker exec -> container='$container' db='$db' -> $outfile"
    # Use inline sh to set PGPASSWORD and run pg_dump, stream output and gzip on host.
    # We wrap pg_dump in sh -c so we can set PGPASSWORD as env var for the command.
    # Note: docker exec may require the image to have sh available (common for official images).
    if docker exec -i "$container" sh -c "PGPASSWORD=${PGPASSWORD_ENV} pg_dump -U postgres -d ${db} -F p" | gzip >"$outfile"; then
      log "Backup complete: $outfile"
      return 0
    else
      err "docker exec pg_dump failed for container '$container' db '$db'"
      return 2
    fi
  else
    return 1
  fi
}

# Try using docker compose exec (some environments require this)
backup_via_docker_compose_exec() {
  svc="$1"
  db="$2"
  outfile="$3"

  if has_cmd docker && docker compose version >/dev/null 2>&1; then
    log "Trying 'docker compose exec' for service '$svc'..."
    if docker compose exec -T "$svc" sh -c "PGPASSWORD=${PGPASSWORD_ENV} pg_dump -U postgres -d ${db} -F p" | gzip >"$outfile"; then
      log "Backup complete: $outfile"
      return 0
    else
      err "docker compose exec pg_dump failed for service '$svc' db '$db'"
      return 2
    fi
  fi
  return 1
}

# Try local pg_dump (connects to host:port)
backup_via_local_pgdump() {
  host="$1"
  port="$2"
  db="$3"
  outfile="$4"

  if ! has_cmd pg_dump; then
    return 1
  fi

  log "Using local pg_dump -> ${host}:${port} db='${db}' -> ${outfile}"
  # Use PGPASSWORD env for authentication
  if PGPASSWORD="${PGPASSWORD_ENV}" pg_dump -h "${host}" -p "${port}" -U postgres -d "${db}" -F p | gzip >"$outfile"; then
    log "Backup complete: $outfile"
    return 0
  else
    err "local pg_dump failed for ${host}:${port} db='${db}'"
    return 2
  fi
}

# Generic wrapper to attempt a sequence of methods
do_backup() {
  container="$1"
  db="$2"
  local_port="$3"
  outfile="$4"

  # 1) docker exec into named container
  if backup_via_docker_exec "$container" "$db" "$outfile"; then
    return 0
  fi

  # 2) docker compose exec (service name)
  if backup_via_docker_compose_exec "$container" "$db" "$outfile"; then
    return 0
  fi

  # 3) local pg_dump connecting to host:local_port
  if backup_via_local_pgdump "localhost" "$local_port" "$db" "$outfile"; then
    return 0
  fi

  # none worked
  return 1
}

# Backup auth DB
auth_out="${OUT_DIR}/postgres-${AUTH_DB}-${TIMESTAMP}.sql.gz"
log "Starting backup for auth DB (${AUTH_DB})..."
if do_backup "$AUTH_CONTAINER" "$AUTH_DB" "5432" "$auth_out"; then
  log "Auth DB backup saved to: $auth_out"
else
  err "Failed to backup auth DB (${AUTH_DB})."
fi

# Backup courses DB
courses_out="${OUT_DIR}/postgres-${COURSES_DB}-${TIMESTAMP}.sql.gz"
log "Starting backup for courses DB (${COURSES_DB})..."
if do_backup "$COURSES_CONTAINER" "$COURSES_DB" "5433" "$courses_out"; then
  log "Courses DB backup saved to: $courses_out"
else
  err "Failed to backup courses DB (${COURSES_DB})."
fi

log "All done. Backups (if successful) are in: ${OUT_DIR}"
