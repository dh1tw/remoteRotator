package yaesu

import (
	"bytes"
	"reflect"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/dh1tw/remoteRotator/rotator"
)

type dummyPort struct {
	sendBuf *bytes.Buffer
	rxBuf   *bytes.Buffer
}

func (p *dummyPort) Read(b []byte) (int, error) {
	a, err := p.rxBuf.Read(b)
	return a, err
}

func (p *dummyPort) Write(b []byte) (int, error) {
	a, err := p.sendBuf.Write(b)
	return a, err
}

func (p *dummyPort) Flush() error {
	return nil
}

func (p *dummyPort) Close() error {
	return nil
}

func TestHasAzimuth(t *testing.T) {
	yaesu := Yaesu{
		hasAzimuth: true,
	}

	if yaesu.HasAzimuth() != true {
		t.Error("should be true")
	}
}

func TestHasElevation(t *testing.T) {
	yaesu := Yaesu{
		hasElevation: true,
	}

	if yaesu.HasElevation() != true {
		t.Error("should be true")
	}
}

func TestElevation(t *testing.T) {
	yaesu := Yaesu{
		elevation: 150,
	}

	if yaesu.Elevation() != 150 {
		t.Error("should return 150")
	}
}

func TestAzimuth(t *testing.T) {
	yaesu := Yaesu{
		azimuth: 340,
	}

	if yaesu.Azimuth() != 340 {
		t.Error("should return 340")
	}
}

func TestSetAzimuth(t *testing.T) {

	tt := []struct {
		name     string
		value    int
		expValue int
		expMsg   []byte
	}{
		{"150 deg", 150, 150, []byte("M150\r\n")},
		{"451 deg", 451, 450, []byte("M450\r\n")},
		{"-100 deg", -100, 0, []byte("M000\r\n")},
		{"1000 deg", 1000, 450, []byte("M450\r\n")},
	}

	for _, tc := range tt {

		dp := dummyPort{
			sendBuf: &bytes.Buffer{},
			rxBuf:   &bytes.Buffer{},
		}

		yaesu := Yaesu{
			hasAzimuth: true,
			sp:         &dp,
		}

		t.Run(tc.name, func(t *testing.T) {
			err := yaesu.SetAzimuth(tc.value)
			if err != nil {
				t.Fatalf("unable to set azimuth to %v; got error: %q", tc.name, err)
			}
			res := dp.sendBuf.Bytes()
			if !bytes.Equal(tc.expMsg, res) {
				t.Fatalf("expecting '%s' (Hex: % 02x) to be sent to the serial port. Instead got '%s' (Hex: % 02x)",
					replaceLineBreaks(tc.expMsg),
					replaceLineBreaks(tc.expMsg),
					replaceLineBreaks(res),
					replaceLineBreaks(res))
			}
			if yaesu.AzPreset() != tc.expValue {
				t.Fatalf("expecting azimuth preset %v, but got %v", yaesu.AzPreset(), tc.expValue)
			}
		})
	}
}

func TestSetAzimuthButNotEnabled(t *testing.T) {
	dp := dummyPort{
		sendBuf: &bytes.Buffer{},
		rxBuf:   &bytes.Buffer{},
	}

	yaesu := Yaesu{
		hasAzimuth: false,
		sp:         &dp,
	}

	v := 200

	err := yaesu.SetAzimuth(v)
	if err != nil {
		t.Fatal(err)
	}

	if yaesu.Azimuth() == v {
		t.Fatal("azimuth must not be set if not enabled")
	}
}

