package grpc

import (
	"io"

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
		// end of downstream
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		es.batcher.Add(req.Span)
	}
	return stream.SendAndClose(resp)
}

func NewServerSpanServiceServer(coll *collector.Collector) *ServerSpanServiceServer {
	return &ServerSpanServiceServer{batcher: coll.GetServerSpansBatch(coll.GetCurrentProjectId())}
}
