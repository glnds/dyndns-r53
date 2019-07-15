[![Go Report Card](https://goreportcard.com/badge/github.com/glnds/dyndns-r53)](https://goreportcard.com/report/github.com/glnds/dyndns-r53)

# DynDNS Route53
This is a little program written in [Go](https://golang.org/project/) that 
takes the WAN ip of your current infrastructure to update a hostname hosted
on [Amazon Route53](https://aws.amazon.com/route53/).

## Installation

### Make a binary
One of the main reasons to choose Go to develop this program was the fact
that Go can build executable binaries, which makes installation very easy. Secondly 
with Go it's also very easy to cross compile the program for different platforms.

To make a binary just run:
```
go build dyndns_route53.go
```
This will result in an executable binary named `dyndns_route53`.

#### Cross compile
This real power however is that you can cross compile the code for a number of 
platforms.
Personally I have this program running on my (Synology) NAS. 
This NAS is running Linux and has a x86-64 architecture. In order to
get the program running on that device you need to cross compile it for that specific platform.
To achieve this, you have to specify two extra parameters on the build command: 
`GOOS` and `GOARCH`, you can find the appropriate values for 
these variables [here](https://golang.org/doc/install/source#environment).

Here's how I build an executable binary for my NAS:
```
# env GOOS=linux GOARCH=amd64 go build dyndns_route53.go
```

### Configuration
Put a file named `config.json` in the same directory as you executable. A file   named `config.example` is included in the repository to make your life easy. It looks like this:
```
{
  "aws_access_key_id":    "ABC...",
  "aws_secret_access_key: "XYZ...",
  "hosted_zone_id":       "A1..",
  "fqdn":                 "www.example.com"
}
```

### Logging
The program will write its output to a file named `dyndns.log` under the same directory as you executable.

## Usage

Test and run locally:
```
go build dyndns_route53.go
go run dyndns_route53.go
```

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D

## License

MIT: http://rem.mit-license.org
