package cache

import (
	"encoding/json"

	"github.com/go-redis/redis"
)

func (c *RedisCache) PublishMessage(channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return c.client.Publish(channel, string(data)).Err()
}

func (c *RedisCache) Subscribe(channels ...string) (PubSubSubscription, error) {
	pubsub := c.client.Subscribe(channels...)
	return &RedisPubSubSubscription{pubsub: pubsub}, nil
}

type RedisPubSubMessage struct {
	message *redis.Message
}

func (m *RedisPubSubMessage) GetChannel() string {
	return m.message.Channel
}

func (m *RedisPubSubMessage) GetPayload() string {
	return m.message.Payload
}

type RedisPubSubSubscription struct {
	pubsub *redis.PubSub
}

func (s *RedisPubSubSubscription) ReceiveMessage() (PubSubMessage, error) {
	msg, err := s.pubsub.ReceiveMessage()
	if err != nil {
		return nil, err
	}
	return &RedisPubSubMessage{message: msg}, nil
}

func (s *RedisPubSubSubscription) Close() error {
	return s.pubsub.Close()
}
