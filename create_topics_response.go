package sarama

import (
	"fmt"
	"time"
)

type CreateTopicsResponse struct {
	// Version defines the protocol version to use for encode and decode
	Version int16
	// ThrottleTime contains the duration for which the request was throttled due
	// to a quota violation, or zero if the request did not violate any quota.
	ThrottleTime time.Duration
	// TopicErrors contains a map of any errors for the topics we tried to create.
	TopicErrors map[string]*TopicError
}

func (c *CreateTopicsResponse) setVersion(v int16) {
	c.Version = v
}

func (c *CreateTopicsResponse) encode(pe packetEncoder) error {
	if c.Version >= 2 {
		pe.putInt32(int32(c.ThrottleTime / time.Millisecond))
	}

	if err := pe.putArrayLength(len(c.TopicErrors)); err != nil {
		return err
	}
	for topic, topicError := range c.TopicErrors {
		if err := pe.putString(topic); err != nil {
			return err
		}
		if err := topicError.encode(pe, c.Version); err != nil {
			return err
		}
	}

	return nil
}

func (c *CreateTopicsResponse) decode(pd packetDecoder, version int16) (err error) {
	c.Version = version

	if version >= 2 {
		throttleTime, err := pd.getInt32()
		if err != nil {
			return err
		}
		c.ThrottleTime = time.Duration(throttleTime) * time.Millisecond
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	c.TopicErrors = make(map[string]*TopicError, n)
	for i := 0; i < n; i++ {
		topic, err := pd.getString()
		if err != nil {
			return err
		}
		c.TopicErrors[topic] = new(TopicError)
		if err := c.TopicErrors[topic].decode(pd, version); err != nil {
			return err
		}
	}

	return nil
}

func (c *CreateTopicsResponse) key() int16 {
	return apiKeyCreateTopics
}

func (c *CreateTopicsResponse) version() int16 {
	return c.Version
}

func (c *CreateTopicsResponse) headerVersion() int16 {
	return 0
}

func (c *CreateTopicsResponse) isValidVersion() bool {
	return c.Version >= 0 && c.Version <= 3
}

func (c *CreateTopicsResponse) requiredVersion() KafkaVersion {
	switch c.Version {
	case 3:
		return V2_0_0_0
	case 2:
		return V0_11_0_0
	case 1:
		return V0_10_2_0
	case 0:
		return V0_10_1_0
	default:
		return V2_8_0_0
	}
}

func (r *CreateTopicsResponse) throttleTime() time.Duration {
	return r.ThrottleTime
}

type TopicError struct {
	Err    KError
	ErrMsg *string
}

func (t *TopicError) Error() string {
	text := t.Err.Error()
	if t.ErrMsg != nil {
		text = fmt.Sprintf("%s - %s", text, *t.ErrMsg)
	}
	return text
}

func (t *TopicError) Unwrap() error {
	return t.Err
}

func (t *TopicError) encode(pe packetEncoder, version int16) error {
	pe.putInt16(int16(t.Err))

	if version >= 1 {
		if err := pe.putNullableString(t.ErrMsg); err != nil {
			return err
		}
	}

	return nil
}

func (t *TopicError) decode(pd packetDecoder, version int16) (err error) {
	kErr, err := pd.getInt16()
	if err != nil {
		return err
	}
	t.Err = KError(kErr)

	if version >= 1 {
		if t.ErrMsg, err = pd.getNullableString(); err != nil {
			return err
		}
	}

	return nil
}
