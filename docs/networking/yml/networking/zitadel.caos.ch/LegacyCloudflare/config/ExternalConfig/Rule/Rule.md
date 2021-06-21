# Rule 
 

## Structure 
 

| Attribute   | Description                                                                                                                                                                                | Default | Collection | Map  |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------- | ---------- | ---  |
| description | Description given to the firewall rule                                                                                                                                                     |         |            |      |
| priority    | The priority of the rule to allow control of processing order. A lower number indicates high priority. If not provided, any rules with a priority will be sequenced before those without.  |         |            |      |
| action      | The action to apply to a matched request. Note that action "log" is only available for enterprise customers.                                                                               |         |            |      |
| filters     | Definition of the filter used , [here](Filter/Filter.md)                                                                                                                                   |         | X          |      |