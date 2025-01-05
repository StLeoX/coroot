package grpc

import (
	"context"
	v1 "github.com/coroot/coroot/api/proto/coroot/service/v1"
	"github.com/coroot/coroot/collector"
	"github.com/coroot/coroot/collector/event"
)

type EventServiceServer struct {
	v1.UnimplementedEventServiceServer
	batcher *event.ServerSpansBatch
}

func (es *EventServiceServer) Upload(_ context.Context, req *v1.EventServiceUploadRequest) (*v1.EventServiceUploadResponse, error) {
	es.batcher.Add(req)
	return &v1.EventServiceUploadResponse{}, nil
}

func NewEventServiceServer(coll *collector.Collector) *EventServiceServer {
	return &EventServiceServer{
		batcher: coll.GetServerSpansBatch(coll.GetCurrentProjectId()),
	}
}
