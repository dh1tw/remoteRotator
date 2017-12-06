package yaesu

import (
	"bytes"
	"regexp"
	"testing"

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
	p.rxBuf = nil
	p.sendBuf = nil
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
			if bytes.Compare(tc.expMsg, res) != 0 {
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
		name     string
		value    int
		expValue int
		expMsg   []byte
	}{
		{"150 deg", 150, 150, []byte("N150\r\n")},
		{"451 deg", 181, 180, []byte("N180\r\n")},
		{"-100 deg", -100, 0, []byte("N000\r\n")},
		{"1000 deg", 1000, 180, []byte("N180\r\n")},
	}

	for _, tc := range tt {

		dp := dummyPort{
			sendBuf: &bytes.Buffer{},
			rxBuf:   &bytes.Buffer{},
		}

		yaesu := Yaesu{
			hasElevation: true,
			sp:           &dp,
		}

		t.Run(tc.name, func(t *testing.T) {
			err := yaesu.SetElevation(tc.value)
			if err != nil {
				t.Fatalf("unable to set elevation to %v; got error: %q", tc.name, err)
			}
			res := dp.sendBuf.Bytes()
			if bytes.Compare(tc.expMsg, res) != 0 {
				t.Fatalf("expecting '%s' (Hex: % 02x) to be sent to the serial port. Instead got '%s' (Hex: % 02x)",
					replaceLineBreaks(tc.expMsg),
					replaceLineBreaks(tc.expMsg),
					replaceLineBreaks(res),
					replaceLineBreaks(res))
			}
			if yaesu.ElPreset() != tc.expValue {
				t.Fatalf("expecting elevation preset %v, but got %v", yaesu.ElPreset(), tc.expValue)
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

				if bytes.Compare(dp.sendBuf.Bytes(), tc.expMsg) != 0 {
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

				if bytes.Compare(dp.sendBuf.Bytes(), tc.expMsg) != 0 {
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

				if bytes.Compare(dp.sendBuf.Bytes(), tc.expMsg) != 0 {
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

	bothCb := func(r rotator.Rotator, ev rotator.Event, v ...interface{}) {
		if ev != rotator.Azimuth && ev != rotator.Elevation {
			t.Fatalf("expected event 'Azimuth' or 'Elevation', got: %v", ev)
		}
		az := v[0].(rotator.Status).Azimuth
		el := v[0].(rotator.Status).Elevation
		if az <= 0 {
			t.Fatalf("expected value must be > 0, got %v", az)
		}
		if el < 0 {
			t.Fatalf("expected value must be > 0, got %v", el)
		}

	}

	tt := []struct {
		name      string
		input     string
		evHandler func(rotator.Rotator, rotator.Event, ...interface{})
	}{
		{"azimuth", "+0030", bothCb},
		{"elevation", "+0030+0090", bothCb},
		{"prompt", "?>", nil},
		{"garbage", "der43$§PkoJOIo;\n\r", nil},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			headingPattern, err := regexp.Compile("[\\d]{4}")
			if err != nil {
				t.Fatalf(err.Error())
			}
			yaesu := &Yaesu{
				eventHandler:   tc.evHandler,
				headingPattern: headingPattern,
			}
			yaesu.parseMsg(tc.input)
		})
	}
}

func TestSetValueAndCallEvent(t *testing.T) {

	azCb := func(r rotator.Rotator, ev rotator.Event, v ...interface{}) {
		if ev != rotator.Azimuth {
			t.Fatalf("expected event 'Azimuth', got: %v", ev)
		}
		if v[0].(rotator.Status).Azimuth != 30 {
			t.Fatalf("expected value must be 30, got %v", v[0])
		}
	}

	elCb := func(r rotator.Rotator, ev rotator.Event, v ...interface{}) {
		if ev != rotator.Elevation {
			t.Fatalf("expected event 'Elevation', got: %v", ev)
		}
		if v[0].(rotator.Status).Elevation != 60 {
			t.Fatalf("expected value must be 30, got %v", v[0])
		}
	}

	tt := []struct {
		name      string
		event     rotator.Event
		value     int
		evHandler func(rotator.Rotator, rotator.Event, ...interface{})
	}{
		{"azimuth", rotator.Azimuth, 30, azCb},
		{"elevation", rotator.Elevation, 60, elCb},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			yaesu := &Yaesu{
				eventHandler: tc.evHandler,
			}
			yaesu.setValueAndCallEvent(tc.event, tc.value)

			if tc.event == rotator.Azimuth {
				if yaesu.Azimuth() != tc.value {
					t.Fatalf("expected %v value %d, but got %d",
						tc.name, tc.value, yaesu.Azimuth())
				}
			}

			if tc.event == rotator.Elevation {
				if yaesu.Elevation() != tc.value {
					t.Fatalf("expected %v value %d, but got %d",
						tc.name, tc.value, yaesu.Azimuth())
				}
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
	if bytes.Compare(value, expValue) != 0 {
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
		portName func(*Yaesu)
		expError string
		altError string
	}{
		{"port does not exist", Portname("/dev/ttyXXXXX"), "open /dev/ttyXXXXX: no such file or directory", ""},
		{"invalid serial port", Portname("/dev/null"), "File is not a tty", "inappropriate ioctl for device"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewYaesu(tc.portName)
			if err.Error() != tc.expError {
				if err.Error() != tc.altError {
					t.Fatalf("expected error '%s', got '%s'", tc.expError, err.Error())
				}
			}
		})
	}
}
