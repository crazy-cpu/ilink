package ilink

import (
	"fmt"
	emqx "github.com/eclipse/paho.mqtt.golang"
)

type (
	Protocol byte
	Body     int
)

var version = "1.1"

const (
	BinaryBody   = Body(-2)
	JsonBody     = Body(-1)
	ProtocolMqtt = Protocol(0)
	ProtocolTcp  = Protocol(1)
)

const (
	upTopic   = "up/bridge"
	downTopic = "down/bridge"
)

var mqttiLink *ILink

type ILink struct {
	cli      interface{}
	protocol Protocol
	buffer   chan subscribe
}

type subscribe struct {
	Operate command
	Body    []byte
}

type Communication interface {
	HeartBeat() error                                                 //心跳上报
	Connect() error                                                   //插件连接网关
	SyncChannelTagStart() error                                       //同步CHANNEL和TAG通知
	SyncChannelTagEndResponse() error                                 //同步结束响应
	DeleteChannelResponse() error                                     //删除单个通道响应
	DeleteAllChannelResponse() error                                  //删除所有通道响应
	GetChannelStatusRes(channelId string, status ChannelStatus) error //获取通道状态
	ChannelStatusUp() error                                           //通道状态上报
	TagWriteResp() error                                              //点位写入
	TagReadResp() error                                               //点位读取
	TagUp() error                                                     //点位上报
}

func (ilink ILink) CommandsSubscribe(pluginId string) (c <-chan subscribe) {
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{client: ilink.cli.(emqx.Client)}
		mqtt.commandsSubscribe(pluginId, ilink.buffer)
		return ilink.buffer
	}

	return nil
}

func (ilink ILink) HeartBeat() error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{client: ilink.cli.(emqx.Client)}

		if err := mqtt.heartBeat(); err != nil {
			return err
		}
	}

	return fmt.Errorf("暂不支持的协议%v", ilink.protocol)
}

func (ilink ILink) Connect(ver string) error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{client: ilink.cli.(emqx.Client)}

		if err := mqtt.connect(ver); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) SyncChannelTagStart() error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{client: ilink.cli.(emqx.Client)}

		if err := mqtt.syncChannelTagStart(); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) deleteChannelResponse() error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{client: ilink.cli.(emqx.Client)}

		if err := mqtt.deleteChannelRes(); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) deleteAllChannelResponse() error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{client: ilink.cli.(emqx.Client)}

		if err := mqtt.deleteAllChannelRes(); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) syncChannelTagEndResponse() error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{client: ilink.cli.(emqx.Client)}

		if err := mqtt.syncChannelTagEndResponse(); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) getChannelStatusRes(channelId string, status ChannelStatus) error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{client: ilink.cli.(emqx.Client)}

		if err := mqtt.getChannelStatusRes(channelId, status); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) channelStatusUp(channelId string, status ChannelStatus) error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{client: ilink.cli.(emqx.Client)}

		if err := mqtt.channelStatusUp(channelId, status); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) tagReadResp(channelId string, status ChannelStatus) error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{client: ilink.cli.(emqx.Client)}

		if err := mqtt.channelStatusUp(channelId, status); err != nil {
			return err
		}
	}
	return nil
}

func NewMqtt(address string) *ILink {
	if mqttiLink != nil {
		return mqttiLink
	}

	c, err := newMqtt(address)
	if err != nil {
		return nil
	}
	mqttiLink = &ILink{cli: c, protocol: ProtocolMqtt, buffer: make(chan subscribe, 100)}
	return mqttiLink
}

func NewTcpILink() *ILink {
	return nil
}
