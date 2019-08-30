import logging
import json
from functools import partial
from enum import Enum
from base64 import b64encode
try:
    import thread
except ImportError:
    import _thread as thread
import websocket
import requests

log = logging.getLogger("hub")


class MESSAGE_TYPE(Enum):
    PLAIN = "PLAIN"
    MARKDOWN = "MARKDOWN"
    JSON = "JSON"
    HTML = "HTML"
    IMAGE = "IMAGE"


def on_message(ws, message):
    log.info(message)


def on_error(ws, error):
    log.error(error)


def on_close(ws):
    log.info("websocket closed")


def on_open(ws, bee=None):
    thread.start_new_thread(partial(bee, ws), ())


def run_until_close(bee=None, on_msg=None):
    # websocket.enableTrace(True)
    ws = websocket.WebSocketApp(
        "wss://hub.drink.cafe/ws",
        on_message=on_msg or on_message,
        on_error=on_error,
        on_close=on_close,
    )
    if bee:
        ws.on_open = partial(on_open, bee=bee)
    ws.run_forever(ping_interval=10)


def data_to_str(data, type):
    if type in [MESSAGE_TYPE.JSON.name]:
        return json.dumps(data, ensure_ascii=False)
    if type in [MESSAGE_TYPE.IMAGE.name]:
        if not isinstance(data, str):
            return b64encode(data).encode('utf8')
    return str(data)


def new_pub_message(data, type=MESSAGE_TYPE.PLAIN.name, topics=('global', )):
    return {
        'action': "PUB",
        'topics': topics,
        'message': {
            'type': type,
            'data': data_to_str(data, type),
        }
    }


def http_post_data_to_hub(data, topics):  # type: str
    msg = new_pub_message(
        data,
        type=MESSAGE_TYPE.JSON.name,
        topics=topics
    )
    return requests.post("https://hub.drink.cafe/http", json=msg)


def sub_topics(ws, topics):
    msg = {'action': "SUB", "topics": topics}
    ws.send(json.dumps(msg, ensure_ascii=False))
