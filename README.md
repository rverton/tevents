# tevents

tevents is a small tool to be deployed in your [tailnet](https://tailscale.com/kb/1136/tailnet/) to collect events/hooks from other services.
By using [tsnet](https://tailscale.com/blog/tsnet-virtual-private-services/), it can be deployed as a virtual private service. This allows to centrally collect
events in your network and display them on a web interface.

## Events
An event holds the following fields:

```
origin: (unique) identifier of the sender
type: type of the event (log, monitor)
body: content
owner: tailnet owner
```

There are two different types of events:

#### Log events:
Log events are simple one-time HTTP requests to to notify of a specific event.

```
2021-05-01 12:00:00 - networkwatcher - new device found connected to network
```

#### Monitor events
Monitor events are events that are sent periodically and allow you to graph execution. This allows for example to watch cron jobs for execution.

```
2021-06-01 12:00:00 - cron:backup - backup executed
2021-07-01 12:00:00 - cron:backup - backup executed
2021-09-01 12:00:00 - cron:backup - backup executed
```

Monitor events will be plotted by their tags.

## Submission

Events can be submitted via HTTP.

```
# example for log event
curl http://tevents/.log?origin=networkwatcher -d "new device found connected to network"

# example for monitoring a cron job executed every morning
0 1 * * * /usr/local/bin/backup.sh && curl http://tevents/.monitor?origin=cron:backup -d "executed";
```

## Development

All relevant tasks can be done via `make`:

```
make watch              # restart web server on code changes
make tailwind-watch     # watch tailwind css changes
make tailwind           # only build tailwind resources
make run                # run web server
make                    # build executable with all assets embedded
```
