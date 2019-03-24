package scheduler

import (	"encoding/json"
			"fmt"
			"os"
			"io/ioutil"	
			"github.com/TTK4145-students-2019/project-thefuturezebras/internal/configuration"		
		)

type OrderList struct{
	Order []OrdersType
}

type OrdersType struct {
	Dir		int
	Floor	int
}

//need to configure Filepath!!


//turning an array of orders into json format before saving to a file. Saves to temporary file before
//overwriting to make sure no data is lost if an error occurs.
func savetofile(conf Config, currentOrders OrderList) {
	path := conf.FilePath
	tofile, err := json.Marshal(currentOrders)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("marshal performed successfully")
	err = ioutil.WriteFile(path + "/temporaryOrders.json", tofile, 0644)	//
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("written to file")
	err = os.Rename(path + "/temporaryOrders.json", path + "/orderList.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("saved to new file")
}