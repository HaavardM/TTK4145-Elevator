package network

import (
	"strconv"
	"strings"

	"github.com/TTK4145/Network-go/network/localip"
)

//GetIDFromIP returns an ID based on the computer IP
func GetIDFromIP() (int, error) {
	ip, err := localip.LocalIP()
	if err != nil {
		return 0, err
	}

	strRes := strings.Join(strings.Split(ip, "."), "")
	res, err := strconv.Atoi(strRes)
	if err != nil {
		return 0, err
	}

	return res, nil
}
