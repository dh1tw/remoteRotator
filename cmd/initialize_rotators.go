package cmd

import (
	"fmt"
	"strings"

	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/dh1tw/remoteRotator/rotator/dummy"
	"github.com/dh1tw/remoteRotator/rotator/yaesu"
	"github.com/spf13/viper"
)

// init rotator initializes a rotator
func initRotator(rType string, eventHdlr rotator.EventHandler, errorCh chan struct{}) (rotator.Rotator, error) {

	switch strings.ToUpper(rType) {

	case "YAESU":
		evHandler := yaesu.EventHandler(eventHdlr)
		name := yaesu.Name(viper.GetString("rotator.name"))
		interval := yaesu.UpdateInterval(viper.GetDuration("rotator.pollingrate"))
		spPortName := yaesu.Portname(viper.GetString("rotator.portname"))
		baudrate := yaesu.Baudrate(viper.GetInt("rotator.baudrate"))
		hasAzimuth := yaesu.HasAzimuth(viper.GetBool("rotator.has-azimuth"))
		hasElevation := yaesu.HasElevation(viper.GetBool("rotator.has-elevation"))
		azMin := yaesu.AzimuthMin(viper.GetInt("rotator.azimuth-min"))
		azMax := yaesu.AzimuthMax(viper.GetInt("rotator.azimuth-max"))
		elMin := yaesu.ElevationMin(viper.GetInt("rotator.elevation-min"))
		elMax := yaesu.ElevationMax(viper.GetInt("rotator.elevation-max"))
		azStop := yaesu.AzimuthStop(viper.GetInt("rotator.azimuth-stop"))
		errorCh := yaesu.ErrorCh(errorCh)

		yaesu, err := yaesu.New(name, interval, evHandler,
			spPortName, baudrate, hasAzimuth, hasElevation, azMin, azMax, elMin,
			elMax, azStop, errorCh)

		if err != nil {
			return nil, err
		}
		return yaesu, err

	case "DUMMY":
		evHandler := dummy.EventHandler(eventHdlr)
		name := dummy.Name(viper.GetString("rotator.name"))
		hasAzimuth := dummy.HasAzimuth(viper.GetBool("rotator.has-azimuth"))
		hasElevation := dummy.HasElevation(viper.GetBool("rotator.has-elevation"))
		azMin := dummy.AzimuthMin(viper.GetInt("rotator.azimuth-min"))
		azMax := dummy.AzimuthMax(viper.GetInt("rotator.azimuth-max"))
		elMin := dummy.ElevationMin(viper.GetInt("rotator.elevation-min"))
		elMax := dummy.ElevationMax(viper.GetInt("rotator.elevation-max"))
		azStop := dummy.AzimuthStop(viper.GetInt("rotator.azimuth-stop"))

		dummyRotator, err := dummy.New(name, evHandler, hasAzimuth, hasElevation, azMin, azMax, azStop, elMin, elMax)
		if err != nil {
			return nil, err
		}
		return dummyRotator, err

	default:
		return nil, fmt.Errorf("unknown rotator type (%v)", rType)
	}
}
