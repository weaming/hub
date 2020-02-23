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
const GlobalTopicID = "global"

// types of internal messages
const (
	MTPlain    string = "PLAIN"
	MTMarkdown string = "MARKDOWN"
	MTJSON     string = "JSON"
	MTHTML     string = "HTML"
	MTImage    string = "IMAGE"
)

// types of websocket messages
const (
	MTFeedback string = "FEEDBACK" // used for async event feedback
	MTResponse string = "RESPONSE" // used for message response
	MTMessage  string = "MESSAGE"  // used for publish messages
)

type PubMessage struct {
	Type         string        `json:"type"`
	Data         string        `json:"data"`          // string or base64 of bytes
	ExtendedData []string      `json:"extended_data"` // string or base64 of bytes, for sending multiple photos
	Captions     []string      `json:"captions"`      // all captions of Data and ExtendedData
}
```
