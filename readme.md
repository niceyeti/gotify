# Gotify: Go Notifications for Linux Dudes

![It may not spark joy, but it gets the job done](goatboy.png)

Sometimes you're a hillbilly with a wrench.

Sometimes you're a hillbilly with a wrench and want to know when a certain year and model are available.

Sometimes you're a hillbilly with a wrench and want to know when a certain year and model are available via login notifications.

Gotify is a surly, single-minded, synchronous, under-abstracted login program used to notify myself of things available on the web,
using pre-compiled definitions (parsers).

As an automotive restorer I want to be notified of the availability of rare models, but surfing online does not encompass a comprehensive search space and is a poor use of personal time. But so is building a fully abstracted solution with external query/parser definitions, configurable outputs/notifications, or feature-bloated structured information. All I need is a rude kick between the shoulder bladers to notify me using a 
couple bits of information: "model _____ is available at ____".

## The gist
Gotify can be compiled and called synchronously via login or job scripts.
When called, gotify queries and parses compiled sources for specific information; if found, it sends notifications via notify-send.
This is intentionally under-abstracted and single-purpose.

## Requirements
* No nice things: single file of code
* Code golf: solutions not systems
* Hold my beer: run/crash test coverage


## Ethics
Gotify only queries online sources in a polite, synchronous, and non-polling manner, to bring business to the
sources it queries. Gotify is intended for personal use, not as a metasearcher, and utilizes web resources in the same
manner as an ordinary user, while consuming even fewer resources.
