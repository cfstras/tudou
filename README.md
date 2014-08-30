# Tools for Tudou Scraping

## install

    sudo apt-get install golang

To install go (at least on Debian-ish systems)

    go get bitbucket.org/cfstras/tudou/cmd/...

Will install all commands.

## dev setup

    go get -d bitbucket.org/cfstras/tudou/
    cd $(go env GOPATH)/src/bitbucket.org/cfstras/tudou/
    ./b

Will download your repo and build all the packages in `cmd/` to `bin/`.


## commands

### tudou-load

Loads a video list from a source and saves it as json and optionally tsv.

Example:

    tudou-load -id 12345678 -tsv

Loads the video IDs of userID 12345678 and save as `12345678.json` and `12345678.tsv`.


## libs

### data

Loader for Tudou Lists and users
