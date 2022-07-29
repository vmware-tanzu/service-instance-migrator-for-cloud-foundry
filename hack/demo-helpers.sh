#!/usr/bin/env bash

function die() {
    2>&1 echo "$@"
    exit 1
}

function gen_uuid() {
  local uuid
  if hash uuidgen 2>/dev/null; then
    uuid="$(uuidgen)"
  else
    uuid=$(cat /proc/sys/kernel/random/uuid)
  fi
  echo "$uuid"
}

function verify_environment_variables() {
  [[ -z "${CF_SOURCE_APPS_DOMAIN}" ]] && die "Missing \$CF_SOURCE_APPS_DOMAIN -- The apps domain for the source foundation is required"
  [[ -z "${CF_SOURCE_SYS_DOMAIN}" ]] && die "Missing \$CF_SOURCE_SYS_DOMAIN -- The system domain for the source foundation is required"
  [[ -z "${CF_SOURCE_USERNAME}" ]] && die "Missing \$CF_SOURCE_USERNAME -- The username for the source foundation is required"
  [[ -z "${CF_SOURCE_PASSWORD}" ]] && die "Missing \$CF_SOURCE_PASSWORD -- The password for the source foundation is required"
  [[ -z "${CF_TARGET_APPS_DOMAIN}" ]] && die "Missing \$CF_TARGET_APPS_DOMAIN -- The apps domain for the target foundation is required"
  [[ -z "${CF_TARGET_SYS_DOMAIN}" ]] && die "Missing \$CF_TARGET_SYS_DOMAIN -- The system domain for the target foundation is required"
  [[ -z "${CF_TARGET_USERNAME}" ]] && die "Missing \$CF_TARGET_USERNAME -- The username for the target foundation is required"
  [[ -z "${CF_TARGET_PASSWORD}" ]] && die "Missing \$CF_TARGET_PASSWORD -- The password for the target foundation is required"
  return 0
}

function prompt() {
  local msg=${1:-"Are you sure?"}
  read -p "${msg} " -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    return 1
  fi
}

function mvnw() {
  if hash mvn 2>/dev/null; then
    mvn "$@"
  else
    ./mvnw "$@"
  fi
}

function gradlew() {
  if hash gradle 2>/dev/null; then
    gradle "$@"
  else
    ./gradlew "$@"
  fi
}

function login_tas1() {
  export CF_HOME=$HOME/.cf_tas1
  [[ -z "${CF_SOURCE_SYS_DOMAIN}" ]] && die "Source system domain is required"
  [[ -z "${CF_SOURCE_USERNAME}" ]] && die "Source username is required"
  [[ -z "${CF_SOURCE_PASSWORD}" ]] && die "Source password is required"
  [[ -z "${CF_SOURCE_ORG}" ]] && die "Source org is required"
  [[ -z "${CF_SOURCE_SPACE}" ]] && die "Source space is required"
  login "$CF_SOURCE_SYS_DOMAIN" "$CF_SOURCE_ORG" "$CF_SOURCE_SPACE" "$CF_SOURCE_USERNAME" "$CF_SOURCE_PASSWORD"
}

function login_tas2() {
  export CF_HOME=$HOME/.cf_tas2
  [[ -z "${CF_TARGET_SYS_DOMAIN}" ]] && die "Target system domain is required"
  [[ -z "${CF_TARGET_USERNAME}" ]] && die "Target username is required"
  [[ -z "${CF_TARGET_PASSWORD}" ]] && die "Target password is required"
  [[ -z "${CF_TARGET_ORG}" ]] && die "Target org is required"
  [[ -z "${CF_TARGET_SPACE}" ]] && die "Target space is required"
  login "$CF_TARGET_SYS_DOMAIN" "$CF_TARGET_ORG" "$CF_TARGET_SPACE" "$CF_TARGET_USERNAME" "$CF_TARGET_PASSWORD"
}

