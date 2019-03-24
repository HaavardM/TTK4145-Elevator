package scheduler
import (	"encoding/json"
			"io/ioutil"	
			"os"
			"fmt"	
			"github.com/TTK4145-students-2019/project-thefuturezebras/internal/configuration"	
		)

func readfromfile(conf Config) OrderList {
	path := conf.FilePath
	jsonOrders, err := os.Open(path + "/orderList.json")
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
