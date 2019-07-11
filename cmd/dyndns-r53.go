package main

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/kardianos/osext"

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

	// 2. Read config file
	conf := config.GetConfig(logger)
	if conf.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	// Get the current directory
	dir, err := osext.ExecutableFolder()
	perror(err, log)

	// Initialze a log file
	logFile, err := os.OpenFile(path.Join(dir, "dyndns.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	loggingBackend := logging.NewLogBackend(logFile, "", 0)
	backendFormatter := logging.NewBackendFormatter(loggingBackend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(logging.INFO, "")
	logging.SetBackend(backendLeveled)
	perror(err, log)

	// Read the config file
	configFile, err := os.Open(path.Join(dir, "config.json"))
	perror(err, log)
	decoder := json.NewDecoder(configFile)
	var config Configuration
	err = decoder.Decode(&config)
	perror(err, log)

	// Request your WAN ip
	url := "https://api.ipify.org?format=json"
	timeout := time.Duration(30 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	res, err := client.Get(url)
	perror(err, log)
	defer res.Body.Close()
	decoder = json.NewDecoder(res.Body)
	var body Response
	err = decoder.Decode(&body)
	perror(err, log)
	wanIp := body.Ip
	log.Debugf("Current WAN ip: %s", wanIp)

	// Obtain the current ip bounded to the FQDN
	ips, err := net.LookupHost(config.Fqdn)
	currentIp := ips[0]
	log.Debugf("Current ip bounded to '%s': %s", config.Fqdn, currentIp)

	// Update the FQDN's ip in case the current WAN ip is different from the ip bounded to the FQDN
	if currentIp != wanIp {

		log.Infof("'%s' out-of-date update '%s' to '%s'", config.Fqdn, currentIp, wanIp)

		var token string
		creds := credentials.NewStaticCredentials(config.AwsAccessKeyId, config.AwsSecretAccessKey, token)

		svc := route53.New(session.New(), &aws.Config{
			Credentials: creds,
		})

		params := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &route53.ChangeBatch{ // Required
				Changes: []*route53.Change{ // Required
					{ // Required
						Action: aws.String("UPSERT"), // Required
						ResourceRecordSet: &route53.ResourceRecordSet{ // Required
							Name: aws.String("synology.pixxis.be"), // Required
							Type: aws.String("A"),                  // Required
							ResourceRecords: []*route53.ResourceRecord{
								{ // Required
									Value: aws.String(body.Ip), // Required
								},
							},
							TTL: aws.Int64(111),
						},
					},
				},
				Comment: aws.String("IP update by GO script"),
			},
			HostedZoneId: aws.String(config.HostedZoneId), // Required
		}
		resp, err := svc.ChangeResourceRecordSets(params)
		perror(err, log)

		// Pretty-print the response data.
		log.Debugf("Route53 response: %v", resp)

	} else {
		log.Infof("'%s' is up-to-date", config.Fqdn)
	}
}
