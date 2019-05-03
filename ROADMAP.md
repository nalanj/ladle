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
  - Add meta for confl events
  - Route based on meta
  - Flesh out request conversions

- Add timings to call logs
- Automatic function restart

- TEST EVERYTHING

- Simple UI for seeing list of registered functions
- List of registered routes in gateway
- Call functions from UI
- Predefined payloads

- Support for auth
- Support for SQS
- Support for sockets
- Support for timers/cron


