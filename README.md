# Web-2-Peer

**Web-2-Peer** is a decentralized platform to upload and consult static websites (for now) on the W2Pnet which is composed of every node that runs the client.

## Getting Starded

### Installation

In order to use W2P simply clone the repo and build it
```bash
go get github.com/yaanst/W2P
```

### Usage

W2P is run with the cli and you need to precise a few argument to the node
```bash
W2P -name="A" -addr=127.0.0.1:10000 -peers="127.0.0.1:10001"
```

- **name** is the name of the node your want to run
- **addr** is the IPv4 address on which your node will listen for other nodes
- **peers** is a list of already runing nodes which will help to enter the network
