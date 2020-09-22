mfool
=====
Check the status of hosts, monitors and channels for forgot to enable again on Mackerel.

Prepare
-------
You need these variables.

- Mackerel API Key
- Slack Bot token

Download
--------
check release page

Run
---
```
$ ./mfool --mackerel-api-key '' --slack-token '' --slack-channel-id '' 
```

or

```
$ MFOOL_MACKEREL_API_KEY=''
$ MFOOL_SLACK_TOKEN=''
$ MFOOL_SLACK_CHANNEL_ID=''
$ ./mfool
```