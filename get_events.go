package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

func getTimelapseEvents(ctx context.Context, vim25C *vim25.Client, timelapse int, host string) ([]types.BaseEvent, error) {

	m := event.NewManager(vim25C)

	ref := vim25C.ServiceContent.RootFolder

	now, err := methods.GetCurrentTime(ctx, vim25C) // vCenter server time (UTC)
	if err != nil {
		return nil, err
	}

	filter := types.EventFilterSpec{
		Entity: &types.EventFilterSpecByEntity{
			Entity:    ref,
			Recursion: types.EventFilterSpecRecursionOptionAll,
		},
		Type: []string{"VmBeingCreatedEvent", "VmMigratedEvent", "VmCreatedEvent", "VmRemovedEvent", "VmBeingClonedEvent", "VmRelocatedEvent"},
	}

	if timelapse > 0 {
		begin := time.Duration(timelapse) * time.Minute
		filter.Time = &types.EventFilterSpecByTime{
			BeginTime: types.NewTime(now.Add(begin * -1)),
		}
	}

	event_collector, err := m.CreateCollectorForEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	defer event_collector.Destroy(ctx)

	formatedEvents := []types.BaseEvent{}

	events, err := event_collector.ReadNextEvents(ctx, 100)
	if err != nil {
		return nil, err
	}

	if len(host) > 0 {
		events = filterEventsByHost(events, host)
	}

	for _, e := range events {

		event := e.GetEvent()

		if event.Host != nil && event.Ds != nil {

			formatedEvents = append(formatedEvents, e)
		}
	}

	return formatedEvents, nil
}

func filterEventsByHost(events []types.BaseEvent, host string) []types.BaseEvent {
	count := 0

	for i := range events {
		event := events[i].GetEvent()
		if event.Host != nil {
			if event.Host.Name == host {
				count++
			}
		}
	}

	filteredEvents := make([]types.BaseEvent, count)
	index := 0

	for i := range events {
		event := events[i].GetEvent()
		if event.Host != nil {
			if event.Host.Name == host {
				filteredEvents[index] = event
				index++
			}
		}
	}

	return filteredEvents
}

func authenticateGovmonmi(credentialsURL string) (*vim25.Client, *govmomi.Client, error) {
	vcenterInfo, err := url.Parse(credentialsURL)
	if err != nil {
		log.Println("Error parsing vCenter URL " + credentialsURL)
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	govC, err := govmomi.NewClient(ctx, vcenterInfo, true)
	if err != nil {
		log.Println("Error creating Govmomi Client")
		return nil, nil, err
	}

	log.Println("Log in successful in vCenter " + vcenterInfo.Hostname())

	vim25C, err := vim25.NewClient(ctx, govC.RoundTripper)
	if err != nil {
		log.Println("Error creating vim25 Client")
		return nil, nil, err
	}

	return vim25C, govC, nil

}

func printEvents(events []types.BaseEvent) {

	for i := range events {
		event := events[i].GetEvent()
		fmt.Printf("[%s] %s\n", event.CreatedTime.Format(time.ANSIC), event.FullFormattedMessage)
		fmt.Println(event.Host)
	}
}

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("It was not possible to load the information present in the .env file: %s", err)
	}

	vcenterIP := os.Getenv("VCENTER_IP")
	username := os.Getenv("VCENTER_USERNAME")
	password := os.Getenv("VCENTER_PASSWORD")
	host := ""     // set to a ESXi host's IP to only get events from that host
	timelapse := 0 // set to get events from the last timelapse minutes

	vcenterURL := "https://" + username + ":" + password + "@" + vcenterIP + "/sdk"

	vim25Client, _, err := authenticateGovmonmi(vcenterURL)
	if err != nil {
		log.Fatalf("Unable to authenticate vim25 Client to vCenter: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	event_slice, err := getTimelapseEvents(ctx, vim25Client, timelapse, host)
	if err != nil {
		log.Fatalf("Failed to get events from cluster: %s", err)
	}
	printEvents(event_slice)
}
