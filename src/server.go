package main

import (
	"fmt"
	"net/http"
)

//Routes Defines all routes
var Routes map[string](*HandlerFunction)

//RoutesHandler handles routes
type RoutesHandler struct {
	routes   *map[string](*HandlerFunction)
	workPool *WorkerPool
}

//RegisterRoute Registers a URL Route
func (rh *RoutesHandler) RegisterRoute(route string, fun func(w *http.ResponseWriter, r *http.Request, c chan<- bool)) {
	handle := HandlerFunction(fun)
	fmt.Println(&handle)
	(*rh.routes)[route] = &handle
}

//Dispatch Dispatches a task from the map to work-queue
func (rh *RoutesHandler) Dispatch(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.EscapedPath()

	handler, ok := (*rh.routes)[uri]
	if !ok {
		fmt.Fprintf(w, "<h4>Not found</h4>")
	} else {
		dataChannel := make(chan bool)
		rh.workPool.SubmitJob(&w, r, handler, dataChannel)

		<-dataChannel
	}
}

//NewRouteHandler create a new route handler
func NewRouteHandler(nWorkers int, queueSize int) *RoutesHandler {
	routesHandler := RoutesHandler{}

	handlerMap := make(map[string](*HandlerFunction))
	routesHandler.routes = &handlerMap
	routesHandler.workPool = NewWorkerPool(nWorkers, queueSize)

	return &routesHandler
}
