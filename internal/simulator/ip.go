package simulator

import (
	"fmt"
	"strconv"
	"strings"

	fp "github.com/novalagung/gubrak"
)

type IP struct {
	ip     string
	prefix uint8
}

func NewIp(ip string) *IP {
	ipSplit := strings.Split(ip, "/")
	var prefix uint8

	if len(ipSplit) == 1 {
		prefix = 0
	} else {
		pref, _ := strconv.ParseUint(ipSplit[1], 10, 8)
		prefix = uint8(pref)
	}

	netIp := &IP{
		ip:     ipSplit[0],
		prefix: prefix,
	}
	return netIp
}

func (ip IP) ToBit() uint32 {
	list, err := fp.Map(
		strings.Split(ip.ip, "."),
		func(part string) uint64 {
			val, _ := strconv.ParseUint(part, 10, 32)
			return val
		},
	)

	if err != nil {
		panic("Failed to convert IP")
	}
	ipList := list.([]uint64)
	bits := fmt.Sprintf(
		"%08b%08b%08b%08b",
		ipList[0],
		ipList[1],
		ipList[2],
		ipList[3],
	)
	res, _ := strconv.ParseUint(bits, 2, 32)
	return uint32(res)
}

func (ip IP) IsSameNet(ipDest IP) bool {
	ipSrcBit := ip.ToBit()
	ipDestBit := ipDest.ToBit()
	fst := MASK & (ipSrcBit & (MASK << (32 - ip.prefix)))
	snd := MASK & (fst | (MASK >> ip.prefix))

	return ipDestBit >= fst && ipDestBit <= snd
}

func (ip IP) ToString() string {
	return fmt.Sprintf(
		"%v/%v",
		ip.ip, ip.prefix,
	)
}
