package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pragkent/hydra-wework/server"
)

var cfg = &server.Config{}
var version = flag.Bool("version", false, "version")

func init() {
	flag.StringVar(&cfg.BindAddr, "bind", ":6666", "bind address")
	flag.StringVar(&cfg.CookieSecret, "cookie-secret", "", "session cookie secret key")
	flag.StringVar(&cfg.HydraURL, "hydra-url", "", "hydra url")
	flag.StringVar(&cfg.HydraClientID, "hydra-client-id", "", "hydra client id")
	flag.StringVar(&cfg.HydraClientSecret, "hydra-client-secret", "", "hydra client secret")
	flag.StringVar(&cfg.WeworkCorpID, "wework-corp-id", "", "wework corp id")
	flag.StringVar(&cfg.WeworkAgentID, "wework-agent-id", "", "wework agent id")
	flag.StringVar(&cfg.WeworkSecret, "wework-secret", "", "wework secret")
	flag.BoolVar(&cfg.HTTPS, "https", true, "use https")
}

func main() {
	flag.Parse()

	if *version {
		fmt.Print(Version())
		return
	}

	if err := run(); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("Config validate error: %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		return err
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("ListenAndServe failed: %v", err)
	}

	return nil
}
