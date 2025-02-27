#!/usr/bin/env bash
scripts_dir=$(dirname $0)
cd $scripts_dir
scripts_dir=$PWD
cd -

if [ ! -d ./vcfa ]
then
    echo "source directory ./vcfa not found"
    exit 1
fi

wanted=$1

timeout=0
if [ -n "$VCFA_TIMEOUT" ]
then
    timeout=$VCFA_TIMEOUT
fi

if [ -n "$DRY_RUN" ]
then
    VERBOSE=1
fi

function check_exit_code {
    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        echo "execution error"
        exit 1
    fi
}

accepted_commands=(static token short acceptance sequential-acceptance binary
    binary-prepare short-provider acceptance-orguser short-provider-orguser binary-validate)

accepted="[${accepted_commands[*]}]"

if [ -z "$wanted" ]
then
    echo "Syntax: test TYPE"
    echo "    where TYPE is one of $accepted"
    exit 1
fi

# Adding some aliases to the accepted methods
if [ "$wanted" == "multi" ]
then
    wanted=multiple
fi
if [ "$wanted" == "acc" ]
then
    wanted=acceptance
fi

# Run test
echo "==> Run test $wanted"

cd vcfa

source_dir=$PWD

function check_for_config_file {
    config_file=vcfa_test_config.json
    if [ -n "${VCFA_CONFIG}" ]
    then
        echo "Using configuration file from variable \$VCFA_CONFIG"
        config_file=$VCFA_CONFIG
    fi
    if [ ! -f $config_file ]
    then
        echo "ERROR: test configuration file $config_file is missing"; \
        exit 1
    fi

}

function unit_test {
    if [ -n "$VERBOSE" ]
    then
        echo "go test -tags unit ${TEST} || exit 1"
        echo "go test -tags unit -v -timeout 5m"
    fi
    if [ -z "$DRY_RUN" ]
    then
        go test -tags unit ${TEST} || exit 1
        go test -tags unit -v -timeout 5m
    fi
}

