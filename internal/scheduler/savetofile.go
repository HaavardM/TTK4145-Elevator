package scheduler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/configuration"
)

type OrderList struct {
	Order []OrdersType `json:"orders"`
}

type OrdersType struct {
	Dir   int `json:"direction"`
	Floor int `json:"floor"`
}

//need to configure Filepath!!

//turning an array of orders into json format before saving to a file. Saves to temporary file before
//overwriting to make sure no data is lost if an error occurs.
func savetofile(folderPath string, currentOrders OrderList) error {
	conf := configuration.GetConfig()
	tofile, err := json.Marshal(currentOrders)
	if err != nil {
		return err
	}
	fmt.Println("marshal performed successfully")
	err = ioutil.WriteFile(folderPath+"/orders.json.tmp", tofile, 0644) //
	if err != nil {
		return err
	}
	fmt.Println("written to file")
	err = os.Rename(folderPath+"/orders.json.tmp", folderPath+"/orders.json")
	if err != nil {
		return err
	}
	return nil
}
