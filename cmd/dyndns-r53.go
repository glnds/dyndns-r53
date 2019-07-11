package main

import (
	// "encoding/json"
	// "net"
	// "net/http"
	"os"
	"os/user"
	// "path"

	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/credentials"
	// "github.com/aws/aws-sdk-go/aws/session"
	// "github.com/aws/aws-sdk-go/service/route53"
	// "github.com/kardianos/osext"
	"github.com/glnds/dyndns-r53/internal/app/dyndns"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

var version, build string

// type Response struct {
// 	Ip string
// }

func main() {

	usr, err := user.Current()
	if err != nil {
		logger.Fatal(err)
	}

	// Create the logger file if doesn't exist. Append to it if it already exists.
	var filename = ".dyndns-r53.log"
	file, err := os.OpenFile(usr.HomeDir+string(os.PathSeparator)+filename,
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
