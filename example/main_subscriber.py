"""
Subscribe on topics and print the message in JSON format when received
"""
import sys
import json
from hub import run_until_close, sub_topics

JSON_DUMP_PARAMS = {"ensure_ascii": False, "indent": 2}


def on_message(ws, msg):
    msg = json.loads(msg)
    if msg["type"] == "MESSAGE":
        print(msg["topic"])
        if msg["message"]["type"] == "JSON":
            print(json.dumps(json.loads(msg["message"]["data"]), **JSON_DUMP_PARAMS))
        else:
            print(json.dumps(msg["message"], **JSON_DUMP_PARAMS))
    else:
        print(msg)
    print("-" * 100)


def bee(ws):
    topics = sys.argv[1:]
    print(topics)
    sub_topics(ws, topics)


run_until_close(on_msg=on_message, bee=bee)
