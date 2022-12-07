package invoke

import (
	"context"

	"github.com/nanobus/iota/go/payload"
	"github.com/nanobus/iota/go/rx/flux"
	"github.com/nanobus/iota/go/rx/mono"
)

type (
	RequestResponseHandler func(context.Context, payload.Payload) mono.Mono[payload.Payload]
	FireAndForgetHandler   func(context.Context, payload.Payload)
	RequestStreamHandler   func(context.Context, payload.Payload) flux.Flux[payload.Payload]
	RequestChannelHandler  func(context.Context, payload.Payload, flux.Flux[payload.Payload]) flux.Flux[payload.Payload]
)

type HandlerInfo struct {
	Namespace string
	Operation string
}

type Operations struct {
	Exported Handlers
	Imported Handlers
}

type Handlers struct {
	RequestResponse []HandlerInfo
	FireAndForget   []HandlerInfo
	RequestStream   []HandlerInfo
	RequestChannel  []HandlerInfo
}

var (
	requestResponseHandlers    = make([]RequestResponseHandler, 0, 20)
	requestResponseHandlerInfo = make([]HandlerInfo, 0, 20)

	fireAndForgetHandlers    = make([]FireAndForgetHandler, 0, 20)
	fireAndForgetHandlerInfo = make([]HandlerInfo, 0, 20)

	requestStreamHandlers    = make([]RequestStreamHandler, 0, 20)
	requestStreamHandlerInfo = make([]HandlerInfo, 0, 20)

	requestChannelHandlers    = make([]RequestChannelHandler, 0, 20)
	requestChannelHandlerInfo = make([]HandlerInfo, 0, 20)

	requestResponseImports = make([]HandlerInfo, 0, 20)
	requestFNFImports      = make([]HandlerInfo, 0, 20)
	requestStreamImports   = make([]HandlerInfo, 0, 20)
	requestChannelImports  = make([]HandlerInfo, 0, 20)
)

func GetOperations() Operations {
	return Operations{
		Exported: Handlers{
			RequestResponse: requestResponseHandlerInfo,
			FireAndForget:   fireAndForgetHandlerInfo,
			RequestStream:   requestStreamHandlerInfo,
			RequestChannel:  requestChannelHandlerInfo,
		},
		Imported: Handlers{
			RequestResponse: requestResponseImports,
			FireAndForget:   requestFNFImports,
			RequestStream:   requestStreamImports,
			RequestChannel:  requestChannelImports,
		},
	}
}

func ExportRequestResponse(namespace, operation string, handler RequestResponseHandler) {
	requestResponseHandlers = append(requestResponseHandlers, handler)
	requestResponseHandlerInfo = append(requestResponseHandlerInfo, HandlerInfo{namespace, operation})
}

func GetRequestResponseHandler(operationID uint32) RequestResponseHandler {
	if operationID >= uint32(len(requestResponseHandlers)) {
		return nil
	}
	return requestResponseHandlers[operationID]
}

func ExportFireAndForget(namespace, operation string, handler FireAndForgetHandler) {
	fireAndForgetHandlers = append(fireAndForgetHandlers, handler)
	fireAndForgetHandlerInfo = append(fireAndForgetHandlerInfo, HandlerInfo{namespace, operation})
}

func GetFireAndForgetHandler(operationID uint32) FireAndForgetHandler {
	if operationID >= uint32(len(fireAndForgetHandlers)) {
		return nil
	}
	return fireAndForgetHandlers[operationID]
}

func ExportRequestStream(namespace, operation string, handler RequestStreamHandler) {
	requestStreamHandlers = append(requestStreamHandlers, handler)
	requestStreamHandlerInfo = append(requestStreamHandlerInfo, HandlerInfo{namespace, operation})
}

func GetRequestStreamHandler(operationID uint32) RequestStreamHandler {
	if operationID >= uint32(len(requestStreamHandlers)) {
		return nil
	}
	return requestStreamHandlers[operationID]
}

func ExportRequestChannel(namespace, operation string, handler RequestChannelHandler) {
	requestChannelHandlers = append(requestChannelHandlers, handler)
	requestChannelHandlerInfo = append(requestChannelHandlerInfo, HandlerInfo{namespace, operation})
}

func GetRequestChannelHandler(operationID uint32) RequestChannelHandler {
	if operationID >= uint32(len(requestChannelHandlers)) {
		return nil
	}
	return requestChannelHandlers[operationID]
}

func ImportRequestResponse(namespace, operation string) uint32 {
	for i, op := range requestResponseImports {
		if op.Namespace == namespace && op.Operation == operation {
			return uint32(i)
		}
	}
	id := uint32(len(requestResponseImports))
	requestResponseImports = append(requestResponseImports, HandlerInfo{namespace, operation})
	return id
}

func ImportFireAndForget(namespace, operation string) uint32 {
	for i, op := range requestFNFImports {
		if op.Namespace == namespace && op.Operation == operation {
			return uint32(i)
		}
	}
	id := uint32(len(requestFNFImports))
	requestFNFImports = append(requestFNFImports, HandlerInfo{namespace, operation})
	return id
}

func ImportRequestStream(namespace, operation string) uint32 {
	for i, op := range requestStreamImports {
		if op.Namespace == namespace && op.Operation == operation {
			return uint32(i)
		}
	}
	id := uint32(len(requestStreamImports))
	requestStreamImports = append(requestStreamImports, HandlerInfo{namespace, operation})
	return id
}

func ImportRequestChannel(namespace, operation string) uint32 {
	for i, op := range requestChannelImports {
		if op.Namespace == namespace && op.Operation == operation {
			return uint32(i)
		}
	}
	id := uint32(len(requestChannelImports))
	requestChannelImports = append(requestChannelImports, HandlerInfo{namespace, operation})
	return id
}
