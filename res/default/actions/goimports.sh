#!/bin/bash
set -e #fail early

# Run goimports on a go source file,
# falls back to gofmt if goimports not available.

view=$GOED_VIEW
inst=$GOED_INSTANCE

# get file location
src=`goed --api view_src_loc $inst $view`
cmd=goimports
which goimports 2> /dev/null || cmd=gofmt
# run the command on the file
$cmd -w $src
# reload view buffer from file
goed --api view_reload $inst $view 
