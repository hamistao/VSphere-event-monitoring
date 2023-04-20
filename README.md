# VSphere-event-monitoring
Solution to get and filter events on a VSphere cluster

## get_events
Change the .env to fit your environment and your goal and just run it with `python ./get_events.py`. For practicallity, the desired event types can be changed in the file under the variable EVENTS. The possible event identifiers are on the following file.
The go version uses the same .env and works in a very similar way. You can run it with `go run ./get_events.go`. Beware that VMWare's SDK for Golang has some bugs and may not work well when filtering events by type.

## event_list
List of (i think) every possible VSphere event type identifier.