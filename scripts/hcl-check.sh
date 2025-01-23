#!/usr/bin/env bash

fmt_errors=0
init_errors=0
newline_errors=0
hcl_number=0
debug_accumulated_time=0
dash_line="# ---------------------------------------------------------"
tmp_dir='tmp'

# Prints a full list of possible options of this script.
function get_help {
    echo "Syntax: $(basename "$0") [options]"
    echo "Where options are one or more of the following:"
    echo "  -h              Show this help and exit"
    echo "  -d              Debug mode: shows information on the commands being executed"
    echo "  -v              Gives more info"
    echo ""
    echo "NOTE: debug and verbose modes are mutually exclusive. Only the first one introduced as option will be active."
    echo ""
    echo "Examples"
    echo "$(basename "$0") -v"
    echo "$(basename "$0") -d"
    echo "$(basename "$0") -vd # Will run only verbose mode. Debug is ignored"
    echo "$(basename "$0") -dv # Will run only debug mode. Verbose is ignored"
    echo ""
    exit 0
}

# extract_hcl searches for .markdown files by using glob 'website/docs/{*/,?}*markdown' in:
# * website/docs/*.markdown
# * website/docs/*/*.markdown
# It will look for code blocks starting with '```hcl' and extract their contents until closing '```'
# and store in a file inside a tmp directory. Filename will be "base_filename+total_occurence_number"
# (e.g. edgegateway.html.markdown-100.tf)
function extract_hcl {
    curdir=$PWD

    [ -d "$tmp_dir" ] && rm -r "$tmp_dir"
    mkdir "$tmp_dir"
    cd "$tmp_dir" || exit 1

    awk 'function basename(file) {
        sub(".*/", "", file)
        return file
    } /^```hcl/ {
    flag = 1
    ++n
    s = ""
    next
    }
    /^```$/ {
        if (flag==1) {
            print_verbose "Extracting # ${n} HCL block in file ${basename(FILENAME)}"
            print s > (basename(FILENAME) "-" n ".tf"); close((basename(FILENAME) "-" n ".tf"))
        }
        flag = 0
    }
    flag {
    s = s $0 ORS
    }' ../website/docs/{*/,?}*markdown
    hcl_number="$(ls | wc -l | tr -d ' ')"
    cd "$curdir" || exit 1
}

