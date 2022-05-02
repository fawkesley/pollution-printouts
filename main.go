package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fawkesley/pollution-printouts/addresspollution"
	"github.com/fawkesley/pollution-printouts/leaflet"
)

func main() {

	email := os.Getenv("USER_AGENT_EMAIL")
	if email == "" {
		log.Panic("please set a contact email in USER_AGENT_EMAIL")
	}

	ap, err := addresspollution.NewClient(
		email,
		"https://github.com/fawkesley/pollution-printouts",
	)

	if err != nil {
		log.Panic(err)
	}

	postcodes := []string{
		"L15 0EB",
		"L15 2HD",
		"L15 2HF",
		"L15 2LQ",
		"L15 3JA",
		"L15 3JJ",
		"L15 3JL",
		"L15 3JR",
		"L15 3JT",
		"L15 5AE",
		"L15 5AF",
		"L15 5AG",
		"L15 5AH",
		"L15 5AJ",
		"L15 5AN",
	}

	var addresses []addresspollution.Address

	for _, pc := range postcodes {
		fmt.Println(pc)
		a, err := ap.Addresses(pc)
		if err != nil {
			log.Panic(err)
		}

		addresses = append(addresses, a...)
	}

	for i, addr := range addresses {
		rating, err := ap.PollutionAtAddress(addr.ID)
		if err != nil {
			log.Panic(err)
		}

		fn := fmt.Sprintf("output/%03d-%s.png", i, slugify(rating.FormattedAddress))
		fmt.Printf("writing %s\n", fn)

		f, err := os.Create(fn)
		if err != nil {
			log.Panic(err)
		}
		defer f.Close()

		err = leaflet.RenderPNG(*rating, f)
		if err != nil {
			log.Panic(err)
		}

		fmt.Printf("⚠️ %s AIR POLLUTION ⚠️\n", strings.ToUpper(rating.PollutionDescription))
		fmt.Printf("%s\n\n", rating.FormattedAddress)
		fmt.Printf("At your home, %d out of 3 pollutants\n", rating.NumPollutantsExceedingLimits())
		fmt.Printf("exceed World Health Organisation safe levels.\n\n")

		fmt.Printf("PM2.5\n")
		fmt.Printf("%.1f\n", rating.Pm2_5)
		fmt.Printf("%s safe level\n\n", rating.Pm2_5SafeLevelDescription())

		fmt.Printf("PM10\n")
		fmt.Printf("%.1f\n", rating.Pm10)
		fmt.Printf("%s safe level\n\n", rating.Pm10SafeLevelDescription())

		fmt.Printf("NO2\n")
		fmt.Printf("%.1f\n", rating.No2)
		fmt.Printf("%s safe level\n\n", rating.No2SafeLevelDescription())
	}

}

func slugify(address string) string {
	address = strings.ToLower(address)
	address = strings.Replace(address, " ", "-", -1)
	address = strings.Replace(address, ",", "", -1)
	return address

}
