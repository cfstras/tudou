# tools [![Build Status](https://travis-ci.org/cfstras/tudou.svg?branch=master)](https://travis-ci.org/cfstras/tudou)

This repository contains some small library and commandline tools to effectively
crawl a lot of videos from Tudou, a chinese YouTube clone.

The setup is as follows:

- Create a queue on Amazon Simple Queue Service
- Create an Amazon S3 Bucket for your videos
- Load metadata with `tudou-load`
- Fill metadata into SQS with `tudou-scrape -send`
- Start a bunch of EC2 machines running `tudou-scrape -receive`

There you go, have fun crawling massive amounts of videos on the cheap!

Changing this to support other sites should be easy, and if you do so, I would be
very happy about a pull request.

A quick note:
This is intended only to download and use videos in a legal fashion. If anything
you do with this toolset is forbidden in your country, I urge you **a)** not to
do it and **b)** to leave me out of it.

## install

Install [Go](http://golang.org):

    sudo apt-get install golang

Install all commands:

    go get github.com/cfstras/tudou/cmd/...

## dev setup

    go get -d github.com/cfstras/tudou/
    cd $(go env GOPATH)/src/github.com/cfstras/tudou/
    ./b

Will download your repo and build all the packages in `cmd/` to `bin/`.

## commands

### tudou-load

Loads a video list from a source and saves it as json and optionally tsv.

Example:

    tudou-load -id 12345678 -tsv

Loads the video IDs of userID 12345678 and save as `12345678.json` and `12345678.tsv`.

### tudou-scrape

Sends Video IDs to SQS.

    tudou-scrape -send -json bla.json -queue huxlipux -region us-east-1

Receives Video IDs from SQS, loads videos (if the file is not already on S3) and uploads to S3.

    tudou-scrape -receive -queue huxlipux -region us-east-1 -bucket huxlipux

### tudou-rename

Renames Videos in the form `<Tudou-ID> *.*` to `<Tudou-ID>.*`, for the older version of the scraper which saved to `<id> <title>.json`.

## libs

### data

Loading library for Tudou list and user metadata, and loading from and to JSON and TSV

### dl

Wrapper around youtube-dl

### color

Little wrappers around [daviddengcn/go-colortext](http://github.com/daviddengcn/go-colortext) for easy logging.

# license

Beerware. Patches welcome.
