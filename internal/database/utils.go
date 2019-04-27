package database

import (
	"strconv"
	"strings"
)

func ConvertIntSliceTostring(intSlice []int) string {
	valuesText := []string{}
	for _, number := range intSlice {
		text := strconv.Itoa(number)
		valuesText = append(valuesText, text)
	}
	return strings.Join(valuesText, ",")
}
