package headers

import "time"

const (
	TIME_LAYOUT_NANO = "2006-01-02 15:04:05.000"
	TIME_LAYOUT = "2006-01-02 15:04:05"
)

var(
	loc *time.Location
	err error
)

func init()  {
	loc, err = time.LoadLocation("Local")
}

func ToString(timeTime time.Time, unit string) string {
	return timeTime.Format(unit)
}

func ToStringInt(timeInt int64, unit string)  string{
	return time.Unix(0, timeInt).Format(unit)
}

func ToLocalTime(timeTime time.Time) time.Time {
	return timeTime.In(loc)
}
