package leaderelection

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type ElectorResponse struct {
	Name       string    `json:"name"`
	LastUpdate time.Time `json:"last_update"`
}

func IsLeader(log logrus.FieldLogger) bool {
	electorEndpoint, envIsSet := os.LookupEnv("ELECTOR_GET_URL")
	if !envIsSet {
		log.Errorf("env elector_get_url is not set")
		return false
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.WithError(err).Errorf("unable to get hostname")
		return false
	}

	// #nosec G107
	resp, err := http.Get(electorEndpoint)
	if err != nil {
		log.WithError(err).Errorf("unable to get leader")
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithError(err).Errorf("unexpected http status code")
		return false
	}

	var leader ElectorResponse
	if err := json.NewDecoder(resp.Body).Decode(&leader); err != nil {
		log.WithError(err).Errorf("unable to decode response")
		return false
	}

	return hostname == leader.Name
}
