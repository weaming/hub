# Hub

* The global `Hub` has many `Topic`s.
* Every `Topic`has many subscribers and publisher, they are websocket connections.
* One websocket connection will receive/publish messages from/to other subscribers.
* `Hub` will publish `Message` to `Topic`.
* `Topic` will send `Message` to all subscribers under it.
    * One subscriber maybe receive same messge for more than one times.
* Every subscriber will subscribe the "global" `Topic`.

## Core structures

```go
type PubMessage struct {
	Type         string        `json:"type"`
	Data         string        `json:"data"`          // string or base64 of bytes
	ExtendedData []string      `json:"extended_data"` // string or base64 of bytes, for sending multiple photos
	Captions     []string      `json:"captions"`      // all captions of Data and ExtendedData
}
```
