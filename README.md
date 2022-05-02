# Pollution Printouts

Generates a printable poster / leaflet for a given address or set of addresses based on
data from [addresspollution.org](https://addresspollution.org)

Work in progress.


## Example

```

func main() {
	ap, err := addresspollution.NewClient(
		"",  // INSERT YOUR EMAIL HERE
		"",  // LINK TO YOUR PROJECT HERE (or empty string if it doesn't have a URL)
	)

	if err != nil {
		log.Panic(err)
	}

	addresses, err := ap.Addresses("YO24 4JF")
	if err != nil {
		log.Panic(err)
	}

	for _, addr := range addresses {
		rating, _ := ap.PollutionAtAddress(addr.ID)

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

```
