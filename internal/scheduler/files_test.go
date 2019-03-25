package scheduler

import (
	"reflect"
	"testing"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
)

const FilePath string = "/home/haavard"

func TestSaveRead(t *testing.T) {
	orders := schedOrders{
		OrdersCab: []*SchedulableOrder{
			&SchedulableOrder{
				Order: common.Order{
					Floor: 0,
					Dir:   common.NoDir,
				},
			},
			nil,
			nil,
			nil,
		},
		OrdersUp: []*SchedulableOrder{
			nil,
			&SchedulableOrder{
				Order: common.Order{
					Floor: 1,
					Dir:   common.UpDir,
				},
			},
			nil,
		},
		OrdersDown: []*SchedulableOrder{
			&SchedulableOrder{
				Order: common.Order{
					Floor: 1,
					Dir:   common.DownDir,
				},
			},
			nil,
			nil,
		},
	}
	//go scheduler.savetofile(FilePath, orders)
	//go scheduler.readfromfile(FilePath)
	err := saveToOrdersFile(FilePath, &orders)
	if err != nil {
		t.Error(err)
	}
	x, err := readFromOrdersFile(FilePath)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*x, orders) {
		t.Error("Read not equal to write")
	}
}
