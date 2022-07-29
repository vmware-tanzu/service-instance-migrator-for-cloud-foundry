#!/usr/bin/env bash

# Usage: ./hack/credhub-demo.sh

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

# shellcheck source=/dev/null
if [[ -f "$ROOT_DIR/demo-helpers.sh" ]]; then
  source "$ROOT_DIR/demo-helpers.sh" || die "Could not find $ROOT_DIR/demo-helpers.sh"
fi

BIN_DIR="$ROOT_DIR/../bin"

function usage() {
    echo "Usage:"
    echo "  $0 [command] [flags]"
    printf "\n"
    echo "Available Commands:"
    printf "  %s\t\t%s\n" "setup" "Create credhub service and bind app to it"
    printf "  %s\t\t%s\n" "run" "Run service-instance-migrator to migrate the service"
    printf "  %s\t%s\n" "push-source" "Push demo app to source foundation"
    printf "  %s\t%s\n" "push-target" "Push demo app to target foundation"
    printf "  %s\t%s\n" "cleanup" "Delete credhub and apps from both foundations"
    printf "  %s\t\t%s\n" "help" "Prints usage"
    printf "\n"
    echo "Flags:"
    printf "  --source-org string\tThe TAS org to migrate from [default: \$CF_SOURCE_ORG]\n"
    printf "  --source-space string\tThe TAS space to migrate from [default: \$CF_SOURCE_SPACE]\n"
    printf "  --target-org string\tThe TAS org to migrate to [default: \$CF_TARGET_ORG]\n"
    printf "  --target-space string\tThe TAS space to migrate to [default: \$CF_TARGET_SPACE]\n"
    printf "  %s, --help\t\tPrints usage\n" "-h"
    printf "\n"
    echo "Environment Variables:"
    echo -e "  CF_SOURCE_APPS_DOMAIN\t\tSets the apps domain for the source foundation cloud controller api"
    echo -e "  CF_SOURCE_SYS_DOMAIN\t\tSets the system domain for the source foundation cloud controller api"
    echo -e "  CF_SOURCE_USERNAME\t\tSets the username for the source foundation cloud controller api"
    echo -e "  CF_SOURCE_PASSWORD\t\tSets the password for the source foundation cloud controller api"
    echo -e "  CF_SOURCE_ORG\t\t\tSets the org for the source foundation cloud controller api"
    echo -e "  CF_SOURCE_SPACE\t\tSets the space for the source foundation cloud controller api"
    echo -e "  CF_TARGET_APPS_DOMAIN\t\tSets the apps domain for the target foundation cloud controller api"
    echo -e "  CF_TARGET_SYS_DOMAIN\t\tSets the system domain for the target foundation cloud controller api"
    echo -e "  CF_TARGET_USERNAME\t\tSets the username for the target foundation cloud controller api"
    echo -e "  CF_TARGET_PASSWORD\t\tSets the password for the target foundation cloud controller api"
    echo -e "  CF_TARGET_ORG\t\t\tSets the org for the target foundation cloud controller api"
    echo -e "  CF_TARGET_SPACE\t\tSets the space for the target foundation cloud controller api"
    printf "\n"
    echo "Examples:"
    printf "  %s --source-org=tas1 --source-space=si-migrator-test-space --target-org=tas2 --target-space=si-migrator-test-space" "$0"
    printf "\n"
}

function cleanup() {
  login_tas2
  delete_app "secure-credentials-demo"
  delete_service "secure-credential"
  login_tas1
  delete_app "secure-credentials-demo"
  delete_service "secure-credential"

  echo "Cleanup complete!"
}

function setup() {
  login_tas1
  delete_and_create_service "secure-credentials-demo" "credhub" "default" "secure-credential" "{\"serviceinfo\": {\"url\":\"https://svc.example.com\",\"username\":\"user\", \"password\":\"pwd\"}}"
  clone_build_push "https://github.com/malston/secure-credentials-demo.git" "credhub-migration" "secure-credentials-demo" "$CF_SOURCE_SYS_DOMAIN" "$CF_SOURCE_ORG" "$CF_SOURCE_SPACE"

  printf "Run: %s\n" "curl -sk https://secure-credentials-demo-$CF_SOURCE_ORG-$CF_SOURCE_SPACE.$CF_SOURCE_APPS_DOMAIN/actuator/env | jq -r '.propertySources[] | select(.name == \"vcap\")'"
  read -rp "Press return when finished." -n 1 -r

  echo "Setup complete!"
}

function install_bin() {
  if ! command -v "$BIN_DIR"/service-instance-migrator &> /dev/null; then
    2>&1 echo "service-instance-migrator could not be found."
    exit 1
fi
}

