package timetools

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const layoutYear = "2006"
const layoutMonth = "01"
const layoutDay = "02"

const Timestamp = "timestamp"

type dateFormat struct {
	missDay   bool
	missMonth bool
	missYear  bool
	fmt       string
}

var timeFormats = []dateFormat{
	{true, true, true, "15:04"},
	{true, true, true, "15:04:05"},
	{true, true, true, "15:04:05.00"},

	{false, false, false, "2006-01-02T15:04"},
	{false, false, false, "2006-01-02T15:04:05"},
	{false, false, false, "2006-01-02 15:04"},
	{false, false, false, "2006-01-02 15:04:05"},

	{false, false, true, "Jan 02"},
	{false, false, true, "Jan 02 15:04"},
	{false, false, true, "Jan 02 15:04:05"},

	{false, false, true, "January 02"},
	{false, false, true, "January 02 15:04"},
	{false, false, true, "January 02 15:04:05"},
	{false, false, true, "January 02 15:04:05.00"},

	{false, false, false, "2006-01-02"},
}

func populateDateFormatWithMilliseconds(df dateFormat) {
	for i := 1; i <= 10; i++ {
		newDateFormat := df
		newDateFormat.fmt += "." + strings.Repeat("0", i)
		timeFormats = append(timeFormats, newDateFormat)
		newDateFormat.fmt += "Z"
		timeFormats = append(timeFormats, newDateFormat)
	}
}

func init() {
	populateDateFormatWithMilliseconds(dateFormat{false, false, false, "2006-01-02T15:04:05"})
	populateDateFormatWithMilliseconds(dateFormat{false, false, false, "2006-01-02 15:04:05"})
	populateDateFormatWithMilliseconds(dateFormat{false, false, true, "January 02 15:04:05"})
	populateDateFormatWithMilliseconds(dateFormat{false, false, true, "Jan 02 15:04:05"})
}

func ParseDate(str string, now time.Time) (time.Time, error) {
	var t time.Time
	var err error
	if str == "now" {
		return now, nil
	}

	r, _ := regexp.Compile("^-([0-9]+)([dhms])$")
	res := r.FindStringSubmatch(str)
	if len(res) != 0 {
		var subDuration time.Duration
		switch res[2] {
		case "d":
			subDuration = 24 * time.Hour
		case "h":
			subDuration = time.Hour
		case "m":
			subDuration = time.Minute
		case "s":
			subDuration = time.Second
		}

		multiplier, err := strconv.Atoi(res[1])
		if err != nil {
			panic(fmt.Sprintf("Failed to convert %s to int: %s", res[1], err.Error()))
		}

		subDuration = time.Duration(multiplier) * subDuration
		return now.Add(-1 * subDuration), nil
	}

	for _, tf := range timeFormats {
		date := str
		layout := tf.fmt
		_, err = time.Parse(layout, date)
		if err != nil {
			continue
		}

		// Layout "15:04" is treated like "0000-01-01 15:04:00"
		// golang's time is so weak and ugly
		// I could not find anything in golang as powefull as GNU date

		if tf.missYear {
			date = fmt.Sprintf("%d %s", now.Year(), date)
			layout = fmt.Sprintf("%s %s", layoutYear, layout)
		}

		if tf.missMonth {
			date = fmt.Sprintf("%02d %s", now.Month(), date)
			layout = fmt.Sprintf("%s %s", layoutMonth, layout)
		}

		if tf.missDay {
			date = fmt.Sprintf("%02d %s", now.Day(), date)
			layout = fmt.Sprintf("%s %s", layoutDay, layout)
		}

		t, err = time.Parse(layout, date)
		if err != nil {
			return t, fmt.Errorf("str=%s, layout=%s, %s", date, layout, err.Error())
		}

		return t, nil
	}

	return t, fmt.Errorf("no matching layout was found to parse %s", str)
}

func FormatDate(t time.Time, format string) string {
	switch format {
	case Timestamp, "ts", "unix timestamp":
		return fmt.Sprintf("%d", t.Unix())

	case "RFC3339", "rfc3339":
		return t.Format(time.RFC3339)
	}

	return t.Format(format)
}
