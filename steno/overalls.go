package main

import (
	"encoding/csv"
	"fmt"
	"io"
)

func exportOverallsCSV(arts ArtList, out io.Writer) error {

	pubs := arts.Pubs()
	pubs = append(pubs, "TOTAL")
	days := arts.Days()
	days = append(days, "TOTAL")

	data := map[string]map[string]int{}

	for _, pub := range pubs {
		data[pub] = map[string]int{}
	}

	for _, art := range arts {
		data[art.Pub][art.Day()]++
		data["TOTAL"][art.Day()]++
		data[art.Pub]["TOTAL"]++
		data["TOTAL"]["TOTAL"]++
	}

	w := csv.NewWriter(out)

	// header
	headerRow := make([]string, 1+len(days))
	headerRow[0] = ""
	for i, day := range days {
		if day == "" {
			day = "<missing>"
		}
		headerRow[1+i] = day
	}

	//fmt.Println(headerRow)
	err := w.Write(headerRow)
	if err != nil {
		return err
	}

	// rows
	for _, pub := range pubs {
		row := make([]string, 1+len(days))
		row[0] = pub
		for i, day := range days {
			v := data[pub][day]
			row[1+i] = fmt.Sprintf("%d", v)
		}
		err := w.Write(row)
		//fmt.Println(row)
		if err != nil {
			return err
		}
	}
	w.Flush()

	return nil
}