function login() {
  local cf_target=$1
  local cf_org=$2
  local cf_space=$3
  local cf_username=$4
  local cf_password=$5
  cf login -a "https://api.$cf_target" \
    --skip-ssl-validation \
    -o "$cf_org" \
    -s "$cf_space" \
    -u "$cf_username" \
    -p "$cf_password"
}

function auth_tas1() {
  export CF_HOME=$HOME/.cf_tas1
  [[ -z "${CF_SOURCE_SYS_DOMAIN}" ]] && die "Source system domain is required"
  [[ -z "${CF_SOURCE_USERNAME}" ]] && die "Source username is required"
  [[ -z "${CF_SOURCE_PASSWORD}" ]] && die "Source password is required"
  api_auth "$CF_SOURCE_SYS_DOMAIN" "$CF_SOURCE_USERNAME" "$CF_SOURCE_PASSWORD"
}

function auth_tas2() {
  export CF_HOME=$HOME/.cf_tas2
  [[ -z "${CF_TARGET_SYS_DOMAIN}" ]] && die "Target system domain is required"
  [[ -z "${CF_TARGET_USERNAME}" ]] && die "Target username is required"
  [[ -z "${CF_TARGET_PASSWORD}" ]] && die "Target password is required"
  api_auth "$CF_TARGET_SYS_DOMAIN" "$CF_TARGET_USERNAME" "$CF_TARGET_PASSWORD"
}

function api_auth() {
  local cf_target=$1
  local cf_username=$2
  local cf_password=$3
  cf api "https://api.$cf_target" --skip-ssl-validation >/dev/null
  cf auth "$cf_username" "$cf_password" >/dev/null
}

function target() {
  local cf_org=$1
  local cf_space=$2
  cf target -o "$cf_org" -s "$cf_space" >/dev/null
}

function check_app_exists() {
  local app="${1:?Variable 'app' is required}"
  if cf app "$app" --guid | grep -i 'fail' > /dev/null; then
    return 1
  fi
  return 0
}

function check_service_exists() {
  local service="${1:?Variable 'service' is required}"
  if cf service "$service" --guid | grep -i 'fail' > /dev/null; then
    return 1
  fi
  return 0
}

function check_service_key_exists() {
  local service_instance="${1:?Variable 'service_instance' is required}"
  local service_key="${2:?Variable 'service_key' is required}"
  if cf service-key "$service_instance" "$service_key" --guid | grep -i 'fail' > /dev/null; then
    return 1
  fi
  return 0
}

function check_progress_status() {
  local op="${1:?Variable 'op' is required}"
  local service="${2:?Variable 'service' is required}"
  status=$(cf service "$service" | grep 'succeeded' | awk '{print $2}')
  if [[ $status =~ $op ]]; then
    return 0
  fi
  return 1
}

function wait_for_ready() {
  local service="${1:?Variable 'service' is required}"
  local operation="${2:?Variable 'operation' is required}"
  while true; do
    case $operation in
        [d]* ) msg="delete"; break;;
        [c]* ) msg="create"; break;;
        * ) return 1;;
    esac
  done
  echo "Waiting for service to ${msg}."
  while true; do
    if [ "$operation" = "c" ]; then
      if check_progress_status "create|update" "$service"; then
        break
      fi
      echo "Still waiting for service to create."
    else
      if check_progress_status "delete" "$service"; then
        break
      else
        if ! check_service_exists "$service"; then
            break
        fi
      fi
      echo "Still waiting for service to delete."
    fi
    sleep 10
  done
  echo "Service is ${msg}d."
}

