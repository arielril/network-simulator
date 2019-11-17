package component

type Port struct {
	Number uint8
	NetInterface
}

type Router struct {
	Name        string
	Ip          IP
	Gateway     IP
	ArpTable    struct{}
	PortList    []Port
	RouterTable map[IP]RouterTableEntry
}

type IRouter interface {
	SendArp()
	ReceiveArp()
	SendIcmp()
	ReceiveIcmp()
	SetRouterTable(rt struct{})
}

func (n Router) SendArp()     {}
func (n Router) ReceiveArp()  {}
func (n Router) SendIcmp()    {}
func (n Router) ReceiveIcmp() {}

func (r Router) SetRouterTable(rt map[IP]RouterTableEntry) Router {
	r.RouterTable = rt
	return r
}