function short_test {
    # If we are creating binary test files, we remove the old ones,
    # to avoid leftovers from previous runs to affect the current test
    if [ -n "$VCFA_ADD_PROVIDER" -a -n "$MORE_TAGS" -a -d ./test-artifacts ]
    then
        rm -f ./test-artifacts/*.tf
    fi
    if [ -n "$VERBOSE" ]
    then
        echo "VCFA_SHORT_TEST=1 go test -race -tags '$MORE_TAGS' -v -timeout 5m"
    fi
    if [ -z "$DRY_RUN" ]
    then
        VCFA_SHORT_TEST=1 go test -race -tags "$MORE_TAGS" -v -timeout 5m
        check_exit_code
    fi
    if [ -n "$VCFA_TEST_ORG_USER" ]
    then
        rm -f test-artifacts/cust.*.tf
    fi
}

function acceptance_test {
    tags="$1"
    testoptions="$2"
    if [ -z "$tags" ]
    then
        tags=functional
    fi
    if [ -n "$VERBOSE" ]
    then
        echo "# check for config file"
        echo "TF_ACC=1 go test -tags '$tags' -vcfa-add-provider -v -timeout $timeout"
    fi

    if [ -z "$DRY_RUN" ]
    then
        check_for_config_file
        TF_ACC=1 go test -tags "$tags" $testoptions -vcfa-add-provider -v -timeout $timeout
        check_exit_code
    fi
}

function binary_test {
    cd $source_dir
    if [ ! -d test-artifacts ]
    then
        echo "test-artifacts not found"
        exit 1
    fi
    cp $scripts_dir/test-binary.sh test-artifacts/test-binary.sh
    chmod +x test-artifacts/test-binary.sh
    cd test-artifacts
    for old_file in already_run.txt failed_tests.txt
    do
        if [ -f ${old_file} ]
        then
            rm -f ${old_file}
        fi
    done
    if [ -n "$NORUN" ]
    then
        pwd
        ls -l
        exit
    fi
    if [ -n "$VCFA_ENV_INIT" ]
    then
        ./test-binary.sh test-env-init
        exit $?
    fi

    if [ -n "$VCFA_ENV_APPLY" ]
    then
       ./test-binary.sh test-env-apply
        exit $?
    fi
    if [ -n "$VCFA_ENV_DESTROY" ]
    then
        ./test-binary.sh test-env-destroy
        exit $?
    fi
    if [ -n "$VCFA_ENV_VALIDATE" ]
    then
        ./test-binary.sh validate
        exit $?
    fi
    timestamp=$(date +%Y-%m-%d-%H-%M)
    ./test-binary.sh 2>&1 | tee test-binary-${timestamp}.txt
}

function exists_in_path {
    what=$1
    for dir in $(echo $PATH | tr ':' ' ')
    do
        wanted=$dir/$what
        if [ -x $wanted ]
        then
            echo $wanted
            return
        fi
    done
}

function make_token {
  for required in jq curl base64
  do
    found=$(exists_in_path $required)
    if [ -z "$found" ]
    then
      echo "Program $required not found - Can't retrieve token"
      exit 1
    fi
  done
  check_for_config_file
  echo "# Using credentials from $config_file"
  user=$(jq -r '.provider.user' $config_file)
  password=$(jq -r '.provider.password' $config_file)
  url=$(jq -r '.provider.url' $config_file)
  sysorg=$(jq -r '.provider.sysOrg' $config_file)
  org=$(jq -r '.provider.org' $config_file)

  if [ -z "$sysorg" -o "$sysorg" == "null" ]
  then
    if [ -z "$org" -o "$org" == "null" ]
    then
      echo "missing sysorg (and org) from configuration file. Can't retrieve token"
      exit 1
    fi
    sysorg=$org
  fi

  if [ -z "$user" -o "$user" == "null" ]
  then
    echo "missing user from configuration file. Can't retrieve token"
    exit 1
  fi
  if [ -z "$password" -o "$password" == "null" ]
  then
    echo "missing password from configuration file. Can't retrieve token"
    exit 1
  fi
  if [ -z "$url" -o "$url" == "null" ]
  then
    echo "missing url from configuration file. Can't retrieve token"
    exit 1
  fi

  options=""
  os=$(uname -s)
  is_linux=$(echo "$os" | grep -i linux)
  if [ -n "$is_linux" ]
  then
    options="-w 0"
  fi
  auth=$(echo -n "$user@$sysorg:$password" |base64 $options)

  echo "# Connecting to $url ($sysorg)"
  curl --silent --head --insecure \
    --header "Accept: application/*;version=32.0" \
    --header "Authorization: Basic $auth" \
    --request POST $url/sessions | grep -i authorization
}

function check_static {
    static_check=$(exists_in_path staticcheck)
    if [  -z "$staticcheck" -a -n "$GITHUB_ACTIONS" ]
    then
        # Variables found in staticcheck-config.sh
        # STATICCHECK_URL
        # STATICCHECK_VERSION
        # STATICCHECK_FILE
        if [ -f $scripts_dir/staticcheck-config.sh ]
        then
            source $scripts_dir/staticcheck-config.sh
        else
            echo "File $scripts_dir/staticcheck-config.sh not found - Skipping check"
            exit 0
        fi
        download_name=$STATICCHECK_URL/$STATICCHECK_VERSION/$STATICCHECK_FILE
        wget=$(exists_in_path wget)
        if [ -z "$wget" ]
        then
            echo "'wget' executable not found - Skipping check"
            exit 0
        fi
        $wget $download_name
        if [ -n "$STATICCHECK_FILE" ]
        then
            tar -xzf $STATICCHECK_FILE
            executable=$PWD/staticcheck/staticcheck
            if [ ! -f $executable ]
            then
                echo "Extracted executable not available - Skipping check"
            fi
            chmod +x $executable
            static_check=$executable
        fi
    fi

    if [ -n "$static_check" ]
    then
        echo "## Found $static_check"
        echo -n "## "
        $static_check -version
        echo -n "## Checking "
        pwd
        $static_check -tags ALL .
        exit_code=$?
        if [ "$exit_code" != "0" ]
        then
            exit $exit_code
        fi
    else
        echo "*** staticcheck executable not found - Check skipped"
    fi
}

case $wanted in
    static)
        check_static
        ;;
    token)
        make_token
        ;;
    binary-prepare)
        export NORUN=1
        binary_test
        ;;
     binary)
        binary_test
        ;;
     binary-validate)
        export VCFA_ENV_VALIDATE=1
        binary_test
        ;;
    unit)
        unit_test
        ;;
    short)
        export VCFA_SKIP_TEMPLATE_WRITING=1
        export MORE_TAGS="functional"
        short_test
        ;;
    short-provider-orguser)
        unset VCFA_SKIP_TEMPLATE_WRITING
        export VCFA_TEST_ORG_USER=1
        export VCFA_ADD_PROVIDER=1
        export MORE_TAGS="functional binary"
        short_test
        ;;
     short-provider)
        unset VCFA_SKIP_TEMPLATE_WRITING
        export VCFA_ADD_PROVIDER=1
        export MORE_TAGS="functional binary"
        short_test
        ;;
     short-provider-tm)
        unset VCFA_SKIP_TEMPLATE_WRITING
        export VCFA_ADD_PROVIDER=1
        export MORE_TAGS="tm binary"
        short_test
        ;;
    acceptance-orguser)
        export VCFA_TEST_ORG_USER=1
        acceptance_test functional
        ;;
    acceptance-cci)
        acceptance_test "unit cci" "-vcfa-test-org-user -vcfa-skip-leftovers-removal"
        ;;
    acceptance)
        acceptance_test "unit functional"
        ;;
    acceptance-coverage)
        acceptance_test "unit functional" "-coverprofile testacc.cover"
        ;;
    sequential-acceptance)
        acceptance_test functional "-race --parallel=1"
        ;;
    org)
        acceptance_test org
        ;;
    tm-acceptance)
        acceptance_test tm
        ;;
    tm-coverage)
        acceptance_test tm "-coverprofile=tm.cover"
        ;;
    *)
        echo "Unhandled testing method $wanted"
        echo "Accepted methods: $accepted"
        exit 1
esac

check_exit_code

