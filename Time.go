package vutils

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp/syntax"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"
)

type timeUtils struct{}

func (tu *timeUtils) IsDaylightSavingsTime(inTime time.Time) bool {
	_, timeOffset := inTime.Zone() //with time already in correct myLocation
	_, winterOffset := time.Date(inTime.Year(), 1, 1, 0, 0, 0, 0, inTime.Location()).Zone()
	_, summerOffset := time.Date(inTime.Year(), 7, 1, 0, 0, 0, 0, inTime.Location()).Zone()

	if winterOffset > summerOffset {
		winterOffset, summerOffset = summerOffset, winterOffset
	}

	if winterOffset != summerOffset { // the location has daylight saving
		if timeOffset != winterOffset {
			return true
		}
	}
	return false
}

func (tu *timeUtils) GetDaylightSavingsTimeOffset(inTime time.Time) time.Duration {
	_, timeOffset := inTime.Zone() //with time already in correct myLocation
	winterTime := time.Date(inTime.Year(), 1, 1, 0, 0, 0, 0, inTime.Location())
	summerTime := time.Date(inTime.Year(), 7, 1, 0, 0, 0, 0, inTime.Location())
	_, winterOffset := winterTime.Zone()
	_, summerOffset := summerTime.Zone()

	if winterOffset > summerOffset {
		winterOffset, summerOffset = summerOffset, winterOffset
	}

	if winterOffset != summerOffset { // the location has daylight saving
		if timeOffset != winterOffset {
			return summerTime.Sub(winterTime)
		}
	}
	return time.Duration(0)
}

var zoneDirs = map[string]string{
	"android":   "/system/usr/share/zoneinfo/",
	"darwin":    "/usr/share/zoneinfo/",
	"dragonfly": "/usr/share/zoneinfo/",
	"freebsd":   "/usr/share/zoneinfo/",
	"linux":     "/usr/share/zoneinfo/",
	"netbsd":    "/usr/share/zoneinfo/",
	"openbsd":   "/usr/share/zoneinfo/",
	// "plan9":"/adm/timezone/", -- no way to test this platform
	"solaris": "/usr/share/lib/zoneinfo/",
	"windows": `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Time Zones\`,
}

// inSlice ... check if an element is inside a slice
func inSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// readTZFile ... read timezone file and append into timeZones slice
func readTZFile(tzList *[]string, zoneDir string, path string) {
	files, _ := ioutil.ReadDir(zoneDir + path)
	for _, f := range files {
		if f.Name() != strings.ToUpper(f.Name()[:1])+f.Name()[1:] {
			continue
		}
		if f.IsDir() {
			readTZFile(tzList, zoneDir, path+"/"+f.Name())
		} else {
			tz := (path + "/" + f.Name())[1:]
			// check if tz is already in timeZones slice
			// append if not
			if !inSlice(tz, *tzList) { // need a more efficient method...

				// convert string to rune
				tzRune, _ := utf8.DecodeRuneInString(tz[:1])

				if syntax.IsWordChar(tzRune) { // filter out entry that does not start with A-Za-z such as +VERSION
					*tzList = append(*(tzList), tz)
				}
			}
		}
	}

}

func (tu *timeUtils) ListTimeZones() ([]string, error) {
	tzList := []string{}
	if runtime.GOOS == "nacl" || runtime.GOOS == "" {
		return nil, errors.New("Unsupported platform")
	}

	if runtime.GOOS != "windows" {
		for _, zoneDir := range zoneDirs {
			readTZFile(&tzList, zoneDir, "")
		}
	} else { // let's handle Windows
		// if you're building this on darwin/linux
		// chances are you will encounter
		// undefined: registry in registry.OpenKey error message
		// uncomment below if compiling on Windows platform

		//k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Time Zones`, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)

		//if err != nil {
		// fmt.Println(err)
		//}
		//defer k.Close()

		//names, err := k.ReadSubKeyNames(-1)
		//if err != nil {
		// fmt.Println(err)
		//}

		//fmt.Println("Number of timezones : ", len(names))
		//for i := 0; i <= len(names)-1; i++ {
		// check if tz is already in timeZones slice
		// append if not
		// if !inSlice(names[i], timeZones) { // need a more efficient method...
		//  timeZones = append(timeZones, names[i])
		// }
		//}

		// UPDATE : Reading from registry is not reliable
		// better to parse output result by "tzutil /g" command
		// REMEMBER : There is no time difference between Coordinated Universal Time and Greenwich Mean Time ....
		cmd := exec.Command("tzutil", "/l")

		data, err := cmd.Output()

		if err != nil {
			return nil, err
		}

		fmt.Println("UTC is the same as GMT")
		fmt.Println("There is no time difference between Coordinated Universal Time and Greenwich Mean Time ....")
		GMTed := bytes.Replace(data, []byte("UTC"), []byte("GMT"), -1)

		fmt.Println(string(GMTed))

	}

	now := time.Now()

	for _, v := range tzList {

		if runtime.GOOS != "windows" {

			location, err := time.LoadLocation(v)
			if err != nil {
				fmt.Println(err)
			}

			// extract the GMT
			t := now.In(location)
			t1 := fmt.Sprintf("%s", t.Format(time.RFC822Z))
			tArray := strings.Fields(t1)
			gmtTime := strings.Join(tArray[4:], "")
			hours := gmtTime[0:3]
			minutes := gmtTime[3:]

			gmt := "GMT" + fmt.Sprintf("%s:%s", hours, minutes)
			fmt.Println(gmt + " " + v)

		} else {
			fmt.Println(v)
		}

	}
	return tzList, nil
}

var Time = &timeUtils{}