func TestSetElevation(t *testing.T) {

	tt := []struct {
		name       string
		azValue    int
		elValue    int
		expAzValue int
		expElValue int
		expMsg     []byte
	}{
		{"azimuth 45deg, elevation 150 deg", 45, 150, 45, 150, []byte("W045 150\r\n")},
		{"azimuth 45deg, elevation 451 deg (positive out of range)", 45, 181, 45, 180, []byte("W045 180\r\n")},
		{"azimuth 45deg, elevation -100 deg (negative out of range)", 45, -100, 45, 0, []byte("W045 000\r\n")},
		{"azimuth 45deg, elevation 1000 deg (positive out of range)", 45, 1000, 45, 180, []byte("W045 180\r\n")},
		{"azimuth 45deg, elevation 45 deg", 45, 45, 45, 45, []byte("W045 045\r\n")},
	}

	for _, tc := range tt {

		dp := dummyPort{
			sendBuf: &bytes.Buffer{},
			rxBuf:   &bytes.Buffer{},
		}

		yaesu := Yaesu{
			azPreset:     45,
			elPreset:     0,
			hasElevation: true,
			sp:           &dp,
		}

		t.Run(tc.name, func(t *testing.T) {
			err := yaesu.SetElevation(tc.elValue)
			if err != nil {
				t.Fatalf("unable to set elevation to %v; got error: %q", tc.name, err)
			}
			res := dp.sendBuf.Bytes()
			if !bytes.Equal(tc.expMsg, res) {
				t.Fatalf("expecting '%s' (Hex: % 02x) to be sent to the serial port. Instead got '%s' (Hex: % 02x)",
					replaceLineBreaks(tc.expMsg),
					replaceLineBreaks(tc.expMsg),
					replaceLineBreaks(res),
					replaceLineBreaks(res))
			}
			if yaesu.ElPreset() != tc.expElValue {
				t.Fatalf("expecting elevation preset %v, but got %v", yaesu.ElPreset(), tc.expElValue)
			}
			if yaesu.AzPreset() != tc.expAzValue {
				t.Fatalf("expecting azimuth preset %v, but got %v", yaesu.AzPreset(), tc.expAzValue)
			}
		})
	}
}

func TestSetElevationButNotEnabled(t *testing.T) {
	dp := dummyPort{
		sendBuf: &bytes.Buffer{},
		rxBuf:   &bytes.Buffer{},
	}

	yaesu := Yaesu{
		hasElevation: false,
		sp:           &dp,
	}

	v := 95

	err := yaesu.SetElevation(v)
	if err != nil {
		t.Fatal(err)
	}

	if yaesu.Elevation() == v {
		t.Fatal("elevation must not be set if not enabled")
	}
}

func TestRotatorStop(t *testing.T) {

	tt := []struct {
		name     string
		value    int
		preset   int
		expMsg   []byte
		stopFunc string
	}{
		{"stop azimuth", 120, 20, []byte("A\r\n"), "azimuth"},
		{"stop elevation", 120, 20, []byte("E\r\n"), "elevation"},
		{"stop", 120, 20, []byte("S\r\n"), "both"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			dp := dummyPort{
				sendBuf: &bytes.Buffer{},
				rxBuf:   &bytes.Buffer{},
			}

			yaesu := Yaesu{
				azimuth:   tc.value,
				azPreset:  tc.preset,
				elevation: tc.value,
				elPreset:  tc.preset,
				sp:        &dp,
			}
			switch tc.stopFunc {
			case "azimuth":
				err := yaesu.StopAzimuth()
				if err != nil {
					t.Fatalf("unable to %v", tc.name)
				}

				if !bytes.Equal(dp.sendBuf.Bytes(), tc.expMsg) {
					send := replaceLineBreaks(dp.sendBuf.Bytes())
					exp := replaceLineBreaks(tc.expMsg)
					t.Fatalf("expecting '%s' (Hex: % 02x) to be sent to the serial port. Instead got '%s' (Hex: % 02x)",
						exp, exp, send, send)
				}

				if yaesu.Azimuth() != yaesu.AzPreset() {
					t.Fatalf("expected azimuth and azPreset to be equal, got: %d, %d", tc.value, tc.preset)
				}
			case "elevation":
				err := yaesu.StopElevation()
				if err != nil {
					t.Fatalf("unable to %v", tc.name)
				}

				if !bytes.Equal(dp.sendBuf.Bytes(), tc.expMsg) {
					send := replaceLineBreaks(dp.sendBuf.Bytes())
					exp := replaceLineBreaks(tc.expMsg)
					t.Fatalf("expecting '%s' (Hex: % 02x) to be sent to the serial port. Instead got '%s' (Hex: % 02x)",
						exp, exp, send, send)
				}

				if yaesu.Elevation() != yaesu.ElPreset() {
					t.Fatalf("expected elevation and elPreset to be equal, got: %d, %d", tc.value, tc.preset)
				}
			case "both":
				if err := yaesu.Stop(); err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(dp.sendBuf.Bytes(), tc.expMsg) {
					send := replaceLineBreaks(dp.sendBuf.Bytes())
					exp := replaceLineBreaks(tc.expMsg)
					t.Fatalf("expecting '%s' (Hex: % 02x) to be sent to the serial port. Instead got '%s' (Hex: % 02x)",
						exp, exp, send, send)
				}

				if yaesu.Azimuth() != yaesu.AzPreset() {
					t.Fatalf("expected azimuth and azPreset to be equal, got: %d, %d", tc.value, tc.preset)
				}
				if yaesu.Elevation() != yaesu.ElPreset() {
					t.Fatalf("expected elevation and elPreset to be equal, got: %d, %d", tc.value, tc.preset)
				}
			}
		})
	}
}

