package cmd

import (
	"log"
	// "fmt"
	// "io"
	// "compress/gzip"
	// "bytes"
	// "encoding/base64"
	"strings"
	"encoding/hex"
	"crypto/md5"
	"time"
)

//Time handling
const (
	millisPerSecond     = int64(time.Second / time.Millisecond)
	nanosPerMillisecond = int64(time.Millisecond / time.Nanosecond)
	nanosPerSecond      = int64(time.Second / time.Nanosecond)
)

//NsToTime -
func NsToTime(ns int64) time.Time {
	secs := ns/nanosPerSecond
	nanos := ns - secs * nanosPerSecond
	return time.Unix(secs, nanos)
}

//MsToTime -
func MsToTime(ms int64) time.Time {
	secs := ms/millisPerSecond
	nanos := (ms - secs * millisPerSecond) * nanosPerMillisecond
	return time.Unix(secs, nanos)
}

//CreateHash - md5
func CreateHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}
//DecodeJenkinsConsoleNote - See https://github.com/LarrysGIT/Extract-Jenkins-Raw-Log/blob/master/README.md. Not working yet TODO
func DecodeJenkinsConsoleNote(consoleNote string) (string) {
	return consoleNote
	// PREAMBLE_STR := "\u001B[8mha:"
	// POSTAMBLE_STR := "\u001B[0m"
	// PREAMBLE_STR := `[8mha:`
	// POSTAMBLE_STR := `[0m`
	// pos := strings.Index(PREAMBLE_STR, consoleNote)
	// fmt.Printf("Call DecodeJenkinsConsoleNote pos %d\n%s", pos, consoleNote)
	// if pos == -1 { return consoleNote }
	// pos1 := strings.Index(POSTAMBLE_STR, consoleNote)
	// if pos == -1 { return consoleNote }
	// posStartExtract := pos + len(PREAMBLE_STR)
	// posEndExtract := pos1 - len(POSTAMBLE_STR) - 1
	// data := consoleNote[posStartExtract:posEndExtract]
	// fmt.Print(data)
	// dataByte, err := base64.StdEncoding.DecodeString(data)
	// dataByte1 := dataByte[40:]
	// var buf bytes.Buffer
	// buf.Write(dataByte1)

	// zr, err := gzip.NewReader(&buf)
	// if err != nil {
	// 	fmt.Printf("ERROR Can not ungzip. Will return string as is - %v\n", err)
	// 	return consoleNote
	// }

	// var buf1 bytes.Buffer
	// if _, err := io.Copy(&buf1, zr); err != nil {
	// 	fmt.Printf("ERROR can not copy uncompressed data - %v\n", err)
	// }

	// if err := zr.Close(); err != nil {
	// 	fmt.Printf("ERROR can not close the gzip reader - %v\n", err)
	// }

	// return consoleNote[0:(pos-1)] + buf1.String() + consoleNote[(pos1 + 1):]
}

//CheckAuthorizedUser -
func CheckAuthorizedUser(email string) (bool) {
	conn := GetDBConn()
	defer conn.Close()
	q := "SELECT email FROM user where email = '" + email + "';"
	stmt, err := conn.Prepare(q)
	if err != nil {
		log.Printf("ERROR - %v\n", err)
	}
	defer stmt.Close()

	hasRow, err := stmt.Step()
	if !hasRow {
		return CheckAuthorizedDomain(email)
	} else {
		return true
	}
}

//CheckAuthorizedDomain -
func CheckAuthorizedDomain(email string) (bool) {
	_tmp := strings.Split(email, "@")
	ok, e := Config.AuthorizedDomain[_tmp[1]]
	return ok && e
}