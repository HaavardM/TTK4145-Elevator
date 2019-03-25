package scheduler

import (
	"fmt"
	//"log"
	//"time"
	"encoding/json"
	"io/ioutil"
	"os"
)

const FilePath string = "C:\\Users\\Julie\\Documents\\test_skrivtilfil"

type OrderList struct {
	Order []OrderType
}

type OrderType struct {
	Dir   int
	Floor int
}

func main() {
	orders := OrderList{
		Order: []OrderType{
			OrderType{
				Dir:   1,
				Floor: 3,
			},
			OrderType{
				Dir:   0,
				Floor: 1,
			},
			OrderType{
				Dir:   2,
				Floor: 3,
			},
		},
	}
	//go scheduler.savetofile(FilePath, orders)
	//go scheduler.readfromfile(FilePath)
	savetofile(FilePath, orders)
	x := readfromfile(FilePath)
	fmt.Println(x)

}

func savetofile(FilePath string, currentOrders OrderList) {
	//path := conf.FilePath
	tofile, err := json.Marshal(currentOrders)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("marshal performed successfully")
	err = ioutil.WriteFile(FilePath+"/temporaryOrders.json", tofile, 0644) //
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("written to file")
	err = os.Rename(FilePath+"/temporaryOrders.json", FilePath+"/orderList.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("saved to new file")
}

func readfromfile(FilePath string) OrderList {
	//path := conf.FilePath
	jsonOrders, err := os.Open(FilePath + "/orderList.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Opened json-file")

	defer jsonOrders.Close()

	var orderlist OrderList

	jsonContent, err := ioutil.ReadAll(jsonOrders)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Read from json-file")

	err = json.Unmarshal(jsonContent, &orderlist)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Unmarshal has been performed successfully")
	fmt.Println(orderlist)

	return orderlist
}
