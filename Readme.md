# Conmon

[WIP!] Basic slack alerts for failing containers

This is just a very lightweight golang project that will check for containers failing more than a certain amount in a certain timeframe.
It'll then send a message to a slack webhook.

This is by no means meant to be a proper 'production monitoring solution' - it's just a quick, small sanity check.  We've sometimes had
a situation where a container would get stuck in a rapid crash/restart loop and use up a lot of resources on a QA/test server - it's just to
give us a little heads up.

## Usage (commandline)

```
conmon [-config=/path/to/config.yaml -hostname=my-custom-hostname -proxy=http://your-corporate-proxy.com:8080 -slackchannel=alerts --slackhook=http://hooks.slack.com/your-webhook]
```

## Usage (docker)

```
docker run -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd)/config.yaml:/config.yaml uogsoe/conmon:0.1 conmon -config /config.yaml
```

## Config

The config file is a plain yaml file.  Options are :
```
alert_dies: 2
alert_minutes: 5
alert_frequency: 10
slack_webhook: "http://your-slack-webhook-url"
slack_channel: "your-slack-channel-of-choice"
hostname: "custom-hostname-if-you-want"
proxy: "http://your-proxy-if-you-need-one"
```
`alert_dies` is the number of times a container can die in `alert_minutes`. The `alert_frequency` is how many minutes to wait before
notifying you again about the same container.

## Building

If you want to build this yourself, you can do (assuming you have golang installed)
```
go get -v github.com/ohnotnow/conmon/...
go build -a -o conmon .
```
You can also build your own docker image - have a look at the provided Dockerfile.

## Notes

Conmon bases it's idea of 'a container' on the _image name_. So if you have multiple services running the same image and they fall over `conmon` treats
them as one container.  That's what made sense in my context, so there it is...
