package core

import (
	"strconv"
	"goSkylar/core/nmap"
)

func ScannerResultTransfer(resultStruct nmap.NmapResultStruct) string {
	return resultStruct.Ip + "§§§§" + strconv.Itoa(resultStruct.PortId) + "§§§§" + resultStruct.Protocol + "§§§§" + resultStruct.Service
}
