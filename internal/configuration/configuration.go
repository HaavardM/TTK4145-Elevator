package configuration

import "flag"

//Config type contains system configuration
type Config struct {
	NetworkID int
	BasePort  int
}

//GetConfig returs config based on default values and provided flags
func GetConfig() Config {
	conf := Config{}
	flag.IntVar(&conf.NetworkID, "id", 1, "Elevator ID")
	flag.IntVar(&conf.BasePort, "baseport", 2000, "Base network UDP port")

	flag.Parse()

	return conf
}
