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
	qos      byte
	pluginId string
	cli      interface{}
	protocol Protocol
	buffer   chan subscribe
}

type subscribe struct {
	Operate   command
	OperateId int64
	Body      []byte
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

func (ilink ILink) CommandsSubscribe() (c <-chan subscribe) {
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client)}
		mqtt.commandsSubscribe(ilink.buffer)
		return ilink.buffer
	}

	return nil
}

func (ilink ILink) HeartBeat() error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client), qos: ilink.qos}

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
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client), qos: ilink.qos}

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
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client), qos: ilink.qos}

		if err := mqtt.syncChannelTagStart(); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) deleteChannelResponse(operateId int64) error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client), qos: ilink.qos}

		if err := mqtt.deleteChannelRes(operateId); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) deleteAllChannelResponse(operateId int64) error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client), qos: ilink.qos}

		if err := mqtt.deleteAllChannelRes(operateId); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) syncChannelTagEndResponse(operateId int64) error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client), qos: ilink.qos}

		if err := mqtt.syncChannelTagEndResponse(operateId); err != nil {
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
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client), qos: ilink.qos}

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
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client), qos: ilink.qos}

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
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client), qos: ilink.qos}

		if err := mqtt.channelStatusUp(channelId, status); err != nil {
			return err
		}
	}
	return nil
}

/*
TagUp body

"data":{
  "channelId":"",
  "tags":[
	{
      "id":"",
      "v":"",
      "q":0,
      "ts":163982312442
    }
  ]
}
*/
func (ilink ILink) TagUp(channelId string, value string, quality byte) error {
	if ilink.cli == nil {
		return fmt.Errorf("client不允许为空")
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client), qos: ilink.qos}

		if err := mqtt.tagUp(channelId, value, quality); err != nil {
			return err
		}
	}
	return nil
	return nil
}

func NewMqtt(Ip string, port int, qos byte, pluginId string) *ILink {
	if mqttiLink != nil {
		return mqttiLink
	}

	c, err := newMqtt(Ip, port)
	if err != nil {
		return nil
	}
	mqttiLink = &ILink{pluginId: pluginId, cli: c, protocol: ProtocolMqtt, buffer: make(chan subscribe, 100), qos: qos}
	return mqttiLink
}

func NewTcpILink() *ILink {
	return nil
}
