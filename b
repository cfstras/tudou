#!/bin/bash

FOLDERS=$(find . -type d | cut -c3-)
CMDS=$(ls cmd)

function build() {
	echo "# ./cmd/$1"
	go get -d -v -t ./cmd/$1
	go build -v -o bin/$1 ./cmd/$1
}

case "$1" in
	"")
	for i in $CMDS; do
		build $i
	done
	;;
	update)
	for i in $CMDS; do
		echo "# $i"
		go get -d -u -v -t ./cmd/$i
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
	;;
	help)
	cat <<EOF
ultimate go build cmd, v2
usage:
	./b 		get deps & build all packages in cmd/ to bin/
	./b update	update deps for all packages in cmd/
	./b run <x>	build & run cmd x
	./b clean	clean bin
	./b fix		run goimports, go fix and vet on cmd/
EOF
	;;
esac
