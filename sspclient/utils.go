package sspclient

import (
	"regexp"
	"strconv"
	"time"
)

// startTime — точка отсчёта для генерации идентификаторов.
var startTime = time.Date(2021, 1, 1, 0, 0, 1, 0, time.UTC)

var validAdrRe = regexp.MustCompile(`\(\d(\d\s|\s\d)\d{4}\s\d\)\d\d`)

// AddZeroesToStr добавляет ведущие нули к строке до нужной длины.
func AddZeroesToStr(str string, targetlen int) string {
	s := str
	for len([]rune(s)) < targetlen {
		s = "0" + s
	}
	return s
}

// ValidSSPAdr проверяет корректность SSP-адреса.
func ValidSSPAdr(str string) bool {
	if len([]rune(str)) != 13 {
		return false
	}
	return validAdrRe.MatchString(str)
}

// generateKvit генерирует квитанцию по идентификатору.
func generateKvit(id int64) string {
	return AddZeroesToStr(strconv.FormatInt(id, 10), 3) +
		AddZeroesToStr(strconv.FormatInt(time.Now().Unix()-startTime.Unix(), 10), 9)
}

// generateMessageID генерирует ID сообщения.
func generateMessageID(id int64) string {
	return "#" + AddZeroesToStr(strconv.FormatInt(id, 10), 5) + "-" +
		AddZeroesToStr(strconv.FormatInt(time.Now().Unix()-startTime.Unix(), 10), 19)
}
