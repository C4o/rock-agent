package collect

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func str2float(s string) float64 {

	i, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1
	}
	return i / 100
}

func getCpu() map[string]float64 {

	var data string
	var total float64
	var scanner *bufio.Scanner
	file, _ := os.OpenFile("/proc/stat", os.O_RDONLY, 0444)
	defer file.Close()
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		data = scanner.Text()
		break
	}
	Arr := strings.Split(data, " ")
	user := str2float(Arr[2])
	system := str2float(Arr[4])
	idle := str2float(Arr[5])
	for k, v := range Arr {
		if k > 1 {
			total = total + str2float(v)
		}
	}
	cpuinfo := make(map[string]float64)
	cpuinfo = map[string]float64{
		"user":   round(user/total, 4),
		"system": round(system/total, 4),
		"idle":   round(idle/total, 4),
		"total":  round(total, 4),
	}
	return cpuinfo

}
