package utils

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/go-errors/errors"
	"github.com/google/uuid"
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
func IsTimeBetween(targetTime, startTime, endTime time.Time) (bool, error) {
	return (targetTime.Equal(startTime) || targetTime.After(startTime)) && (targetTime.Equal(endTime) || targetTime.Before(endTime)), nil
}

func ParseTimeString(timeStr string, tz *time.Location) (time.Time, error) {
	// to get correct timezone offset if parsing time only, we must include the year
	// issue: https://github.com/golang/go/issues/34101#issuecomment-528260666

	// validate format
	if len(timeStr) != 8 {
		return time.Time{}, errors.New("invalid time format")
	}

	if timeStr[2] != ':' || timeStr[5] != ':' {
		return time.Time{}, errors.New("invalid time format")
	}

	now := time.Now()

	tTimeFormat := "2006 15:04:00"
	t, err := time.ParseInLocation(tTimeFormat, fmt.Sprintf("%s %s", "2025", timeStr), tz)
	if err != nil {
		return t, errors.New(err)
	}

	t = time.Date(t.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, tz)

	return t, nil
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

func EncodeEsAfter(after []types.FieldValue) (string, error) {
	if len(after) == 0 {
		return "", nil
	}

	b, err := json.Marshal(after)
	if err != nil {
		return "", errors.New(err)
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func DecodeESAfter(after string) ([]types.FieldValue, error) {
	var res []types.FieldValue

	if after == "" {
		return res, nil
	}

	b, err := base64.StdEncoding.DecodeString(after)
	if err != nil {
		return res, errors.New(err)
	}

	if err := json.Unmarshal(b, &res); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func DateNowInUTC() time.Time {
	now := time.Now()
	d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return d
}

func GenerateInvoiceRef(id uuid.UUID) string {
	date := time.Now().Format("20060102")

	num := new(big.Int).SetBytes(id[:8])
	base36 := strings.ToUpper(num.Text(36))

	padded := fmt.Sprintf("%010s", base36)

	return fmt.Sprintf("%s-%s", date, padded)
}
