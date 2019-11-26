package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func sanityCheckRotatorInputs() error {

	if len(viper.GetString("rotator.name")) == 0 {
		return fmt.Errorf("rotator name must not be empty")
	}

	forbiddenChars := "./\\_"
	if strings.ContainsAny(viper.GetString("rotator.name"), forbiddenChars) {
		return fmt.Errorf("rotator name must not contain '.', '/', '\\', '_' characters")
	}

	if viper.GetBool("rotator.has-azimuth") {

		if viper.GetInt("rotator.azimuth-min") >= viper.GetInt("rotator.azimuth-max") {
			return fmt.Errorf("azimuth-min must be smaller than azimuth-max")
		}

		if viper.GetInt("rotator.azimuth-max") > 360 && viper.GetInt("rotator.azimuth-min") > 360 {
			return fmt.Errorf("if azimuth-max is >360, azimuth-min must be < 360")
		}

		if viper.GetInt("rotator.azimuth-min") < 0 {
			return fmt.Errorf("azimuth-min must be >= 0")
		}

		if viper.GetInt("rotator.azimuth-max") > 500 {
			return fmt.Errorf("azimuth-max must be <= 500")
		}
	}

	if viper.GetBool("rotator.has-elevation") {

		if viper.GetInt("rotator.elevation-min") < 0 {
			return fmt.Errorf("elevation-min must be >= 0")
		}

		if viper.GetInt("rotator.elevation-max") > 180 {
			return fmt.Errorf("elevation-min must be <= 180")
		}

		if viper.GetInt("rotator.elevation-min") >= viper.GetInt("rotator.elevation-max") {
			return fmt.Errorf("elevation-min must be smaller than elevation-max")
		}
	}

	return nil
}

func sanityCheckDiscovery() error {

	if viper.GetBool("discovery.enabled") && !viper.GetBool("http.enabled") {
		return fmt.Errorf("for discovery, HTTP must be enabled")
	}

	return nil
}
