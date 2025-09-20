package eventbus

type Event interface {
	isEvent()
}
type BaseEvent struct {}

func (e BaseEvent) isEvent() {}

type EventHandler func(Event)
