[![CircleCI](https://circleci.com/gh/nalanj/ladle.svg?style=shield)](https://circleci.com/gh/nalanj/ladle)

![Ladle](https://user-images.githubusercontent.com/5594/57937720-85651d00-7894-11e9-8232-b4714b1d0872.jpg)

# Ladle

Ladle simplifies development of Go based Lambda functions by providing a local runtime environment and api gateway.

## Installation

Install ladle with `go install`:

```
go install github.com/nalanj/ladle
```

## Configuration

Ladle uses [Confl](https://github.com/nalanj/confl) for configuration:

```
# The Functions section defines functions. Each function has a name and
# a Package that is built to generate the executable.
Functions={
  Echo={
    Package="github.com/nalanj/ladle/lambdas/echo"
  }
}

# The events section defines events that trigger functions
Events=[

  # API events are fired from the built-in API Gateway
  {Source=API Target=Echo Meta={Route="/Echo/{name}"}}
]
```

## Commands

### Build functions

`ladle build` builds functions and stores their executables to the `.ladle`
directory, where Ladle looks for functions when invoking them. 

To build all configured functions:

```
ladle build
```

To build a specific function:

```
ladle build [Function]
```

### Serve Functions and API Gateway Requests

`ladle serve` starts the Ladle server and listens for function invocations and
API Gateway requests.

```
ladle serve
```

The API Gateway listens on port 3001 by default, but this can be configured with the `-a` flag:

```
ladle serve -a :3005
```

### Invoke Functions

`ladle invoke` invokes a function and returns its result.

To invoke a function from stdin:

```
echo "[payload]" | ladle invoke [function]
```

or to invoke a function based on a payload file:

```
ladle invoke [function] [payload.json]
```

## API Gateway

At present the API Gateway supports non-proxy routes. Proxy routes and websocket
support are planned.

The gateway also supports serving static resources from the `public/` directory. 
If a request matches a file in `public/` that file will be returned, rather than
invoking any functions.

---

*Ladle image courtesy National Gallery of Art, Washington*
