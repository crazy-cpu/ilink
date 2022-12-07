package ilink

import (
	"errors"
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
	operateId int64
	qos       byte
	pluginId  string
	cli       interface{}
	protocol  Protocol
	buffer    chan subscribe
}

type subscribe struct {
	Operate   command
	OperateId int64
	Body      []byte
}

type Communication interface {
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

func (ilink ILink) Subscribe() (c <-chan subscribe) {
	return ilink.buffer
}

// PluginConnectGateway block function
//func (ilink ILink) PluginConnectGateway(version string) (<-chan subscribe, error) {
//	if mq != nil {
//		return ilink.buffer, nil
//	}
//
//	if ilink.protocol == ProtocolMqtt {
//		if ilink.cli == nil {
//			return ilink.buffer, ErrClientNull
//		}
//
//		mq.commandsSubscribe(ilink.buffer)
//		if err := mq.connect(version); err != nil {
//			return ilink.buffer, err
//		}
//
//		ilink.HeartBeat()
//
//	}
//	return ilink.buffer, nil
//}

//HeartBeat 默认30秒一次
//func (ilink ILink) HeartBeat() error {
//	if ilink.cli == nil {
//		return ErrClientNull
//	}
//	if ilink.protocol == ProtocolMqtt {
//		t := time.NewTicker(30 * time.Second)
//		go func() {
//			for {
//				<-t.C
//				mq.heartBeat()
//
//			}
//		}()
//
//	}
//
//	return fmt.Errorf("暂不支持的协议%v", ilink.protocol)
//}

func (ilink ILink) Connect(ver string) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := mq.connect(ver); err != nil {
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
		if err := mq.syncChannelTagStart(); err != nil {
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
		if err := mq.deleteChannelRes(operateId); err != nil {
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
		if err := mq.deleteAllChannelRes(operateId); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) SyncChannelTagEndResponse(operateId int64) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := mq.syncChannelTagEndResponse(operateId); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) GetChannelStatusRes(channelId string, status ChannelStatus) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := mq.getChannelStatusRes(channelId, status); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) ChannelStatusUp(channelId string, status ChannelStatus) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := mq.channelStatusUp(channelId, status); err != nil {
			return err
		}
	}
	return nil
}

func (ilink ILink) TagReadResp(channelId string, status ChannelStatus) error {
	if ilink.cli == nil {
		return ErrClientNull
	}
	if ilink.protocol == ProtocolMqtt {
		if err := mq.channelStatusUp(channelId, status); err != nil {
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
		if err := mq.tagUp(channelId, tagId, value, quality); err != nil {
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
	mqttiLink = &ILink{operateId: 0, pluginId: pluginId, cli: c, protocol: ProtocolMqtt, buffer: make(chan subscribe, 100), qos: qos}
	if err = newEmq(mqttiLink.pluginId, mqttiLink.cli.(emqx.Client), mqttiLink.qos); err != nil {
		return nil
	}

	//监听下发指令
	mq.commandsSubscribe(mqttiLink.buffer)

	//定时心跳
	t := time.NewTicker(30 * time.Second)

	go func() {
		for {
			<-t.C
			mq.heartBeat()
		}
	}()

	return mqttiLink
}

func NewTcpILink() *ILink {
	return nil
}