function verify_service_instance_migrator_config_environment_variables() {
  [[ -z "${CF_SOURCE_APPS_DOMAIN}" ]] && die "Missing \$CF_SOURCE_APPS_DOMAIN -- The apps domain for the source foundation is required"
  [[ -z "${CF_SOURCE_SYS_DOMAIN}" ]] && die "Missing \$CF_SOURCE_SYS_DOMAIN -- The system domain for the source foundation is required"
  [[ -z "${CF_SOURCE_USERNAME}" ]] && die "Missing \$CF_SOURCE_USERNAME -- The username for the source foundation is required"
  [[ -z "${CF_SOURCE_PASSWORD}" ]] && die "Missing \$CF_SOURCE_PASSWORD -- The password for the source foundation is required"
  [[ -z "${CF_SOURCE_CCDB_HOST}" ]] && die "Host for the source foundation cloud controller database is required"
  [[ -z "${CF_SOURCE_CCDB_USERNAME}" ]] && die "The username for the source foundation cloud controller database is required"
  [[ -z "${CF_SOURCE_CCDB_PASSWORD}" ]] && die "The password for the source foundation cloud controller database is required"
  [[ -z "${CF_SOURCE_CCDB_ENCRYPTION_KEY}" ]] && die "The encryption key used to store creds in the cloud controller database for the source foundation is required"
  [[ -z "${OPSMAN_SOURCE_HOSTNAME}" ]] && die "The username for the source foundation opsman is required"
  [[ -z "${OPSMAN_SOURCE_SSH_USERNAME}" ]] && die "The ssh user for the source foundation opsman is required"
  [[ -z "${OPSMAN_SOURCE_SSH_KEY}" ]] && die "The path to the ssh key for the source foundation opsman is required"
  [[ -z "${CF_TARGET_APPS_DOMAIN}" ]] && die "Missing \$CF_TARGET_APPS_DOMAIN -- The apps domain for the target foundation is required"
  [[ -z "${CF_TARGET_SYS_DOMAIN}" ]] && die "Missing \$CF_TARGET_SYS_DOMAIN -- The system domain for the target foundation is required"
  [[ -z "${CF_TARGET_USERNAME}" ]] && die "Missing \$CF_TARGET_USERNAME -- The username for the target foundation is required"
  [[ -z "${CF_TARGET_PASSWORD}" ]] && die "Missing \$CF_TARGET_PASSWORD -- The password for the target foundation is required"
  [[ -z "${CF_TARGET_CCDB_HOST}" ]] && die "The host for the target foundation cloud controller database is required"
  [[ -z "${CF_TARGET_CCDB_USERNAME}" ]] && die "The username for the target foundation cloud controller database is required"
  [[ -z "${CF_TARGET_CCDB_PASSWORD}" ]] && die "The password for the target foundation cloud controller database is required"
  [[ -z "${CF_TARGET_CCDB_ENCRYPTION_KEY}" ]] && die "The encryption key used to store creds in the cloud controller database for the target foundation is required"
  [[ -z "${OPSMAN_TARGET_HOSTNAME}" ]] && die "The username for the target foundation opsman is required"
  [[ -z "${OPSMAN_TARGET_SSH_USERNAME}" ]] && die "The ssh user for the target foundation opsman is required"
  [[ -z "${OPSMAN_TARGET_SSH_KEY}" ]] && die "The path to the ssh key for the target foundation opsman is required"
  return 0
}

