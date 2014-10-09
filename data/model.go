package data

import (
	"sort"
	"time"
)

type StatusMsg string

var (
	Success StatusMsg = "success"
)

type Response struct {
	Code int          `json:"code"`
	Msg  StatusMsg    `json:"msg"`
	Data ResponseData `json:"data"`
}

type ResponseData struct {
	PageNumber            int    `json:"pageNumber"`
	PageSize              int    `json:"pageSize"`
	TotalNumberOfElements int    `json:"totalNumberOfElements"`
	Data                  []Item `json:"data"`
}

type Item struct {
	Code      string        `json:"code"`
	ItemID    int64         `json:"itemID"`
	PicUrl    string        `json:"picurl"`
	PlayNum   int64         `json:"playNum"`
	PubDate   time.Time     `json:"pubDate"`
	Title     string        `json:"title"`
	TotalTime time.Duration `json:"totalTime"`
}

type ItemSlice []Item

func (s ItemSlice) Len() int           { return len(s) }
func (s ItemSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ItemSlice) Less(i, j int) bool { return s[i].Title < s[j].Title }

func (s ItemSlice) Sort() {
	sort.Sort(s)
}
