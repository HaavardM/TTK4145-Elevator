package scheduler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

//need to configure Filepath!!

//turning an array of orders into json format before saving to a file. Saves to temporary file before
//overwriting to make sure no data is lost if an error occurs.
func savetofile(folderPath string, currentOrders *schedOrders) error {
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

func readfromfile(folderPath string) (*schedOrders, error) {
	jsonOrders, err := os.Open(folderPath + "/orders.json")
	if err != nil {
		return nil, err
	}
	defer jsonOrders.Close()

	var orderlist schedOrders

	jsonContent, err := ioutil.ReadAll(jsonOrders)
	if err != nil {
		return nil, err
	}
	fmt.Println("Read from json-file")

	err = json.Unmarshal(jsonContent, &orderlist)
	if err != nil {
		return nil, err
	}
	fmt.Println("Unmarshal has been performed successfully")
	fmt.Println(orderlist)

	return &orderlist, nil
}

func deleteOrdersFile(folderPath string) {
	os.Remove(folderPath + "/orders.json")
}
