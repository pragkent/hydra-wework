package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/golang/glog"
	"github.com/pragkent/hydra-wework/server"
)

var version bool

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		glog.Exitf("%v", err)
	}

	os.Exit(0)
}

func parseFlags() *server.Config {
	cfg := &server.Config{}
	var fs = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fs.StringVar(&cfg.BindAddr, "bind", ":6666", "bind address")
	fs.StringVar(&cfg.CookieSecret, "cookie-secret", "", "session cookie secret key")
	fs.StringVar(&cfg.HydraURL, "hydra-url", "", "hydra url")
	fs.StringVar(&cfg.HydraClientID, "hydra-client-id", "", "hydra client id")
	fs.StringVar(&cfg.HydraClientSecret, "hydra-client-secret", "", "hydra client secret")
	fs.StringVar(&cfg.WeworkCorpID, "wework-corp-id", "", "wework corp id")
	fs.StringVar(&cfg.WeworkAgentID, "wework-agent-id", "", "wework agent id")
	fs.StringVar(&cfg.WeworkSecret, "wework-secret", "", "wework secret")
	fs.BoolVar(&cfg.HTTPS, "https", true, "use https")

	version := fs.Bool("version", false, "version")
	verbosity := fs.Int("v", 0, "log verbvosity level")

	fs.Parse(os.Args[1:])

	if *version {
		fmt.Print(Version())
		return nil
	}

	initLogging(*verbosity)
	return cfg
}

func initLogging(verbosity int) {
	flag.Parse()

	// initlaizing glog
	flag.Set("logtostderr", "true")
	flag.Set("v", strconv.Itoa(verbosity))
}

func run(cfg *server.Config) error {
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
