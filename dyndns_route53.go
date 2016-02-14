package main

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/kardianos/osext"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("dyndns")
var format = logging.MustStringFormatter(`%{time:2006-01-02T15:04:05.999999999} %{shortfunc} -  %{level:.5s} %{message}`)

type Configuration struct {
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	HostedZoneId       string
	Fqdn               string
}

type Response struct {
	Ip string
}

func perror(err error, logger *logging.Logger) {
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
}

func main() {
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
	res, err := http.Get(url)
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
		log.Debug(resp)

	} else {
		log.Infof("'%s' is up-to-date", config.Fqdn)
	}
}
