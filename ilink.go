package ilink

import (
	"encoding/json"
	emqx "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
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

//ChannelTagConfig 网关下发的通道和点位配置
type ChannelTagConfig struct {
	Operate   string
	OperateId int64
	Version   string
	Channel
	Tags []Tag
}

type Channel struct {
	ChannelId     string
	ChannelName   string
	Timeout       int64
	TimeWait      int64
	Type          string
	ChannelConfig interface{}
}

type Tag struct {
	TagId      string
	TagPeriod  int64
	DataType   string
	Rw         int64
	AreaType   string
	AreaParams string
}

type Communication interface {
	Connect(ver string) error                                               //插件连接网关
	SyncChannelTagStart() error                                             //同步CHANNEL和TAG通知
	SyncChannelTagEndResponse() error                                       //同步结束响应
	DeleteChannelResponse() error                                           //删除单个通道响应
	AddCallbackSyncChannelTag()                                             //新增同步通道和点位回调函数
	AddCallbackDelChannel()                                                 //新增删除单个通道回调函数
	AddCallbackDelAllChannel()                                              //新增删除所有通道回调函数
	AddCallbackTagRead()                                                    //新增点位读回调函数
	AddCallbackTagWrite()                                                   //新增点位写回调函数
	DeleteAllChannelResponse() error                                        //删除所有通道响应
	GetChannelStatusRes(channelId string, status ChannelStatus) error       //获取通道状态
	ChannelStatusUp() error                                                 //通道状态上报
	TagWriteResp() error                                                    //点位写入
	TagReadResp() error                                                     //点位读取
	TagUp(channelId string, tagId string, value string, quality byte) error //点位上报

}

func (ilink ILink) AddCallbackSyncChannelTag(f func(tagConfig ChannelTagConfig, up chan<- TagUp)) {
	if ilink.cli == nil {
		return
	}

	if ilink.protocol == ProtocolMqtt {
		mq.addCallbackSyncChannelTag(f)
	}
}

func (ilink ILink) AddCallbackDelChannel(f func()) {
	if ilink.cli == nil {
		return
	}

	if ilink.protocol == ProtocolMqtt {
		mq.addCallbackDelChannel(f)
	}
}

func (ilink ILink) AddCallbackDelAllChannel(f func()) {
	if ilink.cli == nil {
		return
	}

	if ilink.protocol == ProtocolMqtt {
		mq.addCallbackDelAllChannel(f)
	}
}

func (ilink ILink) AddCallbackTagRead(f func()) {
	if ilink.cli == nil {
		return
	}

	if ilink.protocol == ProtocolMqtt {
		mq.addCallbackTagRead(f)
	}
}

func (ilink ILink) AddCallbackTagWrite(f func()) {
	if ilink.cli == nil {
		return
	}
	if ilink.protocol == ProtocolMqtt {
		mq.addCallbackTagWrite(f)
	}
}

func (ilink ILink) Subscribe() (c <-chan subscribe) {
	return ilink.buffer
}

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

func NewMqtt(configPath string) (*ILink, error) {
	if mqttiLink != nil {
		return mqttiLink, nil
	}

	if err := getConfigure(configPath); err != nil {
		return nil, errors.Wrap(err, "getConfigure()")
	}
	c, err := newMqtt(cfg.MqttHost, cfg.MqttPort)
	if err != nil {
		return nil, errors.Wrap(err, "newMqtt()")
	}
	mqttiLink = &ILink{operateId: 0, pluginId: cfg.ModuleId, cli: c, protocol: ProtocolMqtt, buffer: make(chan subscribe, 100), qos: cfg.MqttQos}
	if err = newEmq(mqttiLink.pluginId, mqttiLink.cli.(emqx.Client), mqttiLink.qos); err != nil {
		return nil, errors.Wrap(err, "newEmq()")
	}

	//监听下发指令
	mq.commandsSubscribe()

	//上报点位采集数据
	go mq.up()

	//定时心跳
	t := time.NewTicker(30 * time.Second)

	go func() {
		for {
			<-t.C
			mq.heartBeat()
		}
	}()

	return mqttiLink, nil
}

type config struct {
	MqttHost        string `json:"mqttHost"`
	MqttPort        int    `json:"mqttPort"`
	MqttQos         byte   `json:"mqttQos"`
	ModuleId        string `json:"moduleId"`
	ConnectTimeout  int    `json:"connectTimeout"`
	HeartbeatPeriod int    `json:"heartbeatPeriod"`
	RWTimeout       int    `json:"rwTimeout"`
}

var cfg = &config{}

func getConfigure(filePath string) error {
	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		return errors.Wrap(err, "open config.json failed")
	}

	body, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.Wrap(err, "read config.json failed")
	}
	if err := json.Unmarshal(body, cfg); err != nil {
		return errors.Wrap(err, "json.Unmarshal")
	}
	return err

}

func NewTcpILink() *ILink {
	return nil
}
