# Network Simulator
This is a Network Simulator with nodes, routers and router tables

## Assignment Description
O trabalho consiste em desenvolver um simulador de rede. O simulador deve receber como parâmetros de execução o nome de um arquivo de descrição de topologia (conforme formato especificado), um nó origem, um nó destino e uma mensagem. O simulador deve apresentar na saída as mensagens enviadas pelos nós e roteadores da topologia conforme o formato estabelecido, considerando o envio de um ping (ICMP Echo Request) do nó origem até o nó destino contendo a mensagem indicada por parâmetro. O simulador deverá realizar a transmissão da mensagem através do ping respeitando a topologia da rede e necessidade de fragmentação da mensagem conforme o MTU das interfaces de rede. O simulador considera o MTU somente para fragmentar o campo de dados do pacote ICMP (cabeçalhos não são considerados no valor do MTU).

### Topology Description File

```s
#NODE
<node_name>,<MAC>,<IP/prefix>,<MTU>,<gateway>
#ROUTER
<router_name>,<num_ports>,<MAC0>,<IP0/prefix>,<MTU0>,<MAC1>,<IP1/prefix>,<MTU1>,<MAC2>,<IP2/prefix>,<MTU2> …
#ROUTERTABLE
<router_name>,<net_dest/prefix>,<nexthop>,<port>
```

### Output Example

```s
Pacotes ARP Request: <src_name> box <src_name> : ETH (src=<MAC_src> dst =<MAC_dst>) \n ARP - Who has <IP_dst>? Tell <IP_src>;
Pacotes ARP Reply: <src_name> => <dst_name> : ETH (src=<MAC_src> dst =<MAC_dst>) \n ARP - <src_IP> is at <src_MAC>;
Pacotes ICMP Echo Request: <src_name> => <dst_name> : ETH (src=<MAC_src> dst =<MAC_dst>) \n IP (src=<IP_src> dst=<IP_dst> ttl=<TTL> mf=<mf_flag> off=<offset>) \n ICMP - Echo request (data=<msg>);
Pacotes ICMP Echo Reply: <src_name> => <dst_name> : ETH (src=<MAC_src> dst =<MAC_dst>) \n IP (src=<IP_src> dst=<IP_dst> ttl=<TTL> mf=<mf_flag> off=<offset>) \n ICMP - Echo reply (data=<msg>);
Pacotes ICMP Time Exceeded: <src_name> => <dst_name> : ETH (src=<MAC_src> dst =<MAC_dst>) \n IP (src=<IP_src> dst=<IP_dst> ttl=<TTL>) \n ICMP - Time Exceeded
Processamento final do ICMP Echo Request/Reply no nó: <dst_name> rbox <dst_name> : Received <msg>;
```

### Execution command
```s
$ simulador <topologia> <origem> <destino> <mensagem>
```

### Examples

> Arquivo topologia.txt

```
#NODE
n1,00:00:00:00:00:01,192.168.0.2/24,5,192.168.0.1
n2,00:00:00:00:00:02,192.168.0.3/24,5,192.168.0.1
n3,00:00:00:00:00:03,192.168.1.2/24,5,192.168.1.1
n4,00:00:00:00:00:04,192.168.1.3/24,5,192.168.1.1
#ROUTER
r1,2,00:00:00:00:00:05,192.168.0.1/24,5,00:00:00:00:00:06,192.168.1.1/24,5
#ROUTERTABLE
r1,192.168.0.0/24,0.0.0.0,0
r1,192.168.1.0/24,0.0.0.0,1
```

> Exemplos de execução:

```s
$ simulador topologia.txt n1 n2 hello
n1 box n1 : ETH (src=00:00:00:00:00:01 dst=FF:FF:FF:FF:FF:FF) \n ARP - Who has 192.168.0.3? Tell 192.168.0.2;
n2 => n1 : ETH (src=00:00:00:00:00:02 dst=00:00:00:00:00:01) \n ARP - 192.168.0.3 is at 00:00:00:00:00:02;
n1 => n2 : ETH (src=00:00:00:00:00:01 dst=00:00:00:00:00:02) \n IP (src=192.168.0.2 dst=192.168.0.3 ttl=8 mf=0 off=0) \n ICMP - Echo request (data=hello);
n2 rbox n2 : Received hello;
n2 => n1 : ETH (src=00:00:00:00:00:02 dst=00:00:00:00:00:01) \n IP (src=192.168.0.3 dst=192.168.0.2 ttl=8 mf=0 off=0) \n ICMP - Echo reply (data=hello);
n1 rbox n1 : Received hello;
```

