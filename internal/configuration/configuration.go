package configuration

import (
	"flag"
	"log"
	"os"

	"github.com/TTK4145-students-2019/project-thefuturezebras/pkg/network"
)

//Config type contains system configuration
type Config struct {
	ElevatorID   int
	BasePort     int
	ElevatorPort int
	Floors       int
	FolderPath   string
}

//GetConfig returs config based on default values and provided flags
func GetConfig() Config {
	conf := Config{}
	currentDir, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}
	flag.IntVar(&conf.ElevatorID, "id", -1, "Elevator ID")
	flag.IntVar(&conf.BasePort, "baseport", 2000, "Base network UDP port")
	flag.IntVar(&conf.ElevatorPort, "elevator-port", 15657, "Port for elevator server")
	flag.IntVar(&conf.Floors, "floors", 4, "Number of floors")
	flag.StringVar(&conf.FolderPath, "folder", currentDir, "Folder to store program files in")
	flag.Parse()

	if conf.ElevatorID < 0 {
		conf.ElevatorID, err = network.GetIDFromIP()
		if err != nil {
			log.Panicln(err)
		}
	}

	return conf
}
