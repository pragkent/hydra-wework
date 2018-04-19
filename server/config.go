package server

import "errors"

type Config struct {
	BindAddr          string
	CookieSecret      string
	HydraURL          string
	HydraClientID     string
	HydraClientSecret string
	WeworkCorpID      string
	WeworkAgentID     string
	WeworkSecret      string
	HTTPS             bool
}

func (c *Config) Validate() error {
	if c.CookieSecret == "" {
		return errors.New("cookie secret is missing")
	}

	if c.HydraURL == "" {
		return errors.New("hydra url is missing")
	}

	if c.HydraClientID == "" {
		return errors.New("hydra client id is missing")
	}

	if c.HydraClientSecret == "" {
		return errors.New("hydra client secret is missing")
	}

	if c.WeworkCorpID == "" {
		return errors.New("wework corp id is missing")
	}

	if c.WeworkAgentID == "" {
		return errors.New("wework agent id is missing")
	}

	if c.WeworkSecret == "" {
		return errors.New("wework secret is missing")
	}

	return nil
}
