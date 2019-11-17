package component

type RouterTableEntry struct {
	Name    string
	NetDest IP
	Nexthop IP
	Port    uint8
}