function create_service_instance_migrator_config() {
  if [[ -f "$HOME/.config/si-migrator/config.yml" ]]; then
    cp "$HOME/.config/si-migrator/config.yml" "$HOME/.config/si-migrator/config.yml.backup"
  fi
  rm -f "$HOME/.config/si-migrator/config.yml"
  mkdir -p "$HOME/.config/si-migrator/"
  cat > "$HOME/.config/si-migrator/config.yml" <<EOF
---
foundations:
  source:
    url: https://$OPSMAN_SOURCE_HOSTNAME
    client_id: $OPSMAN_SOURCE_CLIENT_ID
    client_secret: $OPSMAN_SOURCE_CLIENT_SECRET
    username: $OPSMAN_SOURCE_USERNAME
    password: $OPSMAN_SOURCE_PASSWORD
    hostname: $OPSMAN_SOURCE_HOSTNAME
    private_key: $HOME/.ssh/pcfpre-clt
    ssh_user: "ubuntu"
  target:
    url: https://$OPSMAN_TARGET_HOSTNAME
    client_id: $OPSMAN_TARGET_CLIENT_ID
    client_secret: $OPSMAN_TARGET_CLIENT_SECRET
    username: $OPSMAN_TARGET_USERNAME
    password: $OPSMAN_TARGET_PASSWORD
    hostname: $OPSMAN_TARGET_HOSTNAME
    private_key: $HOME/.ssh/taspre-clt
    ssh_user: "ubuntu"
migrators:
  sqlserver:
    source_ccdb:
      db_host: $CF_SOURCE_CCDB_HOST
      db_username: $CF_SOURCE_CCDB_USERNAME
      db_password: $CF_SOURCE_CCDB_PASSWORD
      db_encryption_key: $CF_SOURCE_CCDB_ENCRYPTION_KEY
      ssh_host: $OPSMAN_SOURCE_HOSTNAME
      ssh_username: $OPSMAN_SOURCE_SSH_USERNAME
      ssh_private_key: $OPSMAN_SOURCE_SSH_KEY
      ssh_tunnel: true
    target_ccdb:
      db_host: $CF_TARGET_CCDB_HOST
      db_username: $CF_TARGET_CCDB_USERNAME
      db_password: $CF_TARGET_CCDB_PASSWORD
      db_encryption_key: $CF_TARGET_CCDB_ENCRYPTION_KEY
      ssh_host: $OPSMAN_TARGET_HOSTNAME
      ssh_username: $OPSMAN_TARGET_SSH_USERNAME
      ssh_private_key: $OPSMAN_TARGET_SSH_KEY
      ssh_tunnel: true
  ecs:
    source_ccdb:
      db_host: $CF_SOURCE_CCDB_HOST
      db_username: $CF_SOURCE_CCDB_USERNAME
      db_password: $CF_SOURCE_CCDB_PASSWORD
      db_encryption_key: $CF_SOURCE_CCDB_ENCRYPTION_KEY
      ssh_host: $OPSMAN_SOURCE_HOSTNAME
      ssh_username: $OPSMAN_SOURCE_SSH_USERNAME
      ssh_private_key: $OPSMAN_SOURCE_SSH_KEY
      ssh_tunnel: true
    target_ccdb:
      db_host: $CF_TARGET_CCDB_HOST
      db_username: $CF_TARGET_CCDB_USERNAME
      db_password: $CF_TARGET_CCDB_PASSWORD
      db_encryption_key: $CF_TARGET_CCDB_ENCRYPTION_KEY
      ssh_host: $OPSMAN_TARGET_HOSTNAME
      ssh_username: $OPSMAN_TARGET_SSH_USERNAME
      ssh_private_key: $OPSMAN_TARGET_SSH_KEY
      ssh_tunnel: true
  mysql:
    # backup_type: minio
    # backup_directory: $ROOT_DIR/mysql-backup
    backup_type: scp
    scp:
      username: mysql
      hostname: mysql-backup.plat-svcs.pez.vmware.com
      port: 22
      destination_directory: /var/vcap/data/mysql/backups
      private_key: $OPSMAN_SOURCE_SSH_KEY
    minio:
      alias: mysql-backups
      url: $CF_SOURCE_BACKUP_URL
      access_key: $CF_SOURCE_BACKUP_ACCESS_KEY
      secret_key: $CF_SOURCE_BACKUP_SECRET_KEY
      bucket_name: $CF_SOURCE_BACKUP_BUCKET
      bucket_path: $CF_SOURCE_BACKUP_PATH
      insecure: $CF_SOURCE_BACKUP_INSECURE
EOF
  echo "Created $HOME/.config/si-migrator/config.yml"
}

function cf_push() {
  local repo="${1:?Variable 'repo' is required}"
  local branch="${2:?Variable 'branch' is required}"
  local app="${3:?Variable 'app' is required}"
  local target="${4:?Variable 'target' is required}"
  local tempdir="${5:-"$(mktemp -d)"}"
  if [[ ! -d "$tempdir/${app}" ]]; then
    git clone -b "${branch}" "${repo}" "${tempdir}/${app}"
    cd "${tempdir}/${app}" || die "Failed to find ${app} in ${tempdir}"
    gradlew clean assemble
  fi
  cd "${tempdir}/${app}" || die "Failed to find ${app} in ${tempdir}"
  cf push "${app}" --random-route || die "Failed to push ${app} to api.${target}"
  echo "App is located here: ${tempdir}/${app}"
  cd - || exit 1
}