func TestParseMsg(t *testing.T) {

	tt := []struct {
		name          string
		input         string
		azInitialized bool
		elInitialized bool
		azimuth       int
		elevation     int
		updateNeeded  bool
	}{
		{"azimuth - not initialized", "+0030", false, false, 0, 0, true},
		{"azimuth - not initialized but same position", "+0030", false, false, 30, 0, true},
		{"azimuth - initialized and new position", "+0030", true, false, 45, 0, true},
		{"azimuth - initialized and same position - no update needed", "+0030", true, false, 30, 0, false},
		{"azimuth and elevation - not initialized", "+0030+0090", false, false, 20, 30, true},
		{"azimuth and elevation - initialized and new position", "AZ=030 EL=090", true, true, 10, 0, true},
		{"azimuth and elevation - initialized but same position - no update needed", "AZ=030 EL=090", true, true, 30, 90, false},
		{"prompt", "?>", true, true, 0, 0, false},
		{"garbage", "der43$§PkoJOIo;\n\r", true, true, 0, 0, false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			doneCh := make(chan struct{})

			updateCb := func(rotator.Rotator, rotator.Heading) {
				close(doneCh)
			}

			yaesu := &Yaesu{
				eventHandler:         updateCb,
				azInitialized:        tc.azInitialized,
				elInitialized:        tc.elInitialized,
				azimuth:              tc.azimuth,
				elevation:            tc.elevation,
				headingPatternGS232A: getProtocolRegExp("GS232A", t),
				headingPatternGS232B: getProtocolRegExp("GS232B", t),
			}

			yaesu.parseMsg(tc.input)
			updateCalled := false

			select {
			case <-doneCh:
				updateCalled = true
			case <-time.After(time.Millisecond * 100):
				updateCalled = false
			}

			// ensure the eventHandler / updateCallback only get's executed
			// when needed
			if updateCalled != tc.updateNeeded {
				t.Fatalf("failure in callback execution")
			}

		})
	}
}

func TestParseGS232A(t *testing.T) {

	tt := []struct {
		name   string
		input  string
		output map[string]int
	}{
		{"azimuth", "+0030", map[string]int{"azimuth": 30}},
		{"azimuth and elevation", "+0030+0090", map[string]int{"azimuth": 30, "elevation": 90}},
		{"azimuth GS232B", "AZ=040", map[string]int{}},
		{"prompt", "?>", map[string]int{}},
		{"garbage", "der43$§PkoJOIo;\n\r", map[string]int{}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			yaesu := &Yaesu{
				headingPatternGS232A: getProtocolRegExp("GS232A", t),
			}
			res := yaesu.parseGS232A(tc.input)
			if !reflect.DeepEqual(res, tc.output) {
				t.Fatalf("GS232A parser error. expected %v, but got %v", tc.output, res)
			}
		})
	}
}