```s
$ simulador topologia.txt n1 n2 helloworld
n1 box n1 : ETH (src=00:00:00:00:00:01 dst=FF:FF:FF:FF:FF:FF) \n ARP - Who has 192.168.0.3? Tell 192.168.0.2;
n2 => n1 : ETH (src=00:00:00:00:00:02 dst=00:00:00:00:00:01) \n ARP - 192.168.0.3 is at 00:00:00:00:00:02;
n1 => n2 : ETH (src=00:00:00:00:00:01 dst=00:00:00:00:00:02) \n IP (src=192.168.0.2 dst=192.168.0.3 ttl=8 mf=1 off=0) \n ICMP - Echo request (data=hello);
n1 => n2 : ETH (src=00:00:00:00:00:01 dst=00:00:00:00:00:02) \n IP (src=192.168.0.2 dst=192.168.0.3 ttl=8 mf=0 off=5) \n ICMP - Echo request (data=world);
n2 rbox n2 : Received helloworld;
n2 => n1 : ETH (src=00:00:00:00:00:02 dst=00:00:00:00:00:01) \n IP (src=192.168.0.3 dst=192.168.0.2 ttl=8 mf=1 off=0) \n ICMP - Echo reply (data=hello);
n2 => n1 : ETH (src=00:00:00:00:00:02 dst=00:00:00:00:00:01) \n IP (src=192.168.0.3 dst=192.168.0.2 ttl=8 mf=0 off=5) \n ICMP - Echo reply (data=world);
n1 rbox n1 : Received helloworld;
$ simulador topologia.txt n1 n3 hello
n1 box n1 : ETH (src=00:00:00:00:00:01 dst=FF:FF:FF:FF:FF:FF) \n ARP - Who has 192.168.0.1? Tell 192.168.0.2;
r1 => n1 : ETH (src=00:00:00:00:00:05 dst=00:00:00:00:00:01) \n ARP - 192.168.0.1 is at 00:00:00:00:00:05;
n1 => r1 : ETH (src=00:00:00:00:00:01 dst=00:00:00:00:00:05) \n IP (src=192.168.0.2 dst=192.168.1.2 ttl=8 mf=0 off=0) \n ICMP - Echo request (data=hello);
r1 box r1 : ETH (src=00:00:00:00:00:06 dst=FF:FF:FF:FF:FF:FF) \n ARP - Who has 192.168.1.2? Tell 192.168.1.1;
n3 => r1 : ETH (src=00:00:00:00:00:03 dst=00:00:00:00:00:06) \n ARP - 192.168.1.2 is at 00:00:00:00:00:03;
r1 => n3 : ETH (src=00:00:00:00:00:06 dst=00:00:00:00:00:03) \n IP (src=192.168.0.2 dst=192.168.1.2 ttl=7 mf=0 off=0) \n ICMP - Echo request (data=hello);
n3 rbox n3 : Received hello;
n3 => r1 : ETH (src=00:00:00:00:00:03 dst=00:00:00:00:00:06) \n IP (src=192.168.1.2 dst=192.168.0.2 ttl=8 mf=0 off=0) \n ICMP - Echo reply (data=hello);
r1 => n1 : ETH (src=00:00:00:00:00:05 dst=00:00:00:00:00:01) \n IP (src=192.168.1.2 dst=192.168.0.2 ttl=7 mf=0 off=0) \n ICMP - Echo reply (data=hello);
n1 rbox n1 : Received hello;
```

```s
$ simulador topologia.txt n1 n3 helloworld
n1 box n1 : ETH (src=00:00:00:00:00:01 dst=FF:FF:FF:FF:FF:FF) \n ARP - Who has 192.168.0.1? Tell 192.168.0.2;
r1 => n1 : ETH (src=00:00:00:00:00:05 dst=00:00:00:00:00:01) \n ARP - 192.168.0.1 is at 00:00:00:00:00:05;
n1 => r1 : ETH (src=00:00:00:00:00:01 dst=00:00:00:00:00:05) \n IP (src=192.168.0.2 dst=192.168.1.2 ttl=8 mf=1 off=0) \n ICMP - Echo request (data=hello);
n1 => r1 : ETH (src=00:00:00:00:00:01 dst=00:00:00:00:00:05) \n IP (src=192.168.0.2 dst=192.168.1.2 ttl=8 mf=0 off=5) \n ICMP - Echo request (data=world);
r1 box r1 : ETH (src=00:00:00:00:00:06 dst=FF:FF:FF:FF:FF:FF) \n ARP - Who has 192.168.1.2? Tell 192.168.1.1;
n3 => r1 : ETH (src=00:00:00:00:00:03 dst=00:00:00:00:00:06) \n ARP - 192.168.1.2 is at 00:00:00:00:00:03;
r1 => n3 : ETH (src=00:00:00:00:00:06 dst=00:00:00:00:00:03) \n IP (src=192.168.0.2 dst=192.168.1.2 ttl=7 mf=1 off=0) \n ICMP - Echo request (data=hello);
r1 => n3 : ETH (src=00:00:00:00:00:06 dst=00:00:00:00:00:03) \n IP (src=192.168.0.2 dst=192.168.1.2 ttl=7 mf=0 off=5) \n ICMP - Echo request (data=world);
n3 rbox n3 : Received helloworld;
n3 => r1 : ETH (src=00:00:00:00:00:03 dst=00:00:00:00:00:06) \n IP (src=192.168.1.2 dst=192.168.0.2 ttl=8 mf=1 off=0) \n ICMP - Echo reply (data=hello);
n3 => r1 : ETH (src=00:00:00:00:00:03 dst=00:00:00:00:00:06) \n IP (src=192.168.1.2 dst=192.168.0.2 ttl=8 mf=0 off=5) \n ICMP - Echo reply (data=world);
r1 => n1 : ETH (src=00:00:00:00:00:05 dst=00:00:00:00:00:01) \n IP (src=192.168.1.2 dst=192.168.0.2 ttl=7 mf=1 off=0) \n ICMP - Echo reply (data=hello);
r1 => n1 : ETH (src=00:00:00:00:00:05 dst=00:00:00:00:00:01) \n IP (src=192.168.1.2 dst=192.168.0.2 ttl=7 mf=0 off=5) \n ICMP - Echo reply (data=world);
n1 rbox n1 : Received helloworld;
```