function run() {
  install_bin
  export_dir=$(mktemp -d)
  echo "Exporting credhub to: $export_dir"
  "$BIN_DIR"/service-instance-migrator export space "$CF_SOURCE_SPACE" -o "$CF_SOURCE_ORG" --export-dir="$export_dir" --services="credhub" || die "failed to run service-instance-migrator export space $CF_SOURCE_SPACE -o $CF_SOURCE_ORG --export-dir=$export_dir"

  login_tas2 || die "failed to run login to target foundation"

  if [[ $CF_TARGET_ORG != "$CF_SOURCE_ORG" ]]; then
    mv "$export_dir/$CF_SOURCE_ORG" "$export_dir/$CF_TARGET_ORG"
  fi

  "$BIN_DIR"/service-instance-migrator import space "$CF_TARGET_SPACE" -o "$CF_TARGET_ORG" --import-dir="$export_dir" --services="credhub" || die "failed to run service-instance-migrator import space $CF_TARGET_SPACE -o $CF_TARGET_ORG --import-dir=$export_dir"

  clone_build_push "https://github.com/malston/secure-credentials-demo.git" "credhub-migration" "secure-credentials-demo" "$CF_TARGET_SYS_DOMAIN" "$CF_TARGET_ORG" "$CF_TARGET_SPACE"

  printf "Run: %s\n" "curl -sk https://secure-credentials-demo-$CF_TARGET_ORG-$CF_TARGET_SPACE.$CF_TARGET_APPS_DOMAIN/actuator/env | jq -r '.propertySources[] | select(.name == \"vcap\")'"
  read -rp "Press return when finished." -n 1 -r

  echo "See $export_dir for migrated data"
  echo "Demo complete!"
}

cd "$ROOT_DIR/.." || exit 1

# shellcheck source=/dev/null
if [[ -f "$ROOT_DIR/../.envrc" ]]; then
  source "$ROOT_DIR/../.envrc"
fi

verify_environment_variables

if [[ ! $? ]]; then
  usage
  exit 1
fi

while [ "$1" != "" ]; do
    param=$(echo "$1" | awk -F= '{print $1}')
    value=$(echo "$1" | awk -F= '{print $2}')
    case $param in
      -h | --help)
        usage
        exit
        ;;
      --source-org)
        CF_SOURCE_ORG=$value
        CF_TARGET_ORG=$value
        ;;
      --source-space)
        CF_SOURCE_SPACE=$value
        CF_TARGET_SPACE=$value
        ;;
      --target-org)
        CF_TARGET_ORG=$value
        ;;
      --target-space)
        CF_TARGET_SPACE=$value
        ;;
      setup)
        COMMAND=setup
        ;;
      run)
        COMMAND=run
        ;;
      push-source)
        COMMAND=push-source
        ;;
      push-target)
        COMMAND=push-target
        ;;
      cleanup)
        COMMAND=cleanup
        ;;
      help)
        usage
        exit
        ;;
      *)
        echo ""
        echo "Invalid option: [$param]"
        echo ""
        usage
        exit 1
        ;;
    esac
    shift
done

if [[ -z "$CF_SOURCE_ORG" ]]; then
    echo ""
    echo "Must set the CF_SOURCE_ORG environment variable or specify '--source-org' flag"
    echo ""
    usage
    exit 1
fi

if [[ -z "$CF_SOURCE_SPACE" ]]; then
    echo ""
    echo "Must set the CF_SOURCE_SPACE environment variable or specify '--source-space' flag"
    echo ""
    usage
    exit 1
fi

if [[ -z "$CF_TARGET_ORG" ]]; then
    echo ""
    echo "Must set the CF_TARGET_ORG environment variable or specify '--target-org' flag"
    echo ""
    usage
    exit 1
fi

if [[ -z "$CF_TARGET_SPACE" ]]; then
    echo ""
    echo "Must set the CF_TARGET_SPACE environment variable or specify '--target-space' flag"
    echo ""
    usage
    exit 1
fi

case $COMMAND in
setup)
  setup || die "demo setup failed"
  ;;
run)
  run
  ;;
push-source)
  login_tas1 || die "failed to run login to source foundation"
  clone_build_push "https://github.com/malston/secure-credentials-demo.git" "credhub-migration" "secure-credentials-demo" "$CF_SOURCE_SYS_DOMAIN"
  ;;
push-target)
  login_tas2 || die "failed to run login to target foundation"
  clone_build_push "https://github.com/malston/secure-credentials-demo.git" "credhub-migration" "secure-credentials-demo" "$CF_TARGET_SYS_DOMAIN"
  ;;
cleanup)
  cleanup || die "demo cleanup failed"
  ;;
*)
  setup || die "demo setup failed"
  run
  ;;
esac
