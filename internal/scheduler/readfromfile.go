package scheduler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func readfromfile(folderPath string) (*OrderList, error) {
	jsonOrders, err := os.Open(folderPath + "/orderList.json")
	if err != nil {
		return nil, err
	}
	defer jsonOrders.Close()

	var orderlist OrderList

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
