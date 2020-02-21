package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"

	"github.com/glnds/dyndns-r53/internal/app/dyndns"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

var version, build string

// CLIFlags represents the command line flags
type CLIFlags struct {
	Version bool
}

func main() {

	usr, err := user.Current()
	if err != nil {
		logger.Fatal(err)
	}

	// Create the logger file if doesn't exist. Append to it if it already exists.
	var filename = "dyndns-r53.log"
	file, err := os.OpenFile(usr.HomeDir+string(os.PathSeparator)+".dyndns"+string(os.PathSeparator)+filename,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	logger.Formatter = Formatter
	if err == nil {
		logger.Out = file
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}
	defer file.Close()

	logger.Info("------------------ Starting IP update...  ------------------")
	logger.SetLevel(logrus.InfoLevel)

	// Read the command line flags
	flags := parseFlags()
	logger.WithFields(logrus.Fields{
		"flags": flags,
	}).Info("Parsed the commandline flags")

	//  Read config file
	conf := dyndns.GetConfig(logger)
	if conf.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	wanIP := dyndns.GetWanIP(logger)

	// Update the FQDN's IP in case the current WAN ip is different from the IP bounded to the FQDN
	if dyndns.GetFqdnIP(conf, logger) != wanIP {
		logger.Infof("'%s' out-of-date update to '%s'", conf.Fqdn, wanIP)
		dyndns.UpdateFqdnIP(conf, logger, wanIP)
	} else {
		logger.Infof("'%s' is up-to-date", conf.Fqdn)
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
