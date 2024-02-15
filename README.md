### Component Inventory Service
This is the code for the service sitting between Alloy and other consumers of
component data. It exists to synthesize a common union of facts discovered
by each of Alloys operating modes (`in-band` and `out-of-band`). These modes
refer to whether component data collection was performed in the host by an 
application or if the survey was completed via the BMC of the server.

