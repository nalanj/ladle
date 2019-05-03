x Get Cobra going for cli system

x Invoke command

x Custom port for rpc

x Confl config loading for multiple functions
  - Get confl nodes to include debug info so errors are clearer
  - Clear up errors in parsing config files
  - Confl node.IsText() to help with checking word and string type constantly
  - Confl map iteration solution
  - Block duplicate function names
  - Block duplicate function sections

- HTTP gateway for calling functions by route over http
  x Extract bad routing logic to function
  x Add meta for confl events
  x Route based on meta
  - Flesh out request conversions
  - Proxy matches

x Add request id as uuid on invocation and trace in logs with it

x Makefile

- Block requests to one per lambda at a time to control for logging
x Add timings to call logs
- Automatic function restart

- TEST EVERYTHING

- Base64 body handling

- Simple UI for seeing list of registered functions
- List of registered routes in gateway
- Call functions from UI
- Predefined payloads

- Support for auth
- Support for SQS
- Support for sockets
- Support for timers/cron


