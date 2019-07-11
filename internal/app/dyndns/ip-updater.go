package dyndns

import (
	// "encoding/json"
	"net"
	"net/http"
	"time"
	// "path"

	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/credentials"
	// "github.com/aws/aws-sdk-go/aws/session"
	// "github.com/aws/aws-sdk-go/service/route53"
	// "github.com/kardianos/osext"

	"github.com/sirupsen/logrus"
)

// GetWanIP Call to ipify.org to obtain the host's WAM IP address.
func GetWanIP(log *logrus.Logger) string {
	url := "https://api.ipify.org?format=json"
	timeout := time.Duration(30 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	if res, err := client.Get(url); err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	decoder = json.NewDecoder(res.Body)
	var body Response
	if _, err := decoder.Decode(&body); err != nil {
		log.Fatalln(err)
	}
	log.Debugf("Current WAN ip: %s", body.Ip)
	return body.Ip
}

// GetFqdnIP Get the FQDN's current IP address
func GetFqdnIP(conf Config, log *logrus.Logger) string {
	ips, err := net.LookupHost(config.Fqdn)
	log.Debugf("Current ip bounded to '%s': %s", config.Fqdn, ips[0])
	return ips[0]
}

// UpdateFqdnIP Update the FQDN with the current WAN IP address
func UpdateFqdnIP(conf Config, log *logrus.Logger) string {
	log.Infof("'%s' out-of-date update '%s' to '%s'", config.Fqdn, currentIp, wanIp)

	var token string
	creds := credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, token)

	svc := route53.New(session.New(), &aws.Config{
		Credentials: creds,
	})

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(conf.Fqdn),
						Type: aws.String("A"),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(body.Ip),
							},
						},
						TTL: aws.Int64(111),
					},
				},
			},
			Comment: aws.String("IP update by GO script"),
		},
		HostedZoneId: aws.String(config.HostedZoneId),
	}
	resp, err := svc.ChangeResourceRecordSets(params)
	perror(err, log)

	// Pretty-print the response data.
	log.Debugf("Route53 response: %v", resp)
}
