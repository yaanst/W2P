# Web-2-Peer

**Web-2-Peer** is a decentralized platform to upload and consult static websites (for now) on the W2Pnet which is composed of every node that runs the client.

## Getting Starded

### Installation

Simply clone the repo and build it
```bash
go get github.com/yaanst/W2P
```

### Usage

W2P is run with the cli and you need to precise a few argument to the node, e.g:
```bash
W2P -name="NodeA" -addr="127.0.0.1:10000" -peers="127.0.0.1:10001" -uiPort=4000
```

- **name** is the name of the node your want to run
- **addr** is the IPv4 address on which your node will listen for other nodes
- **peers** is a list of already runing nodes which will help to enter the network
- **uiPort** is the port on which you can point your browser to access the UI
  (default is 8000)

If you wish to run several nodes locally, make sure to run them with different
ports in separate folders and have a copy of the _ui/webpage_ subfolder in each of these folders.

## Features

The idea for this project is to build a p2p network capable of serving distributed static websites. It includes the following functionnalities:
 - Browsing websites
 - Updating a website already created
 - Searching by keywrds
 - Integrity checks 
 - Browser based user interface
