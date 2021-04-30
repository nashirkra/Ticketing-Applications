package helper

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

// Return Length and Boolean from given keys arguments
//  ex: key => "user1" || "*" || "user*"
//  @param c *redis.Client
//  @param key string
func LenKeys(ctx context.Context, c *redis.Client, key string) (int, bool) {
	var lenOfKeys int
	t := c.Do(ctx, "KEYS", key).Val()
	switch reflect.TypeOf(t).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(t)
		lenOfKeys = s.Len()
		/*
			for i := 0; i < s.Len(); i++ {
				fmt.Println(s.Index(i))
			} */
	}
	return lenOfKeys, (lenOfKeys > 0)
}

// NullTime for nil time.Time
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Valid = false
		return nil
	}
	nt.Time, nt.Valid = value.(time.Time), true
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

func (n NullTime) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Time)
	}
	return json.Marshal(nil)
}

func (n *NullTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		n.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &n.Time)
	if err == nil {
		n.Valid = true
	}
	return err
}

// Set Time UTC
var countryTz = map[string]string{
	"Hungary": "Europe/Budapest",
	"Egypt":   "Africa/Cairo",
	"Jakarta": "Asia/Bangkok",
}

func TimeIn(name string) time.Time {
	loc, err := time.LoadLocation(countryTz[name])
	if err != nil {
		panic(err)
	}
	return time.Now().In(loc)
}

// input date 2006-01-02
func ParseIdn(date string) int64 {
	loc, err := time.LoadLocation(countryTz["Jakarta"])
	if err != nil {
		panic(err)
	}
	parseTime, err := time.Parse("2006-01-02", date)
	if err != nil {
		fmt.Printf("parseTime error: %+v\n", err)
		fmt.Printf("date: %+v", date)
	}
	return parseTime.In(loc).UnixNano()
}

func ParseTime(locationName string, value string) time.Time {
	loc, err := time.LoadLocation(countryTz[locationName])
	if err != nil {
		panic(err)
	}
	t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", value)
	if err != nil {
		panic(err)
	}
	t.In(loc)
	return t
}

func HashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)

	if err != nil {
		log.Println(err)
		panic("Failed to hash a password")
	}
	return string(hash)
}
