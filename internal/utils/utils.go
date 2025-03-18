package utils

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/go-errors/errors"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"golang.org/x/text/currency"
)

func ParseLatLngFromHexStr(hexStr string) (float64, float64, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, 0, err
	}

	var ewkbPoint ewkb.Point
	if err := ewkbPoint.Scan(bytes); err != nil {
		return 0, 0, err
	}

	return ewkbPoint.Y(), ewkbPoint.X(), nil
}

func IsDateBetween(targetDate string, startDate, endDate time.Time, layout string) (bool, error) {
	tDate, err := time.Parse(layout, targetDate)
	if err != nil {
		return false, errors.Errorf("invalid target date format: %v", err)
	}

	return (tDate.Equal(startDate) || tDate.After(startDate)) && (tDate.Equal(endDate) || tDate.Before(endDate)), nil
}

// target time format HH:mm:ss
func IsTimeBetween(targetTime string, tTimeZone *time.Location, startTime, endTime time.Time) (bool, error) {
	// to get correct timezone offset if parsing time only, we must include the year
	// issue: https://github.com/golang/go/issues/34101#issuecomment-528260666
	tTimeFormat := "2006 15:04:00"
	_tTime, err := time.ParseInLocation(tTimeFormat, fmt.Sprintf("%s %s", "2025", targetTime), tTimeZone)
	if err != nil {
		return false, errors.Errorf("invalid target time format: %v", err)
	}

	localTz, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		return false, errors.New(err)
	}

	_sTime, err := time.ParseInLocation(time.TimeOnly, startTime.Format(time.TimeOnly), localTz)
	if err != nil {
		return false, errors.New(err)
	}

	_eTime, err := time.ParseInLocation(time.TimeOnly, endTime.Format(time.TimeOnly), localTz)
	if err != nil {
		return false, errors.New(err)
	}

	tTime := time.Date(2025, 0, 0, _tTime.Hour(), _tTime.Minute(), _tTime.Second(), 0, tTimeZone)
	sTime := time.Date(2025, 0, 0, _sTime.Hour(), _sTime.Minute(), _sTime.Second(), 0, localTz)
	eTime := time.Date(2025, 0, 0, _eTime.Hour(), _eTime.Minute(), _eTime.Second(), 0, localTz)

	return (tTime.Equal(sTime) || tTime.After(sTime)) && (tTime.Equal(eTime) || tTime.Before(eTime)), nil
}

func FormatRupiah(a currency.Amount) string {
	formatted := fmt.Sprintf("%s", a)
	formatted = strings.ReplaceAll(formatted, ",", ".")
	formatted = strings.Replace(formatted, "IDR", "Rp", 1)

	return formatted
}

func GenerateDaysInMonth(year int, month time.Month) []time.Time {
	days := []time.Time{}

	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Hour)

	for d := firstDay; d.Before(lastDay) || d.Equal(lastDay); d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}

	return days
}

var errMustBeSlice = errors.New("must be slice")
var errMustBeStruct = errors.New("must be struct")
var errEmptySlice = errors.New("empty slice")

// rows must be a slice of struct with all fields are string
func WriteCSV(rows any, file *os.File) error {
	sliceType := reflect.TypeOf(rows)
	if sliceType.Kind() != reflect.Slice {
		return errors.New(errMustBeSlice)
	}

	if sliceType.Elem().Kind() != reflect.Struct {
		return errors.New(errMustBeStruct)
	}

	sliceValue := reflect.ValueOf(rows)
	if sliceValue.Len() == 0 {
		return errors.New(errEmptySlice)
	}

	firstSlice := sliceValue.Index(0)
	rowValue := reflect.ValueOf(firstSlice.Interface())
	rowType := rowValue.Type()

	for i := range rowValue.NumField() {
		fieldType := rowType.Field(i)
		if fieldType.Type.Kind() != reflect.String {
			return errors.Errorf("invalid type on field %s, must be string", fieldType.Name)
		}
	}

	// append header
	header := []string{}
	for i := range rowValue.NumField() {
		header = append(header, rowType.Field(i).Tag.Get("csv"))
	}

	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	if err := csvWriter.Write(header); err != nil {
		return errors.New(err)
	}

	for i := range sliceValue.Len() {
		row := sliceValue.Index(i)
		rowInterface := row.Interface()
		rowValue := reflect.ValueOf(rowInterface)
		csvRow := []string{}

		for j := range rowValue.NumField() {
			csvRow = append(csvRow, rowValue.Field(j).String())
		}

		if err := csvWriter.Write(csvRow); err != nil {
			return errors.New(err)
		}
	}

	return nil
}
