package memory

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type ResultSet struct {
	Addrs []int64 `json:"addrs"`
}

func SaveResultsJSON(addrs []int64, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(ResultSet{Addrs: addrs})
}

func LoadResultsJSON(r io.Reader) ([]int64, error) {
	var rs ResultSet
	dec := json.NewDecoder(r)
	if err := dec.Decode(&rs); err != nil {
		return nil, err
	}
	return rs.Addrs, nil
}

func SaveResultsCSV(addrs []int64, w io.Writer) error {
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"address"}); err != nil {
		return err
	}
	for _, addr := range addrs {
		if err := writer.Write([]string{fmt.Sprintf("%#x", addr)}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func LoadResultsCSV(r io.Reader) ([]int64, error) {
	reader := csv.NewReader(r)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	var addrs []int64
	for i, row := range rows {
		if i == 0 && len(row) > 0 && row[0] == "address" {
			continue
		}
		if len(row) == 0 {
			continue
		}
		val := row[0]
		addr, err := strconv.ParseInt(strings.TrimPrefix(val, "0x"), 16, 64)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

func SaveResultsJSONToFile(addrs []int64, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return SaveResultsJSON(addrs, file)
}

func LoadResultsJSONFromFile(path string) ([]int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadResultsJSON(file)
}

func SaveResultsCSVToFile(addrs []int64, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return SaveResultsCSV(addrs, file)
}

func LoadResultsCSVFromFile(path string) ([]int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadResultsCSV(file)
}