# Runs the fmt subcommand of Terraform to check syntax errors on temporary HCL files
# that contain the documentation snippets.
function terraform_fmt_check {
    fmt_error_text=''

    for hcl_file in "$tmp_dir"/*.tf
    do
        command_start_time=$(date +%s)
        print_progress "Checking format of $hcl_file..."
        terraform fmt -check $hcl_file &>/dev/null
        retVal=$?
        if [ $retVal -ne 0 ]; then
            ((fmt_errors++))
            fmt_error_text="${fmt_error_text}$(terraform fmt -no-color -diff -check "$hcl_file" 2>&1)"
        fi
        command_end_time=$(date +%s)
        print_times "terraform fmt $hcl_file" "$command_start_time" "$command_end_time"
    done

    echo ""
}

# Runs against all generate HCL files and checks if there are no two (or more) empty lines
function multi_newline_check {
  newline_error_text=''

  for hcl_file in "$tmp_dir"/*.tf
      do
        print_progress "Checking for multiple newlines in $hcl_file..."

        hcl_content=$(cat $hcl_file)
        # 'cat -s' squeezes multiple adjacent empty lines, causing the output to be single spaced.
        hcl_content_without_newlines=$(cat -s $hcl_file)

        if [ "$hcl_content_without_newlines" != "$hcl_content" ]; then
          newline_error_text="${newline_error_text} \n ${hcl_file}"
          ((newline_errors++))
        fi
      done

      echo ""
}

# Performs a Terraform init on a temporary HCL file that contain the documentation snippets.
# If the init command fails, gives an error message and the script will fail.
function terraform_validation_check {
    init_error_text=''
    provider_version="$(git describe --abbrev=0 --tags | cut -d'v' -f 2)"
    curdir="$PWD"

    # Create a temporary folder where we put the file to test and the provider definition file.
    [ -d "$tmp_dir/validate" ] && rm -r "$tmp_dir/validate"
    mkdir "$tmp_dir/validate"

    echo "  (Using v$provider_version VCD provider version)"
    for hcl_file in "$tmp_dir"/*.tf
    do
        command_start_time=$(date +%s)
        print_progress "Validating $hcl_file..."
        # Copy the HCL file to temporary folder for terraform init to scan only this
        cp "$hcl_file" "$tmp_dir/validate/current.tf"
        cd "$tmp_dir/validate" || exit 1

        echo "
terraform {
  required_providers {
    vcd = {
      source  = \"vmware/vcd\"
      version = \"= $provider_version\"
    }
    nsxt = {
      source = \"vmware/nsxt\"
    }
  }
  required_version = \">= 0.13\"
}" > provider_setup.tf

        terraform_result="$(terraform init -no-color 2>&1 > /dev/null)"
        if [ -n "$terraform_result" ]; then
            ((init_errors++))
            init_error_text="${init_error_text}${terraform_result}"
        fi

        rm -f current.tf # We don't remove the provider so we don't download it everytime
        cd "$curdir" || exit 1
        command_end_time=$(date +%s)
        print_times "terraform init $hcl_file" "$command_start_time" "$command_end_time"
    done

    echo ""
}

# Prints a progress symbol or the argument if in verbose mode. Prints nothing if in debug mode.
function print_progress {
    if [[ -n "$VERBOSE" ]]
    then
        echo "$@"
    elif [[ -z "$DEBUG" ]]
    then
        printf '.'
    fi
}

# Prints the time that it took to execute the given command
function print_times {
    exec_command=$1
    command_start_time=$2
    command_end_time=$3
    secs=$((command_end_time-command_start_time))
    debug_accumulated_time=$((debug_accumulated_time+secs))
    if [[ -n "$DEBUG" ]] && [[ -z "$VERBOSE" ]]
    then
        echo "$exec_command | Duration: ${secs}s | Total: ${debug_accumulated_time}s"
    fi
}

# Prints the results of the format checking
function print_summary {
    end_time=$(date +%s)
    end_timestamp=$(date)
    secs=$((end_time-start_time))
    minutes=$((secs/60))
    remainder_sec=$((secs-minutes*60))
    if [[ $minutes -lt 60 ]]
    then
        elapsed=$(printf "%dm:%02ds" ${minutes} ${remainder_sec})
    else
        hours=$((minutes/60))
        remainder_minutes=$((minutes-hours*60))
        elapsed=$(printf "%dh:%dm:%02ds" ${hours} ${remainder_minutes} ${remainder_sec})
    fi

    echo ""
    echo "$dash_line"
    echo "# Summary:"
    echo ""
    echo "# Started:                 $start_timestamp"
    echo "# Ended:                   $end_timestamp"
    echo "# Elapsed:                 $elapsed ($secs sec)"
    echo "# Analyzed snippets:       $hcl_number"
    echo "# terraform fmt errors:    $fmt_errors"
    echo "# terraform init errors:   $init_errors"
    echo "# multiple newline errors: $newline_errors"
    echo "$dash_line"
    echo "# FULL report:"
    echo "# Format errors:"
    if [ -z "$fmt_error_text" ]
    then
        echo "# NONE (All ok!)"
    else
        echo "$fmt_error_text"
    fi
    echo ""
    echo "# Init errors:"
    if [ -z "$init_error_text" ]
    then
        echo "# NONE (All ok!)"
    else
        echo "$init_error_text"
    fi
    echo ""
    echo "# Multiple newline errors:"
    if [ -z "$newline_error_text" ]
    then
        echo "# NONE (All ok!)"
    else
        echo -e "$newline_error_text"
    fi
    echo ""
    echo "$dash_line"
    echo ""
}

if [ ! -d "website" ] 
then
    echo "(!) ERROR: Expected to find 'website' directory. Please run the script from project root directory"
    exit 1
fi

opts=$(getopt "hdv" "$@")
if [ $? -ne 0 ]
then
    get_help
fi
eval set -- "$opts"

while true
do
    case "$1"
    in
        -h)
            get_help;;
        -d)
            if [ -z "$VERBOSE" ]
            then
                export DEBUG=1
            fi
            shift;;
        -v)
            if [ -z "$DEBUG" ]
            then
                export VERBOSE=1
            fi
            shift;;
        --)
            shift
            break;;
    esac
done

start_time=$(date +%s)
start_timestamp=$(date)

# Dump the HCL code in tmp folder
extract_hcl

# Do a double newline check
echo '# Looking for multiple newlines:'
multi_newline_check

# Iterate over all extracted blocks and perform `terraform fmt`
echo '# Checking HCL format:'
terraform_fmt_check

## Iterate over all extracted blocks and perform `terraform init`
echo '# Checking HCL correctness:'
terraform_validation_check

print_summary

# If at least one of checks failed - return non 0 exit code
if [[ $fmt_errors = 0 ]] && [[ $init_errors = 0 ]] && [[ $newline_errors = 0 ]]
then
    echo '# Finished SUCCESSFULLY!'
    exit 0;
fi
echo '# Finished with FAILURES!'
exit 1
