package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"bitbucket.org/cfstras/tudou/data"
)

const (
	ErrorNone = iota
	ErrorArgs
	ErrorLoad
	ErrorFile
	ErrorJSON
	ErrorTSV
)

func main() {
	var listId int64
	var username, jsonSrc string

	var saveTsv, help bool

	flag.Int64Var(&listId, "id", 0, "Source: The user ID to use")
	flag.StringVar(&username, "user", "", "Source: The username to use")
	flag.StringVar(&jsonSrc, "json", "", "Source: The JSON list file to use")

	flag.BoolVar(&saveTsv, "tsv", false, `Save to json AND tsv; format:"id\tTitle\n"`)

	flag.BoolVar(&help, "help", false, "Print this help")
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(1)
	}

	var list data.ItemSlice
	var err error
	var filebaseName string
	if listId != 0 {
		filebaseName = fmt.Sprint(listId)
		list, err = data.LoadList(listId)
		if err != nil {
			fmt.Println(err)
			os.Exit(ErrorLoad)
		}
	} else if username != "" {
		filebaseName = username
		listId, err = data.GetUserId(username)
		if err != nil {
			fmt.Println(err)
			os.Exit(ErrorJSON)
		}
		fmt.Println("User Id:", listId)
		list, err = data.LoadList(listId)
		if err != nil {
			fmt.Println(err)
			os.Exit(ErrorLoad)
		}
	} else if jsonSrc != "" {
		filebaseName = strings.TrimSuffix(jsonSrc, ".json")
		err = list.LoadJSON(jsonSrc)
		if err != nil {
			fmt.Println(err)
			os.Exit(ErrorJSON)
		}
	} else {
		nope("source")
	}
	fmt.Println("Got", len(list), "items")
	fmt.Println("Sorting...")
	list.Sort()

	if filebaseName != "" {
		name := filebaseName + ".json"
		fmt.Println("Saving json to", name)
		err = list.WriteJSON(name)
		if err != nil {
			fmt.Println(err)
			os.Exit(ErrorJSON)
		}
	}

	if saveTsv {
		name := filebaseName + ".tsv"
		fmt.Println("saving tsv to", name)
		err = list.WriteTSV(name)
		if err != nil {
			fmt.Println(err)
			os.Exit(ErrorTSV)
		}
	}
}
func nope(o string) {
	fmt.Println("Please select only one", o)
	flag.Usage()
	os.Exit(ErrorArgs)
}
