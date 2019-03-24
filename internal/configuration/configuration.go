package configuration

import "flag"

//Config type contains system configuration
type Config struct {
	NetworkID    int
	BasePort     int
	ElevatorPort int
	Floors       int
	FilePath	string
}

//GetConfig returs config based on default values and provided flags
func GetConfig() Config {
	conf := Config{}
	flag.IntVar(&conf.NetworkID, "id", 1, "Elevator ID")
	flag.IntVar(&conf.BasePort, "baseport", 2000, "Base network UDP port")
	flag.IntVar(&conf.ElevatorPort, "elevator-port", 15657, "Port for elevator server")
	flag.IntVar(&conf.Floors, "floors", 4, "Number of floors")
	// configure path for saving files

	flag.Parse()

	return conf
}
