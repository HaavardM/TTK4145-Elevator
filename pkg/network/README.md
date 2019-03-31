# Network Module

The network module is seperated into three different modules

## AtMostOnce
AtMostOnce is a "fire-and-forget" netork service which sends datastructures over the network with a sender id. It does not gurantee that the message is delivered in any form. It should be used for data that is frequently published.

## Heartbeat
The heartbeat module detects other elevators on the network. It also sends the order cost for an elevator as part of the heartbeat. The heartbeats are sent using the AtMostOnce module.

## AtLeastOnce
AtLeastOnce builds on the AtMostOnce module. Messages are sent using AtMostOnce with a message id, and is republished until acknowledgements are sent from all available nodes. When the module receives a message sent from another elevator, it automatically sends a new acknowledgement. More than one duplicate of a message might be received by each node. 
When a message is acknowlegded by all other elevators, the message is sent back to the sender as confirmation. The message structure looks like this:

| Field     | Datatype    | Value from                                                                                 |
|-----------|-------------|--------------------------------------------------------------------------------------------|
| Ack       | bool        | set to true by receiver                                                                    |
| MessageID | string      | From a combination of xid and a message counter - generates unique id for each new message |
| SenderID  | int         | Elevator id - either assigned during init or based on IP                                   |
| Data      | interface{} | Any serializable datatype                                                                  |

## External packages
|Package Name|Description|Reason|
|------------|-----------|------|
|[xid](https://github.com/rs/xid) | Generates globally unique IDs | Used to assign messages unique message ids |
|[Network-go](https://github.com/TTK4145/Network-go) | Network module | Used as inspiration as well as a few functions used to create connection and getting IP address. Broadcast module reimplemented to allow better resource management (contexts). Peer module (heartbeats) redesigned to allow adding costs. |
|[context](https://golang.org/x/net/context)|Goroutine context management (included in standard library from Golang 1.7)|To stop the goroutine if the context is no longer valid|


