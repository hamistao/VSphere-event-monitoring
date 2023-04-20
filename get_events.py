from datetime import datetime, timedelta
import ssl, atexit, os
from pyVim import connect # get VMWare's python sdk at https://github.com/vmware/pyvmomi
from pyVmomi import vim
from dotenv import load_dotenv

load_dotenv()

# HOST = "10.141.42.240"
VC = os.getenv("VCENTER_ADDR")
USER = os.getenv("VCENTER_USER")
PWD = os.getenv("VCENTER_PWD")
period = None
host = None
EVENTS = ["VmMigratedEvent", "VmCreatedEvent", "VmRemovedEvent", "VmBeingClonedEvent", "VmRelocatedEvent"] # example of desired events

def get_host_events(period=None, events=None, host=None):
    # Getting the Sevice Instance
    context = ssl.SSLContext(ssl.PROTOCOL_TLSv1_2)
    context.verify_mode = ssl.CERT_NONE
    si = connect.SmartConnect(protocol="https", host=VC, port=443, user=USER, pwd=PWD, sslContext=context)

    #Cleanly disconnect
    atexit.register(connect.Disconnect, si)

    filter_spec = vim.event.EventFilterSpec()

    if events: # if list of event types was provided
        filter_spec.eventTypeId = events
        
    if period: # if period was provided
        time_filter = vim.event.EventFilterSpec.ByTime()
        now = datetime.now()
        time_filter.beginTime = now - timedelta(hours=period)
        time_filter.endTime = now
        filter_spec.time = time_filter
        
    content = si.RetrieveServiceContent()

    event_manager = content.eventManager
    
    # event_res = event_manager.QueryEvents(filter_spec)
    event_collector = event_manager.CreateCollectorForEvents(filter_spec)
    page_size = 1000
    out_events = []
    
    while True:
        events_in_page = event_collector.ReadNextEvents(page_size)
        print(len(events_in_page))
        if not events_in_page:
            break
        
        if not host:
            out_events.extend(events_in_page)
            continue
        
        # if desired host was provided
        page_events = []
        for e in events_in_page:
            if "snapshot" in e.fullFormattedMessage:
            # if e.host and e.host.host.summary.config.name == host:
                page_events.append(e)
        out_events.extend(page_events)
        
    return out_events

for e in get_host_events(events=EVENTS, period=period, host=host):
    print("{} \033[32m@\033[0m {:%Y-%m-%d %H:%M:%S} by {}".format(e.fullFormattedMessage, e.createdTime, e.userName))