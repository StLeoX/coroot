package event

import (
	"github.com/ClickHouse/ch-go"
	chproto "github.com/ClickHouse/ch-go/proto"
	eventv1 "github.com/coroot/coroot/api/proto/coroot/event/v1"
	"k8s.io/klog"
	"sync"
	"time"
)

type ServerSpansBatch struct {
	limit int
	exec  func(query ch.Query) error

	lock sync.Mutex
	done chan struct{}

	Timestamp   *chproto.ColDateTime64
	Duration    *chproto.ColInt64
	ContainerId *chproto.ColLowCardinality[string]
	TgidRead    *chproto.ColStr
	TgidWrite   *chproto.ColStr
	RequestId   *chproto.ColStr
}

func NewServerSpansBatch(limit int, timeout time.Duration, exec func(query ch.Query) error) *ServerSpansBatch {
	b := &ServerSpansBatch{
		limit: limit,
		exec:  exec,
		done:  make(chan struct{}),

		Timestamp:   new(chproto.ColDateTime64).WithPrecision(chproto.PrecisionNano),
		Duration:    new(chproto.ColInt64),
		ContainerId: new(chproto.ColLowCardinality[string]),
		TgidRead:    new(chproto.ColStr),
		TgidWrite:   new(chproto.ColStr),
		RequestId:   new(chproto.ColStr),
	}

	go func() {
		ticker := time.NewTicker(timeout)
		defer ticker.Stop()
		for {
			select {
			case <-b.done:
				return
			case <-ticker.C:
				b.lock.Lock()
				b.save()
				b.lock.Unlock()
			}
		}
	}()

	return b
}

func (b *ServerSpansBatch) Close() {
	b.done <- struct{}{}
	b.lock.Lock()
	defer b.lock.Unlock()
	b.save()
}

func (b *ServerSpansBatch) Add(serverSpan *eventv1.ServerSpan) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.Timestamp.Append(serverSpan.Timestamp.AsTime())
	b.Duration.Append(serverSpan.Duration)
	b.ContainerId.Append(serverSpan.ContainerId)
	b.TgidRead.Append(serverSpan.TgidRead)
	b.TgidWrite.Append(serverSpan.TgidWrite)
	b.RequestId.Append(serverSpan.RequestId)

	if b.Timestamp.Rows() < b.limit {
		return
	}
	b.save()
}

func (b *ServerSpansBatch) save() {
	if b.Timestamp.Rows() == 0 {
		return
	}

	input := chproto.Input{
		{Name: "timestamp", Data: b.Timestamp},
		{Name: "duration", Data: b.Duration},
		{Name: "container_id", Data: b.ContainerId},
		{Name: "tgid_read", Data: b.TgidRead},
		{Name: "tgid_write", Data: b.TgidWrite},
		{Name: "request_id", Data: b.RequestId},
	}
	err := b.exec(ch.Query{Body: input.Into("@@table_ebpf_server_spans@@"), Input: input})
	if err != nil {
		klog.Errorln(err)
	}
	for _, i := range input {
		i.Data.(chproto.Resettable).Reset()
	}
}
