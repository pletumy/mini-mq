package model

type TopicData struct {
	subs    []chan Message
	history []Message
}
