package tcp

import (
	"context"

	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/wflow/proto"
)

type MessageHandler interface {
	Handle(context.Context, GeneralConnection, *proto.Message) error
}

type messageHandler struct {
	processors map[proto.Type][]MessageProcessor
}

func NewMessageHandler(processors map[proto.Type][]MessageProcessor) MessageHandler {
	if processors[proto.Type_PING] == nil {
		processors[proto.Type_PING] = []MessageProcessor{HandlePing}
	}
	return &messageHandler{
		processors: processors,
	}
}

func (h *messageHandler) Handle(ctx context.Context, c GeneralConnection, m *proto.Message) (err error) {
	processors := h.processors[m.Type]
	if processors == nil {
		err = errors.Error("can not find processor for message type " + m.Type.String())
		return
	}
	multiErrors := errors.NewMultiError()
	for _, processor := range processors {
		err = processor(ctx, c, m)
		if err != nil {
			multiErrors.Add(err)
		}
	}
	err = multiErrors
	return
}
