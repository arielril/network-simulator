package component

type Node struct {
	Name     string
	Ip       IP
	Gateway  IP
	ArpTable interface{}
}

type INode interface {
	SendArp()
	ReceiveArp()
	SendIcmp()
	ReceiveIcmp()
}

func (n Node) SendArp()     {}
func (n Node) ReceiveArp()  {}
func (n Node) SendIcmp()    {}
func (n Node) ReceiveIcmp() {}
