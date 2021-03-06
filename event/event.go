package event

import (
	"context"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/hashicorp/go-uuid"
	"github.com/hr3685930/pkg/event/gochan"
	"github.com/hr3685930/pkg/event/http"
	"github.com/hr3685930/pkg/event/kafka"
	event "github.com/hr3685930/pkg/event/rpc"
	"github.com/pkg/errors"
	"log"
)

const DefaultSource = "https://github.com/hr3685930/pkg/event/sender"

var EventErrs = make(chan *EventErr, 1)

type EventErr struct {
	Err   error
	Event cloudevents.Event
}

type CEfn func(ctx context.Context, event cloudevents.Event) protocol.Result

type Event struct {
	CloudEvent
}

type CloudEvent interface {
	SetCloudEventID(id string)
	SetCloudEventType(topic string)
	SetCloudEventSource(source string)
	Send(ctx context.Context, msg interface{}) error
}

func NewHttpEvent(endpoint string, eventName string) *Event {
	httpEvent, err := http.NewHTTPEvent(endpoint)
	if err != nil {
		EventErrs <- &EventErr{
			Err: errors.Errorf("%+v\n", err),
		}
	}
	UUID, err := uuid.GenerateUUID()
	if err != nil {
		EventErrs <- &EventErr{
			Err: errors.Errorf("%+v\n", err),
		}
	}
	httpEvent.SetCloudEventID(UUID)
	httpEvent.SetCloudEventType(eventName)
	httpEvent.SetCloudEventSource(DefaultSource)
	return &Event{CloudEvent: httpEvent}
}

func NewHTTPReceive(ctx context.Context, fn CEfn) (*client.EventReceiver, error) {
	p, err := cloudevents.NewHTTP()
	if err != nil {
		return nil, errors.Errorf("%+v\n", err)
	}

	h, err := cloudevents.NewHTTPReceiveHandler(ctx, p, fn)
	if err != nil {
		return nil, errors.Errorf("%+v\n", err)
	}

	return h, nil
}

func NewKafkaEvent(topic string, eventName string) *Event {
	kafkaEvent, err := kafka.NewKafkaEvent(topic)
	if err != nil {
		EventErrs <- &EventErr{
			Err: errors.Errorf("%+v\n", err),
		}
	}
	UUID, err := uuid.GenerateUUID()
	if err != nil {
		EventErrs <- &EventErr{
			Err: errors.Errorf("%+v\n", err),
		}
	}
	kafkaEvent.SetCloudEventID(UUID)
	kafkaEvent.SetCloudEventType(eventName)
	kafkaEvent.SetCloudEventSource(DefaultSource)
	return &Event{CloudEvent: kafkaEvent}
}

func NewKafkaReceiver(ctx context.Context, topic, group string, fn CEfn) error {
	consumer := kafka_sarama.NewConsumerFromClient(kafka.EventClient, group, topic)
	c, err := cloudevents.NewClient(consumer)
	if err != nil {
		return err
	}

	log.Println("will listen consuming topic :", topic)
	err = c.StartReceiver(ctx, fn)
	if err != nil {
		return err
	} else {
		log.Printf("receiver stopped\n")
	}
	return nil
}

func NewChannelEvent(eventName string) *Event {
	ch, err := gochan.NewChannelEvent()
	if err != nil {
		EventErrs <- &EventErr{
			Err: errors.Errorf("%+v\n", err),
		}
	}
	UUID, err := uuid.GenerateUUID()
	if err != nil {
		EventErrs <- &EventErr{
			Err: errors.Errorf("%+v\n", err),
		}
	}
	ch.SetCloudEventID(UUID)
	ch.SetCloudEventType(eventName)
	ch.SetCloudEventSource(DefaultSource)
	return &Event{CloudEvent: ch}
}

func NewChanReceive(fn CEfn) error {
	ch, err := gochan.NewChannelEvent()
	if err != nil {
		return err
	}
	// Start the receiver
	go func() {
		if err := ch.Client.StartReceiver(ch.Context, fn); err != nil && err.Error() != "context deadline exceeded" {
			log.Fatalf("[receiver] channel event listen stop error: %s", err)
		}
		log.Println("channel event listen stop")
	}()

	return nil
}

func NewRpcEvent(endpoint, eventName string) *Event {
	r := event.NewRpcEvent(endpoint, eventName)
	UUID, err := uuid.GenerateUUID()
	if err != nil {
		EventErrs <- &EventErr{
			Err: errors.Errorf("%+v\n", err),
		}
	}
	r.SetCloudEventID(UUID)
	r.SetCloudEventType(eventName)
	r.SetCloudEventSource(DefaultSource)
	return &Event{CloudEvent: r}
}
