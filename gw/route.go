package gw

import (
	"net/http"
	"strings"

	"github.com/nalanj/ladle/config"
	"github.com/nalanj/ladle/fn"
)

// route converts a request to its corresponding function
func route(conf *config.Config, r *http.Request) (*fn.Function, map[string]string) {
	for _, event := range conf.Events {
		pathParams, ok := routeMatch(r, event)
		if ok {
			return conf.Functions[event.Target], pathParams
		}
	}

	return nil, nil
}

// routeMatch tests if a route matches and returns path parts if it does
func routeMatch(r *http.Request, event *fn.Event) (map[string]string, bool) {

	if event.Source != fn.APISource {
		return nil, false
	}

	reqParts := strings.Split(r.URL.Path, "/")
	if reqParts[0] == "" {
		reqParts = reqParts[1:]
	}

	routeParts := strings.Split(event.Meta["Route"], "/")
	if routeParts[0] == "" {
		routeParts = routeParts[1:]
	}

	pathParams := make(map[string]string)

	readParts := 0
	for i := 0; i < len(routeParts); i++ {
		if len(reqParts) <= i {
			return nil, false
		}

		reqPart := reqParts[i]
		routePart := routeParts[i]

		if strings.HasPrefix(routePart, "{") && strings.HasSuffix(routePart, "}") {
			pathParams[routePart[1:len(routePart)-1]] = reqPart
		} else {
			if routePart != reqPart {
				return nil, false
			}
		}
		readParts++
	}

	if readParts < len(reqParts) {
		return nil, false
	}

	return pathParams, true
}
