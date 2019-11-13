package component

type RouterTable struct {
	Name string
}
type Router struct {
	Name        string
	Ip          IP
	Gateway     IP
	ArpTable    struct{}
	PortList    []struct{}
	RouterTable RouterTable
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

func (r Router) SetRouterTable(rt RouterTable) Router {
	r.RouterTable = rt
	return r
}
