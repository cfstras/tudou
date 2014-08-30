package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
)

const (
	ListBaseUrl = "http://www.tudou.com/home/item/list.do"
	UserBaseUrl = "http://www.tudou.com/home/"
)

var (
	uidRegex = regexp.MustCompile(`uid\s*:\s*'([0-9]+)'\s*,`)
)

func LoadList(listId int64) (list []Item, err error) {
	url := func(listId int64, page int) string {
		return fmt.Sprintf(ListBaseUrl+"?uid=%d&page=%d&pageSize=%d",
			listId, page, 100)
	}
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	got, page, remaining, try := 0, 1, 1, 1
	for remaining > 0 {
		url := url(listId, page)
		fmt.Println("GET", url)
		httpRes, err := http.Get(url)
		if !handle(err, &try) {
			continue
		}
		b, err := ioutil.ReadAll(httpRes.Body)
		if !handle(err, &try) {
			continue
		}
		var res Response
		err = json.Unmarshal(b, &res)
		if err != nil {
			err = errors.New(fmt.Sprint(httpRes.StatusCode, " ", httpRes.Status,
				": ", err.Error()))
		}
		if !handle(err, &try) {
			continue
		}
		if res.Code != 0 || res.Msg != Success {
			if !handle(errors.New(string(res.Msg)), &try) {
				continue
			}
		}
		got += res.Data.PageSize
		remaining = res.Data.TotalNumberOfElements - got
		list = append(list, res.Data.Data...)
		page++

		try = 1
		err = nil
	}
	return
}

func GetUserId(name string) (listId int64, err error) {
	url := UserBaseUrl + name + "/"

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	try := 1
	for {
		fmt.Println("GET", url)
		var httpRes *http.Response
		httpRes, err = http.Get(url)
		if !handle(err, &try) {
			continue
		}
		var b []byte
		b, err = ioutil.ReadAll(httpRes.Body)
		if !handle(err, &try) {
			continue
		}

		res := uidRegex.FindAllSubmatch(b, -1)
		if len(res) == 0 {
			err = errors.New("UserId Regex not found...")
			fmt.Println(string(b))
			return
		}
		if len(res) > 1 {
			for i, s := range res {
				fmt.Println("Match", i, ":", string(s[1]))
			}
			err = errors.New("UserId Regex found too often")
			return
		}
		listId, err = strconv.ParseInt(string(res[0][1]), 10, 64)
		if !handle(err, &try) {
			continue
		}
		err = nil
		return
	}
}

func handle(err error, tries *int) bool {
	if err == nil {
		return true
	}
	if *tries < 5 {
		fmt.Println(err, "- trying again...")
		*tries++
		return false
	} else {
		panic(err)
	}
}