function clone_build_push() {
  local repo="${1:?Variable 'repo' is required}"
  local branch="${2:?Variable 'branch' is required}"
  local app="${3:?Variable 'app' is required}"
  local target="${4:?Variable 'target' is required}"
  local org="${5:?Variable 'org' is required}"
  local space="${6:?Variable 'space' is required}"
  tempdir=$(mktemp -d)
  git clone -b "${branch}" "${repo}" "${tempdir}/${app}"
  cd "${tempdir}/${app}" || die "Failed to clone ${app} to ${tempdir}"
  gradlew clean assemble
  cf target -o "$org" -s "$space"
  cf push "${app}" -n "${app}-$org-$space" -f "${tempdir}/${app}/manifest.yml" || die "Failed to push ${app} to api.${target}"
  echo "App is located here: ${tempdir}/${app}"
  cd - || exit 1
}

function delete_service() {
  local service_instance="${1:?Variable 'service_instance' is required}"
  if check_service_exists "$service_instance"; then
    cf delete-service "$service_instance" -f
    wait_for_ready "$service_instance" "d" || die "failed to delete $service_instance"
  fi
}

function delete_service_key() {
  local service_instance="${1:?Variable 'service_instance' is required}"
  local service_key="${2:?Variable 'service_key' is required}"
  if check_service_key_exists "$service_instance" "$service_key"; then
    cf delete-service-key "$service_instance" "$service_key" -f
  fi
}

function delete_app() {
  local app="${1:?Variable 'app' is required}"
  cf delete "$app" -f -r
}

function delete_and_create_service() {
  local app="${1:?Variable 'app' is required}"
  local service="${2:?Variable 'service' is required}"
  local plan="${3:?Variable 'plan' is required}"
  local service_instance="${4:?Variable 'service_instance' is required}"
  local config_params="${5}"
  local non_interactive="${6}"
  if check_app_exists "$app"; then
    if [ "$non_interactive" = true ] ; then
        delete_app "$app"
    else
      if prompt "Do you want to delete ${app}?"; then
        delete_app "$app"
      fi
    fi
  fi
  if check_service_exists "$service_instance"; then
    if [ "$non_interactive" = true ] ; then
      cf delete-service "$service_instance" -f
      wait_for_ready "$service_instance" "d" || die "failed to delete $service_instance"
    else
      if prompt "$service_instance already exists. do you want to delete it?"; then
        cf delete-service "$service_instance" -f
        wait_for_ready "$service_instance" "d" || die "failed to delete $service_instance"
      fi
    fi
  fi
  if ! check_service_exists "$service_instance"; then
    if [[ -n "$config_params" ]]; then
      cf create-service "$service" "$plan" "$service_instance" -c "$config_params" || die "failed to create $service_instance"
    else
      cf create-service "$service" "$plan" "$service_instance" || die "failed to create $service_instance"
    fi
  fi
  wait_for_ready "$service_instance" "c" || die "failed to create $service_instance"
}

function create_service_export_config() {
  local export_dir="${1:?Variable 'export_dir' is required}"
  local service="${2:?Variable 'service' is required}"
  local plan="${3:?Variable 'plan' is required}"
  local service_instance="${4:?Variable 'service_instance' is required}"
  tas1_instance_guid=$(cf service "$service_instance" --guid || die "failed to get guid from service instance")
  mkdir -p "${export_dir}/app-migration-org/si-migrator" || exit 1
  cat > "${export_dir}/app-migration-org/si-migrator/$service_instance.yml" <<EOF
name: $service
guid: $tas1_instance_guid
type: managed_service_instance
service: $service
plan: $plan
EOF
}
