package main

import "time"

type fileInfo struct {
	filePath string
	cTime    time.Time
	mTime    time.Time
	aTime    time.Time
}

var (
	keys = []string{
		"@C", // create time
		"@M", // modify time
		"@A", // access time
	}
	values = append(keys,
		"@@", // auto choose earliest time
	)
)

func checkKey(s string) bool {
	for _, k := range keys {
		if s == k {
			return true
		}
	}
	return false
}

func checkValue(s string) bool {
	for _, k := range values {
		if s == k {
			return true
		}
	}
	if _, err := time.ParseInLocation("2006/01/02 15:04:05", s, time.Local); err == nil {
		return true
	}
	return false
}
