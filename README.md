# Hub

* The global `Hub` has many `Topic`s.
* Every `Topic`has many subscribers and publisher, they are websocket connections.
* One websocat connection will receive or publish messages from/to other connections.
* `Hub` can publish `Message` to `Topic`.
* `Topic` will send `Message` to all subscribers matched the topic.
    * One subscriber maybe receive same messge for more than one times.
* Every subscriber will subscribe the `global` `Topic`.
