#!/usr/bin/env python

import docker
import os
import sys
import requests
import json
import time

hostname = os.environ["CONMON_HOSTNAME"]
slack_webhook = os.environ["CONMON_SLACK_WEBHOOK"]
time_limit = os.environ["CONMON_TIME_LIMIT"] if "CONMON_TIME_LIMIT" in os.environ else 10 * 60
restart_limit = os.environ["CONMON_RESTART_LIMIT"] if "CONMON_RESTART_LIMIT"  in os.environ else 5
containers = {}
filters = {"type": "container", "event": "die"}

if not hostname:
    print("No CONMON_HOSTNAME set - exiting...")
    sys.exit(1)
if not slack_webhook:
    print("No CONMON_SLACK_WEBHOOK set - exiting...")
    sys.exit(1)

docker_client = docker.from_env()

def send_slack_message(message):
    data = {
        'text': message,
        'username': 'Conmon',
        'icon_emoji': ':bomb:'
    }
    response = requests.post(slack_webhook, data=json.dumps(
        data), headers={'Content-Type': 'application/json'})

for event in docker_client.events(decode=True, filters=filters):
    print(event)
    restart_time = event["time"]
    image_name = event["Actor"]["Attributes"]["image"]
    if (image_name in containers):
        c = containers[image_name]
        c["restarts_list"].append(restart_time)
        c["restarts_list"] = c["restarts_list"][-(restart_limit + 1):]
        c["restarts"] = c["restarts"] + 1
        if (c["restarts"] > restart_limit):
            if (c["restarts_list"][0] > restart_time - time_limit):
                extra_info = ""
                if "com.docker.swarm.node.id" in event["Actor"]["Attributes"]:
                    at = event["Actor"]["Attributes"]
                    extra_info = f" node_id {at['com.docker.swarm.node.id']} service_id {at['com.docker.swarm.service.id']} service_name {at['com.docker.swarm.service.name']}"
                print(f"Image {image_name} is restarting a lot on {hostname} {extra_info}")
                c["restarts"] = 1
                c["restarts_list"] = [restart_time]
                if c["last_slack_alert"] < time.time() - (10 * 60):
                    c["last_slack_alert"] = time.time()
                    send_slack_message(f"Image {image_name} is restarting a lot on {hostname} {extra_info}")
    else:
        containers[image_name] = {"restarts": 1, "restarts_list": [restart_time], "last_slack_alert": 0}
