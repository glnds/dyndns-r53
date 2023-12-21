package dyndns

import (
	"log"
	"log/slog"
	"os"
	"os/user"

	"github.com/BurntSushi/toml"
)

// Config represents the application's config file
type Config struct {
	AccessKeyID     string `toml:"AccessKeyID"`
	SecretAccessKey string `toml:"SecretAccessKey"`
	HostedZoneID    string `toml:"HostedZoneID"`
	Fqdn            string `toml:"Fqdn"`
	Debug           bool   `toml:"Debug"`
}

// GetConfig reads the .dyndns-r53.toml configuration file for initialization.
func GetConfig(logger *slog.Logger) Config {

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Read dyndns-r53.toml config file for initialization
	conf := Config{Debug: false} // Set default values
	if _, err := toml.DecodeFile(usr.HomeDir+string(os.PathSeparator)+".dyndns"+string(os.PathSeparator)+"config.toml", &conf); err != nil {
		log.Fatal(err.Error())
	}

	logger.Info("Config values", "AccessKeyID", conf.AccessKeyID, "HostedZoneID",
		conf.HostedZoneID, "Fqdn", conf.Fqdn, "Debug", conf.Debug)

	return conf
}
