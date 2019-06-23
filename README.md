# Hub

* The global `Hub` has many `Topic`s.
* Every `Topic`has many subscribers and publisher, they are websocket connections.
* One websocket connection will receive/publish messages from/to other subscribers.
* `Hub` will publish `Message` to `Topic`.
* `Topic` will send `Message` to all subscribers under it.
    * One subscriber maybe receive same messge for more than one times.
* Every subscriber will subscribe the "global" `Topic`.
