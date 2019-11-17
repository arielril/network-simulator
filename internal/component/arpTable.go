package component

type ArpTable struct {
	table map[IP]string
}

type IArpTable interface {
	GetDestination(ip IP)
}

// The GetDestination function returns the MAC address of the destination.
// If the destination doesn't exists inside the ARP Table
// the return will be `""` (empty string). Otherwise,
// it will return the destination MAC address
func (arpTb ArpTable) GetDestination(ip IP) string {
	if arpTb.table == nil || len(arpTb.table) == 0 {
		return ""
	}

	dstMAC, ok := arpTb.table[ip]

	if ok {
		return dstMAC
	}

	return ""
}
