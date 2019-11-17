package component

type Node struct {
	Name     string
	NetInt   NetInterface
	Gateway  IP
	ArpTable ArpTable
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
