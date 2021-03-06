#!/bin/bash

if ! which envsubst
then
    echo -e "Please install envsubst. OSX -> brew install gettext ; brew link --force gettext"
    exit 1
fi

# client options
CLIENT_CONFIG=()
CLIENT_CONFIG+=("--log debug")
CLIENT_CONFIG+=("--endpoints http://127.0.0.1:8800")
config=$(echo ${CLIENT_CONFIG[@]} | i=0 envsubst )

go run $GOPATH/src/github.com/bbva/qed/main.go client membership $config --event $1
