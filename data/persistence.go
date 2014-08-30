package data

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func (s *ItemSlice) LoadJSON(file string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, s)
	if err != nil {
		return err
	}
	return nil
}

func (s ItemSlice) WriteJSON(file string) error {
	b, err := json.MarshalIndent(&s, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, b, 0664)
	if err != nil {
		return err
	}
	return nil
}

func (s ItemSlice) WriteTSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	w.Comma = '\t'
	for _, item := range s {
		w.Write([]string{item.Code, item.Title})
	}
	w.Flush()
	if w.Error() != nil {
		return w.Error()
	}
	return nil
}

func (s *ItemSlice) LoadTSV(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	r := csv.NewReader(file)
	r.Comma = '\t'
	for err != io.EOF {
		var rec []string
		rec, err = r.Read()
		if err != nil && err != io.EOF {
			return err
		}
		var i Item
		if len(rec) != 2 {
			return errors.New(fmt.Sprint("Record ", len(*s)+1,
				" has wrong length ", len(rec), ": ", rec))
		}
		i.Code = rec[0]
		i.Title = rec[1]
		*s = append(*s, i)
	}
	return nil
}
