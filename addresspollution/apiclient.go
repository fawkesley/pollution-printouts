package addresspollution

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/time/rate"
)

// NewClient returns an APIClient.
func NewClient(contactEmail, projectURL string) (*APIClient, error) {
	if !strings.Contains(contactEmail, "@") {
		return nil, fmt.Errorf("contactEmail must be set to a valid address that can reach you")
	}

	uaParts := []string{
		"bot",
		contactEmail,
	}
	if projectURL != "" {
		uaParts = append(uaParts, projectURL)
	}
	userAgent := strings.Join(uaParts, " | ")

	return &APIClient{
		client:      http.Client{Timeout: 10 * time.Second},
		userAgent:   userAgent,
		rateLimiter: rate.NewLimiter(rate.Every(2*time.Second), 1), // 1 req every 2s
	}, nil
}

// APIClient provides API access to addresspollution.org's unofficial API
type APIClient struct {
	client      http.Client
	userAgent   string
	rateLimiter *rate.Limiter
}

func (c *APIClient) do(req *http.Request) (*http.Response, error) {
	ctx := context.Background()
	err := c.rateLimiter.Wait(ctx) // This is a blocking call. Honors the rate limit
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *APIClient) getWithJSONResponse(url string, decodeInto interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned HTTP %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&decodeInto)
	if err != nil {
		return err
	}

	return nil
}

// Addresses searches for a list of addresses based on a postcode.
func (c APIClient) Addresses(postcode string) ([]Address, error) {

	url := fmt.Sprintf("%s/addresses?postcode=%s", apiBaseURL, url.QueryEscape(postcode))
	response := AddressesResponse{}

	err := c.getWithJSONResponse(url, &response)
	return response.Data, err
}

// PollutionAtAddress returns the pollution level for a given address ID (as returned by
// the Addresses API call)
func (c APIClient) PollutionAtAddress(addressID uuid.UUID) (*PollutionLevels, error) {
	url := fmt.Sprintf("%s/addresses/%s", apiBaseURL, addressID)
	response := AddressResponse{}

	err := c.getWithJSONResponse(url, &response)
	if err != nil {
		return nil, err
	}

	p := &PollutionLevels{
		FormattedAddress:     response.Data.FormattedAddress,
		PollutionDescription: response.Data.AirPollution.Rating.LevelDesc,
	}

	err = parsePollutantsFromResponse(response, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// AddressesResponse represents a response from the addresses API call.
type AddressesResponse struct {
	Data []Address `json:"data"`
}

// Address represents a search result in the Addresses API call.
type Address struct {
	ID               uuid.UUID `json:"id"`
	FormattedAddress string    `json:"formatted_address"`
	Line1            string    `json:"line_1"`
	Line2            string    `json:"line_2"`
}

// AddressResponse represents a response from a single address pollution lookup
type AddressResponse struct {
	Data struct {
		ID       string `json:"id"`
		Postcode struct {
			Postcode     string `json:"postcode"`
			District     string `json:"district"`
			Constituency string `json:"constituency"`
			Mp           struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"mp"`
		} `json:"postcode"`
		AirPollution struct {
			Concentration string `json:"concentration"`
			Rating        struct {
				HealthCosts string `json:"healthCosts"`
				Level       int    `json:"level"`
				WhoLimit    string `json:"whoLimit"`
				LevelDesc   string `json:"levelDesc"`
			} `json:"rating"`
			Percentile string `json:"percentile"`
		} `json:"airPollution"`
		Solutions []struct {
			Type  string `json:"type"`
			Title string `json:"title"`
			Body  string `json:"body"`
		} `json:"solutions"`
		FormattedAddress string  `json:"formatted_address"`
		Line1            string  `json:"line1"`
		Line2            string  `json:"line2"`
		District         string  `json:"district"`
		City             string  `json:"city"`
		County           string  `json:"county"`
		Latitude         float64 `json:"latitude"`
		Longitude        float64 `json:"longitude"`
		DesktopImage     string  `json:"desktopImage"`
		MobileImage      string  `json:"mobileImage"`
	} `json:"data"`
}

// PollutionLevels contains the address, pollution description and 3 pollutant
// levels for a specific address.
type PollutionLevels struct {
	FormattedAddress     string // e.g. "48 Lindley Street, York"
	PollutionDescription string // e.g. "Significant", "Very high"
	No2                  float64
	Pm2_5                float64
	Pm10                 float64
}

// NumPollutantsExceedingLimits returns a number between 0 and 3 depending on
// how many of PM2.5, PM10 and NO2 exceed WHO limits.
func (p PollutionLevels) NumPollutantsExceedingLimits() uint {
	count := uint(0)

	if p.Pm2_5LimitMultiplier() > 1 {
		count++
	}

	if p.Pm10LimitMultiplier() > 1 {
		count++
	}

	if p.No2LimitMultiplier() > 1 {
		count++
	}

	return count
}

// Pm2_5LimitMultiplier returns how many times the PM2.5 level exceeds WHO limit
// e.g. 2.5 or 0.9 (if it's within the limit)
func (p PollutionLevels) Pm2_5LimitMultiplier() float64 {
	return p.Pm2_5 / pm2_5Limit
}

// Pm10LimitMultiplier returns how many times the PM2.5 level exceeds WHO limit
// e.g. 2.5 or 0.9 (if it's within the limit)
func (p PollutionLevels) Pm10LimitMultiplier() float64 {
	return p.Pm10 / pm10Limit
}

// No2LimitMultiplier returns how many times the PM2.5 level exceeds WHO limit
// e.g. 2.5 or 0.9 (if it's within the limit)
func (p PollutionLevels) No2LimitMultiplier() float64 {
	return p.No2 / no2Limit
}

// Pm2_5SafeLevelDescription returns a word that can be used in front of "safe level"
// e.g. "2x safe level" or "within safe level"
func (p PollutionLevels) Pm2_5SafeLevelDescription() string {
	x := p.Pm2_5LimitMultiplier()
	if x < 1.0 {
		return "within"
	}

	return fmt.Sprintf("%.1fx", x)
}

// Pm10SafeLevelDescription returns a word that can be used in front of "safe level"
// e.g. "2x safe level" or "within safe level"
func (p PollutionLevels) Pm10SafeLevelDescription() string {
	x := p.Pm10LimitMultiplier()
	if x < 1.0 {
		return "within"
	}

	return fmt.Sprintf("%.1fx", x)
}

// No2SafeLevelDescription returns a word that can be used in front of "safe level"
// e.g. "2x safe level" or "within safe level"
func (p PollutionLevels) No2SafeLevelDescription() string {
	x := p.No2LimitMultiplier()
	if x < 1.0 {
		return "within"
	}

	return fmt.Sprintf("%.1fx", x)
}

const (
	apiBaseURL = "https://api.addresspollution.org/api/v2"
	no2Limit   = 10.0
	pm2_5Limit = 5
	pm10Limit  = 15
)
