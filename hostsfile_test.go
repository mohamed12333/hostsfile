package hostsfile

import (
	"bytes"
	"fmt"
	"net"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\texp: %#v\n\tgot: %#v\033[39m\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func TestDecode(t *testing.T) {
	sampledata := "127.0.0.1 foobar\n# this is a comment\n10.0.0.1 anotheralias"
	h, err := Decode(strings.NewReader(sampledata))
	if err != nil {
		t.Error(err.Error())
	}
	firstRecord := h.records[0]

	assert(t, firstRecord.IpAddress.Equal(net.ParseIP("127.0.0.1")), "IP address should have been 127.0.0.1, was %s", firstRecord.IpAddress)
	equals(t, firstRecord.Hostnames["foobar"], true)
	equals(t, len(firstRecord.Hostnames), 1)

	aliasSample := "127.0.0.1 name alias1 alias2 alias3"
	h, err = Decode(strings.NewReader(aliasSample))
	ok(t, err)
	hns := h.records[0].Hostnames
	equals(t, len(hns), 4)
	equals(t, hns["alias3"], true)

	badline := strings.NewReader("blah")
	h, err = Decode(badline)
	if err == nil {
		t.Error("expected Decode(\"blah\") to return invalid, got no error")
	}
	if err.Error() != "Invalid hostsfile entry: blah" {
		t.Errorf("expected Decode(\"blah\") to return invalid, got %s", err.Error())
	}
}

var sampleHostsfile = Hostsfile{
	records: []Record{
		Record{
			IpAddress: net.ParseIP("127.0.0.1"),
			Hostnames: map[string]bool{"foobar": true},
		},
		Record{
			IpAddress: net.ParseIP("192.168.0.1"),
			Hostnames: map[string]bool{"bazbaz": true},
		},
	},
}

func TestEncode(t *testing.T) {
	b := new(bytes.Buffer)
	err := Encode(b, sampleHostsfile)
	ok(t, err)
	equals(t, b.String(), "127.0.0.1 foobar\n192.168.0.1 bazbaz\n")
}

func TestSet(t *testing.T) {
	hCopy := sampleHostsfile
	hCopy.Set(net.ParseIP("10.0.0.1"), "tendot")
	equals(t, len(hCopy.records), 3)
	equals(t, len(sampleHostsfile.records), 2)
}