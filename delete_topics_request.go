package sarama

import "time"

type DeleteTopicsRequest struct {
	Version int16
	Topics  []string
	Timeout time.Duration
}

func (d *DeleteTopicsRequest) setVersion(v int16) {
	d.Version = v
}

func NewDeleteTopicsRequest(version KafkaVersion, topics []string, timeout time.Duration) *DeleteTopicsRequest {
	d := &DeleteTopicsRequest{
		Topics:  topics,
		Timeout: timeout,
	}
	if version.IsAtLeast(V2_1_0_0) {
		d.Version = 3
	} else if version.IsAtLeast(V2_0_0_0) {
		d.Version = 2
	} else if version.IsAtLeast(V0_11_0_0) {
		d.Version = 1
	}
	return d
}

func (d *DeleteTopicsRequest) encode(pe packetEncoder) error {
	if err := pe.putStringArray(d.Topics); err != nil {
		return err
	}
	pe.putInt32(int32(d.Timeout / time.Millisecond))

	return nil
}

func (d *DeleteTopicsRequest) decode(pd packetDecoder, version int16) (err error) {
	if d.Topics, err = pd.getStringArray(); err != nil {
		return err
	}
	timeout, err := pd.getInt32()
	if err != nil {
		return err
	}
	d.Timeout = time.Duration(timeout) * time.Millisecond
	d.Version = version
	return nil
}

func (d *DeleteTopicsRequest) key() int16 {
	return apiKeyDeleteTopics
}

func (d *DeleteTopicsRequest) version() int16 {
	return d.Version
}

func (d *DeleteTopicsRequest) headerVersion() int16 {
	return 1
}

func (d *DeleteTopicsRequest) isValidVersion() bool {
	return d.Version >= 0 && d.Version <= 3
}

func (d *DeleteTopicsRequest) requiredVersion() KafkaVersion {
	switch d.Version {
	case 3:
		return V2_1_0_0
	case 2:
		return V2_0_0_0
	case 1:
		return V0_11_0_0
	case 0:
		return V0_10_1_0
	default:
		return V2_2_0_0
	}
}
