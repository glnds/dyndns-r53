package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/user"

	"github.com/glnds/dyndns-r53/internal/app/dyndns"
)

var version, build string

// CLIFlags represents the command line flags
type CLIFlags struct {
	Version bool
}

func main() {

	usr, err := user.Current()
	if err != nil {
		slog.Error(err.Error())
	}

	// Create the logger file if doesn't exist. Append to it if it already exists.
	var filename = "dyndns-r53.log"
	file, err := os.OpenFile(usr.HomeDir+string(os.PathSeparator)+".dyndns"+string(os.PathSeparator)+filename,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)

	if err != nil {
		slog.Error("Failed to log to file, using default stderr")
	}
	defer file.Close()

	var programLevel = new(slog.LevelVar) // Info by default
	h := slog.NewTextHandler(file, &slog.HandlerOptions{Level: programLevel})
	logger := slog.New(h)

	logger.Info("------------------ Starting IP update...  ------------------")

	// Read the command line flags
	flags := parseFlags()
	logger.Info("Parsed the commadline flags", "flags", flags)

	//  Read config file
	conf := dyndns.GetConfig(logger)
	if conf.Debug {
		programLevel.Set(slog.LevelDebug) // Update log level to Debug
	}

	wanIP := dyndns.GetWanIP(logger)

	// Update the FQDN's IP in case the current WAN ip is different from the IP bounded to the FQDN
	if dyndns.GetFqdnIP(conf, logger) != wanIP {
		logger.Info("out-of-date update to", conf.Fqdn, wanIP)
		dyndns.UpdateFqdnIP(conf, logger, wanIP)
	} else {
		logger.Info("'%s' is up-to-date", conf.Fqdn)
	}
}

func parseFlags() CLIFlags {
	flags := new(CLIFlags)

	flag.BoolVar(&flags.Version, "version", false, "prints dyndns-r53 version")
	flag.Parse()
	if flags.Version {
		fmt.Printf("dyndsn-r53 version: %s, build: %s\n", version, build)
		os.Exit(0)
	}
	return *flags
}
