package scheduler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

//Turning an array of orders into json format before saving to a file. Saves to temporary file before
//overwriting to make sure no data is lost if an error occurs
func saveToOrdersFile(filePath string, currentOrders *schedOrders) error {
	tofile, err := json.Marshal(currentOrders)
	if err != nil {
		return err
	}
	//Save to temporary file
	err = ioutil.WriteFile(filePath+".tmp", tofile, 0644) //
	if err != nil {
		return err
	}
	//Rename file when write was successfull
	err = os.Rename(filePath+".tmp", filePath)
	if err != nil {
		return err
	}
	return nil
}

//Reads orders from a file and turn them back into an array of Order type from json format
func readFromOrdersFile(filePath string) (*schedOrders, error) {
	//Opens the json file and saves it to the variable jsonOrders
	jsonOrders, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer jsonOrders.Close()

	var orderlist schedOrders

	//If it was successfully opened, the content is read and put into the jsonContent variable
	jsonContent, err := ioutil.ReadAll(jsonOrders)
	if err != nil {
		return nil, err
	}

	//If successfully read from the file, the content of the jsonContent variable is unmarshalled and put back into original form in the orderList
	err = json.Unmarshal(jsonContent, &orderlist)
	if err != nil {
		return nil, err
	}
	return &orderlist, nil
}

//Checks if the file of orders exists
func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		log.Panicf("Unknown error: %s\n", err)
	}
	return false
}

//Deletes orders from the filepath
func deleteOrdersFile(filePath string) {
	os.Remove(filePath)
}