func TestParseGS232B(t *testing.T) {

	tt := []struct {
		name   string
		input  string
		output map[string]int
	}{
		{"azimuth", "AZ=030", map[string]int{"azimuth": 30}},
		{"elevation", "EL=090", map[string]int{"elevation": 90}},
		{"azimuth and elevation", "AZ=030 EL=090", map[string]int{"azimuth": 30, "elevation": 90}},
		{"azimuth and elevation wide spacing", "AZ=030    EL=090", map[string]int{"azimuth": 30, "elevation": 90}},
		{"azimuth GS232A", "+0030", map[string]int{}},
		{"prompt", "?>", map[string]int{}},
		{"garbage", "der43$§PkoJOIo;\n\r", map[string]int{}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			yaesu := &Yaesu{
				headingPatternGS232B: getProtocolRegExp("GS232B", t),
			}
			res := yaesu.parseGS232B(tc.input)
			if !reflect.DeepEqual(res, tc.output) {
				t.Fatalf("GS232B parser error. expected %v, but got %v", tc.output, res)
			}
		})
	}
}

func TestQuery(t *testing.T) {
	dp := &dummyPort{
		sendBuf: &bytes.Buffer{},
		rxBuf:   &bytes.Buffer{},
	}

	yaesu := &Yaesu{
		sp: dp,
	}

	if err := yaesu.query(); err != nil {
		t.Fatalf("unable to send query; %v", err)
	}

	value := dp.sendBuf.Bytes()
	expValue := []byte("C2\r\n")
	if !bytes.Equal(value, expValue) {
		v := replaceLineBreaks(value)
		exp := replaceLineBreaks(expValue)
		t.Fatalf("expected '%s', got %s", exp, v)
	}
}

func TestRead(t *testing.T) {
	dp := &dummyPort{
		sendBuf: &bytes.Buffer{},
		rxBuf:   &bytes.Buffer{},
	}

	yaesu := &Yaesu{
		sp: dp,
	}

	expValue := "+0300+0150\r\n"
	dp.rxBuf.WriteString(expValue)

	res, err := yaesu.read()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if res != expValue {
		t.Fatalf("expected %s, got %s", expValue, res)
	}

	// if dp.rxBuf != nil || dp.sendBuf != nil {
	// 	t.Fatalf("close not called correctly")
	// }
}

func replaceLineBreaks(input []byte) []byte {
	s := bytes.Replace(input, []byte("\n"), []byte("\\n"), -1)
	return bytes.Replace(s, []byte("\r"), []byte("\\r"), -1)
}

func TestNewYaesuPortNotExist(t *testing.T) {

	tt := []struct {
		name     string
		os       string
		portName func(*Yaesu)
		expError string
	}{
		{"port does not exist", "linux", Portname("/dev/ttyXXXXX"), "open /dev/ttyXXXXX: no such file or directory"},
		{"port does not exist", "darwin", Portname("/dev/ttyXXXXX"), "open /dev/ttyXXXXX: no such file or directory"},
		{"port does not exist", "windows", Portname("/dev/ttyXXXXX"), "The system cannot find the path specified."},
		{"invalid serial port", "linux", Portname("/dev/null"), "inappropriate ioctl for device"},
		{"invalid serial port", "darwin", Portname("/dev/null"), "File is not a tty"},
		{"invalid serial port", "windows", Portname("/dev/null"), "The system cannot find the path specified."},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if runtime.GOOS != tc.os {
				t.Skip()
			}
			_, err := New(tc.portName)
			if err.Error() != tc.expError {
				t.Fatalf("expected error '%s', got '%s'", tc.expError, err.Error())
			}
		})
	}
}

func getProtocolRegExp(rotType string, t *testing.T) *regexp.Regexp {

	var pattern *regexp.Regexp = nil

	switch rotType {
	case "GS232A":
		p, err := regexp.Compile(`\+[\d]{4}`)
		if err != nil {
			t.Fatal("unable to compile gs232 regexp")
		}
		pattern = p
	case "GS232B":
		p, err := regexp.Compile(`((AZ)|(EL))=[\d]{3}`)
		if err != nil {
			t.Fatal("unable to compile gs232 regexp")
		}
		pattern = p
	}

	return pattern
}
