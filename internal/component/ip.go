package component

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/novalagung/gubrak"
)

const MASK uint32 = 0xffffffff

type IP struct {
	Ip     string
	Prefix uint8
}

type IPInterface interface {
	ToBit() uint32
	IsSameNet(ipDest IP) bool
}

func (ip IP) ToBit() uint32 {
	list, err := Map(
		strings.Split(ip.Ip, "."),
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
	fst := MASK & (ipSrcBit & (MASK << (32 - ip.Prefix)))
	snd := MASK & (fst | (MASK >> ip.Prefix))

	return ipDestBit >= fst && ipDestBit <= snd
}
