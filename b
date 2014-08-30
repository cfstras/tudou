#!/bin/bash

BASEPATH=bitbucket.org/cfstras/tudou

FOLDERS=$(find * -type d)
CMDS=$(ls cmd)

function build() {
	echo "# $BASEPATH/cmd/$1"
	go get -d $BASEPATH/cmd/$1
	go build -v -o bin/$1 $BASEPATH/cmd/$1
}

case "$1" in
	"")
	for i in $CMDS; do
		echo "# $i"
		build $i
	done
	;;
	run)
	if [[ "$2" == "" ]]; then
		echo "possibilities:"
		for i in $CMDS; do
			echo "    $i"
		done
		echo "which one?"
	else
		build $2 && bin/$2
	fi
	;;
	clean)
		rm -rf bin
	;;
	fix)
		[[ -x $GOPATH/bin/goimports ]] || go get \
			code.google.com/p/go.tools/cmd/goimports
		$GOPATH/bin/goimports -l -w $FOLDERS
		for f in $(find . -type f -name "*.go"); do \
			go fix "$f"; \
			go tool vet -composites=false "$f"; \
		done
esac
