package component

type Router struct {
	Name     string
	Ip       IP
	Gateway  IP
	ArpTable interface{}
	PortList []interface{}
}

type IRouter interface {
	SendArp()
	ReceiveArp()
	SendIcmp()
	ReceiveIcmp()
}

func (n Router) SendArp()     {}
func (n Router) ReceiveArp()  {}
func (n Router) SendIcmp()    {}
func (n Router) ReceiveIcmp() {}
