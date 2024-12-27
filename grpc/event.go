package grpc

import (
	"io"

	"k8s.io/klog"

	xcorootv1 "github.com/StLeoX/coroot-extend-api/api/proto/coroot/service/v1"
	"github.com/coroot/coroot/collector"
	"github.com/coroot/coroot/collector/event"
)

type ServerSpanServiceServer struct {
	xcorootv1.UnimplementedServerSpanServiceServer
	batcher *event.ServerSpansBatch
}

func (es *ServerSpanServiceServer) Upload(stream xcorootv1.ServerSpanService_UploadServer) error {
	resp := &xcorootv1.ServerSpanServiceUploadResponse{}

	for {
		req, err := stream.Recv()
		// client stream closed
		if err == io.EOF {
			break
		}
		if err != nil {
			klog.Warningln(err)
			return err
		}
		es.batcher.Add(req.Span)
	}

	err := stream.SendAndClose(resp)
	if err != nil {
		klog.Warningln(err)
		return err
	}
	return nil
}

func NewServerSpanServiceServer(coll *collector.Collector) *ServerSpanServiceServer {
	return &ServerSpanServiceServer{batcher: coll.GetServerSpansBatch(coll.GetCurrentProjectId())}
}
