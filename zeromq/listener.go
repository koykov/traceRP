package zeromq

import (
	"bytes"
	"context"

	"github.com/koykov/fastconv"
	"github.com/koykov/traceID"
	"github.com/koykov/traceID/listener"
	"github.com/pebbe/zmq4"
)

type ZeroMQ struct {
	listener.Base
}

func (l ZeroMQ) Listen(ctx context.Context, out chan []byte) (err error) {
	conf := l.GetConfig()
	if len(conf.Topic) == 0 {
		conf.Topic = traceID.DefaultZeroMQTopic
	}

	var (
		ztx *zmq4.Context
		zsk *zmq4.Socket
	)
	if ztx, err = zmq4.NewContext(); err != nil {
		return
	}
	if zsk, err = ztx.NewSocket(zmq4.SUB); err != nil {
		return
	}
	if conf.HWM == 0 {
		conf.HWM = traceID.DefaultZeroMQHWM
	}
	if err = zsk.SetSndhwm(int(conf.HWM)); err != nil {
		return
	}
	if err = zsk.Connect(conf.Addr); err != nil {
		return
	}
	if err = zsk.SetSubscribe(conf.Topic); err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			_ = zsk.Close()
			return
		default:
			var p []byte
			for {
				if p, err = zsk.RecvBytes(0); err != nil || len(p) == 0 {
					continue
				}
				if l.isTopic(p) {
					continue
				}
				break
			}
			out <- p
		}
	}
}

func (l ZeroMQ) isTopic(p []byte) bool {
	return bytes.Equal(p, fastconv.S2B(traceID.DefaultZeroMQTopic)) || bytes.Equal(p, fastconv.S2B(traceID.ProtobufZeroMQTopic))
}
