package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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

	var list []data.Item
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
		b, err := ioutil.ReadFile(jsonSrc)
		if err != nil {
			fmt.Println(err)
			os.Exit(ErrorFile)
		}
		err = json.Unmarshal(b, &list)
		if err != nil {
			fmt.Println(err)
			os.Exit(ErrorJSON)
		}
	} else {
		nope("source")
	}
	fmt.Println("Got", len(list), "items")
	fmt.Println("Sorting...")
	data.ItemSlice(list).Sort()

	if filebaseName != "" {
		name := filebaseName + ".json"
		fmt.Println("Saving json to", name)
		b, err := json.MarshalIndent(&list, "", "  ")
		if err != nil {
			fmt.Println(err)
			os.Exit(ErrorJSON)
		}
		err = ioutil.WriteFile(name, b, 0664)
		if err != nil {
			fmt.Println(err)
			os.Exit(ErrorFile)
		}
	}

	if saveTsv {
		name := filebaseName + ".tsv"
		fmt.Println("saving tsv to", name)
		file, err := os.Create(name)
		if err != nil {
			fmt.Println(err)
			os.Exit(ErrorFile)
		}
		w := csv.NewWriter(file)
		w.Comma = '\t'
		for _, item := range list {
			w.Write([]string{item.Code, item.Title})
		}
		w.Flush()
		if w.Error() != nil {
			fmt.Println(w.Error())
			os.Exit(ErrorTSV)
		}
	}
}
func nope(o string) {
	fmt.Println("Please select only one", o)
	flag.Usage()
	os.Exit(ErrorArgs)
}
