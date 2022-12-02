package ilink

import (
	"errors"
	"fmt"
	emqx "github.com/eclipse/paho.mqtt.golang"
	"time"
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

var (
	ErrClientNull = errors.New("client不允许为空")
)

type ILink struct {
	OperateId int64
	qos       byte
	pluginId  string
	cli       interface{}
	protocol  Protocol
	buffer    chan subscribe
}

type subscribe struct {
	Operate   string
	OperateId int64
	Body      []byte
}

type Communication interface {
	HeartBeat() error                                                       //心跳上报
	Connect() error                                                         //插件连接网关
	SyncChannelTagStart() error                                             //同步CHANNEL和TAG通知
	SyncChannelTagEndResponse() error                                       //同步结束响应
	DeleteChannelResponse() error                                           //删除单个通道响应
	DeleteAllChannelResponse() error                                        //删除所有通道响应
	GetChannelStatusRes(channelId string, status ChannelStatus) error       //获取通道状态
	ChannelStatusUp() error                                                 //通道状态上报
	TagWriteResp() error                                                    //点位写入
	TagReadResp() error                                                     //点位读取
	TagUp(channelId string, tagId string, value string, quality byte) error //点位上报

}

func (ilink ILink) CommandsSubscribe() (c <-chan subscribe) {
	if ilink.protocol == ProtocolMqtt {
		mqtt := emq{pluginId: ilink.pluginId, client: ilink.cli.(emqx.Client)}
		mqtt.commandsSubscribe(ilink.buffer)
		return ilink.buffer
	}

	return nil
}

//func (ilink ILink) Sync(version string) error {
//	if ilink.cli == nil {
//		return ErrClientNull
//	}
//	if err := ilink.Connect(version); err != nil {
//		return err
//	}
//}

//HeartBeat 默认30秒一次
func (ilink ILink) HeartBeat() error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		mqtt := newEmq(ilink.pluginId, ilink.cli.(emqx.Client), ilink.qos)
		t := time.NewTicker(30 * time.Second)
		for {
			<-t.C
			if err := mqtt.heartBeat(); err != nil {
				return err
			}
		}

	}

	return fmt.Errorf("暂不支持的协议%v", ilink.protocol)
}

func (ilink ILink) Connect(ver string) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := newEmq(ilink.pluginId, ilink.cli.(emqx.Client), ilink.qos).connect(ver); err != nil {
			return err
		}

	}
	return nil
}

func (ilink ILink) SyncChannelTagStart() error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := newEmq(ilink.pluginId, ilink.cli.(emqx.Client), ilink.qos).syncChannelTagStart(); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) DeleteChannelResponse(operateId int64) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := newEmq(ilink.pluginId, ilink.cli.(emqx.Client), ilink.qos).deleteChannelRes(operateId); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) DeleteAllChannelResponse(operateId int64) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := newEmq(ilink.pluginId, ilink.cli.(emqx.Client), ilink.qos).deleteAllChannelRes(operateId); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) syncChannelTagEndResponse(operateId int64) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := newEmq(ilink.pluginId, ilink.cli.(emqx.Client), ilink.qos).syncChannelTagEndResponse(operateId); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) getChannelStatusRes(channelId string, status ChannelStatus) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := newEmq(ilink.pluginId, ilink.cli.(emqx.Client), ilink.qos).getChannelStatusRes(channelId, status); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) channelStatusUp(channelId string, status ChannelStatus) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := newEmq(ilink.pluginId, ilink.cli.(emqx.Client), ilink.qos).channelStatusUp(channelId, status); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) tagReadResp(channelId string, status ChannelStatus) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := newEmq(ilink.pluginId, ilink.cli.(emqx.Client), ilink.qos).channelStatusUp(channelId, status); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) TagUp(channelId string, tagId string, value string, quality byte) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := newEmq(ilink.pluginId, ilink.cli.(emqx.Client), ilink.qos).tagUp(channelId, tagId, value, quality); err != nil {
			return err
		}
	}
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
	mqttiLink = &ILink{OperateId: 0, pluginId: pluginId, cli: c, protocol: ProtocolMqtt, buffer: make(chan subscribe, 100), qos: qos}
	return mqttiLink
}

func NewTcpILink() *ILink {
	return nil
}
