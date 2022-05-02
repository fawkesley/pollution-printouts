package addresspollution

import (
	"fmt"
	"regexp"
	"strconv"
)

func parsePollutantsFromResponse(resp AddressResponse, pollution *PollutionLevels) error {
	var err error

	if pollution.No2, err = parseNo2(resp); err != nil {
		return err
	}

	if pollution.Pm2_5, err = parsePm2_5(resp); err != nil {
		return err
	}

	if pollution.Pm10, err = parsePm10(resp); err != nil {
		return err
	}

	return nil
}

func parseNo2(resp AddressResponse) (float64, error) {
	strconv.ParseFloat(resp.Data.AirPollution.Concentration, 64)

	x, err := strconv.ParseFloat(resp.Data.AirPollution.Concentration, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse NO2 `%s`: %v", resp.Data.AirPollution.Concentration, err)
	}

	return x, nil
}

func parsePm2_5(resp AddressResponse) (float64, error) {
	match := pm2_5Pattern.FindStringSubmatch(resp.Data.AirPollution.Rating.HealthCosts)

	if len(match) == 0 {
		return 0, fmt.Errorf("couldn't parse PM2.5 reading")
	}
	x, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, err
	}

	return x, nil
}

func parsePm10(resp AddressResponse) (float64, error) {
	match := pm10Pattern.FindStringSubmatch(resp.Data.AirPollution.Rating.HealthCosts)

	if len(match) == 0 {
		return 0, fmt.Errorf("couldn't parse PM10 reading")
	}
	x, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, err
	}

	return x, nil
}

var (
	pm2_5Pattern = regexp.MustCompile(`annual average of the pollutant PM2.5 is (\d+\.\d+)mcg/m3`)
	pm10Pattern  = regexp.MustCompile(`reading for PM10 at this address is (\d+\.\d+)mcg/m3`)
)
