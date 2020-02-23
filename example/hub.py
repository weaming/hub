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
            return b64encode(data).decode('utf8')
    return str(data)


def new_pub_message(
    data,
    type=MESSAGE_TYPE.PLAIN.name,
    topics=('global',),
    extended_data=[],
    captions=[],
):
    return {
        'action': "PUB",
        'topics': topics,
        'message': {
            'type': type,
            'data': data_to_str(data, type),
            'extended_data': [data_to_str(x, type) for x in extended_data],
            'captions': captions,
        },
    }


def http_post_data_to_hub(data, topics, type=MESSAGE_TYPE.JSON.name):
    msg = new_pub_message(data, type=type, topics=topics)
    return requests.post("https://hub.drink.cafe/http", json=msg)


def http_post_photos_to_hub(data: list, topics, captions=[]):
    msg = new_pub_message(
        data[0],
        type=MESSAGE_TYPE.IMAGE.name,
        topics=topics,
        extended_data=data[1:],
        captions=captions,
    )
    return requests.post("https://hub.drink.cafe/http", json=msg)


def sub_topics(ws, topics):
    msg = {'action': "SUB", "topics": topics}
    ws.send(json.dumps(msg, ensure_ascii=False))


if __name__ == '__main__':
    """
    e.g. python utils/message_hub.py hello -t test admin/ping
    """
    import argparse

    parser = argparse.ArgumentParser()
    parser.add_argument("--topics", nargs='+')
    parser.add_argument("--text", help="message text which will be send")
    parser.add_argument("--photo", nargs='*', help="path of photos")
    parser.add_argument(
        "--caption", nargs='*', help="captions of photos, missing at tail is allowed"
    )
    args = parser.parse_args()
    if args.photo:
        print(
            http_post_photos_to_hub(
                [
                    x if x.startswith('http') else open(x, 'rb').read()
                    for x in args.photo
                ],
                args.topics,
                captions=args.caption,
            ).text
        )
    else:
        print(
            http_post_data_to_hub(
                args.text.replace("\\n", "\n"),
                args.topics,
                type=MESSAGE_TYPE.PLAIN.name,
            ).text
        )
