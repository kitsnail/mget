package main

import "fmt"

const (
	_          = iota
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

// bytesToHuman humen readable bumber
// num unit default Byte
// ret unit B,KB,MB,GB,TB,PB,EB,ZB,YB
// eg. num = 1024
// return 1,KB
func bytesToHuman(num float64) (ret float64, unit string) {
	switch {
	case num >= 0 && num < KB:
		ret = num
		unit = "B"
	case num >= KB && num < MB:
		ret = num / KB
		unit = "KiB"
	case num >= MB && num < GB:
		ret = num / MB
		unit = "MiB"
	case num >= GB && num < TB:
		ret = num / GB
		unit = "GiB"
	case num >= TB && num < PB:
		ret = num / TB
		unit = "TiB"
	case num >= PB && num < EB:
		ret = num / PB
		unit = "PiB"
	case num >= EB && num < ZB:
		ret = num / EB
		unit = "EiB"
	case num >= ZB && num < YB:
		ret = num / ZB
		unit = "ZiB"
	case num >= YB:
		ret = num / YB
		unit = "YiB"
	default:
		ret = num
		unit = "UNKOWN"
	}
	return
}

func BytesToHuman(num float64) (ret string) {
	rn, unit := bytesToHuman(num)
	return fmt.Sprintf("%.2f%s", rn, unit)
}
