package ilink

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	emqx "github.com/eclipse/paho.mqtt.golang"
	"github.com/tidwall/gjson"
	"reflect"
	"time"
)

type emq struct {
	pluginId string
	client   emqx.Client
	qos      byte
}

type ChannelStatus uint8
type TagsQuality uint8

const (
	ChannelStatusStopped = ChannelStatus(0)
	ChannelStatusRunning = ChannelStatus(1)
	ChannelStatusError   = ChannelStatus(2)

	Quality = TagsQuality(0)
)

func newMqtt(addr string, auth ...string) (emqx.Client, error) {
	ops := emqx.NewClientOptions().AddBroker("tcp://" + addr).
		SetTLSConfig(&tls.Config{InsecureSkipVerify: true}).
		SetAutoReconnect(true).SetProtocolVersion(4).
		SetCleanSession(true)

	if auth != nil && len(auth) == 2 {
		ops.SetUsername(auth[0])
		ops.SetPassword(auth[1])
	}

	client := emqx.NewClient(ops)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

type baseReq struct {
	Operate   command     `json:"operate"`
	OperateId int         `json:"operateId"`
	Version   string      `json:"version"`
	Data      interface{} `json:"data"`
}

type baseRes struct {
	Operate   command `json:"operate"`
	OperateId int     `json:"operateId"`
	Code      int     `json:"code"`
}

type baseResWithData struct {
	Operate   command     `json:"operate"`
	OperateId int         `json:"operateId"`
	Code      int         `json:"code"`
	Data      interface{} `json:"data"`
}

func (e emq) commandsSubscribe(c chan subscribe) {
	e.client.Subscribe(downTopic+"/"+e.pluginId, e.qos, func(client emqx.Client, message emqx.Message) {
		operate := gjson.Get(string(message.Payload()), "operate").String()
		sub := subscribe{
			Operate: command(operate),
			Body:    message.Payload(),
		}
		c <- sub
	})

}

func (e emq) heartBeat() error {
	h := baseReq{
		Operate:   CmdHeartBeat,
		OperateId: cmd[CmdHeartBeat],
		Version:   version,
		Data: map[string]int64{
			"time":      time.Now().Unix(),
			"runStatus": 2,
		},
	}
	payload, err := json.Marshal(h)
	if err != nil {
		return err
	}
	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, payload); token.Wait() && token.Error() != nil {
		return err
	}

	return nil
}

func (e emq) connect(ver string) error {
	if ver == "" {
		return fmt.Errorf("协议组件版本参数不能为空")
	}

	connect := baseReq{
		Operate:   CmdConnect,
		OperateId: cmd[CmdConnect],
		Version:   version,
		Data: map[string]string{
			"pid":     "",
			"version": ver,
		},
	}
	body, err := json.Marshal(connect)
	if err != nil {
		return err
	}
	fmt.Println("body:", string(body))
	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, body); token.Wait() && token.Error() != nil {
		return err
	}
	return nil
}

func (e emq) syncChannelTagStart() error {
	type req struct {
		Operate   command `json:"operate"`
		OperateId int     `json:"operateId"`
		Version   string  `json:"version"`
	}

	r := req{
		Operate:   CmdSyncChannelTagStart,
		OperateId: cmd[CmdSyncChannelTagStart],
		Version:   version,
	}
	body, err := json.Marshal(r)
	if err != nil {
		return err
	}

	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, body); token.Wait() && token.Error() != nil {
		return err
	}

	return nil
}

func (e emq) syncChannelTagEndResponse() error {
	base := baseRes{
		Operate:   CmdSyncChannelTagEndRes,
		OperateId: cmd[CmdSyncChannelTagEndRes],
		Code:      0,
	}

	body, err := json.Marshal(base)
	if err != nil {
		return err
	}

	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, body); token.Wait() && token.Error() != nil {
		return err
	}

	return nil
}

func (e emq) deleteChannelRes() error {
	base := baseRes{
		Operate:   CmdDelAllChannelRes,
		OperateId: cmd[CmdDelAllChannelRes],
		Code:      0,
	}
	body, err := json.Marshal(base)
	if err != nil {
		return err
	}

	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, body); token.Wait() && token.Error() != nil {
		return err
	}
	return nil
}

func (e emq) deleteAllChannelRes() error {
	res := baseRes{
		Operate:   CmdDelAllChannelRes,
		OperateId: cmd[CmdDelAllChannelRes],
		Code:      0,
	}
	body, err := json.Marshal(res)
	if err != nil {
		return err
	}

	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, body); token.Wait() && token.Error() != nil {
		return err
	}
	return nil
}

func (e emq) getChannelStatusRes(channelId string, stat ChannelStatus) error {
	if channelId == "" || reflect.ValueOf(stat).IsNil() {
		return fmt.Errorf("channelId或status参数不能为空")
	}

	type channels struct {
		Id     string        `json:"id"`
		Status ChannelStatus `json:"channelStatus"`
	}

	type status struct {
		Operate   command    `json:"operate"`
		OperateId int        `json:"operateId"`
		Code      int        `json:"code"`
		Data      []channels `json:"data"`
	}

	chans := []channels{
		{
			Id:     channelId,
			Status: stat,
		},
	}
	s := status{
		Operate:   CmdGetChannelStatusRes,
		OperateId: cmd[CmdGetChannelStatusRes],
		Code:      0,
		Data:      chans,
	}

	payload, err := json.Marshal(s)

	if err != nil {
		return err
	}

	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, payload); token.Wait() && token.Error() != nil {
		return err
	}
	return nil
}

type Tags struct {
	Id   string `json:"tags"`
	Code int    `json:"code"`
}

func (e emq) tagWriteRes(channelId string) error {
	var tags []Tags
	tags = append(tags, Tags{
		Id: channelId,
	})
	base := baseResWithData{
		Operate:   CmdTagReadRes,
		OperateId: cmd[CmdTagReadRes],
		Code:      0,
		Data:      tags,
	}

	payload, err := json.Marshal(base)
	if err != nil {
		return err
	}

	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, payload); token.Wait() && token.Error() != nil {
		return err
	}

	return nil
}

type tagValue struct {
	Id   string `json:"id"`
	V    string `json:"v"`
	Q    byte   `json:"Q"`
	Code int    `json:"code"`
}

func (e emq) tagReadRes(channelId string, value string, quality byte) error {
	var tags []tagValue
	tags = append(tags, tagValue{
		Id:   channelId,
		V:    value,
		Q:    quality,
		Code: 0,
	})
	base := baseResWithData{
		Operate:   CmdTagReadRes,
		OperateId: cmd[CmdTagReadRes],
		Code:      0,
		Data:      tags,
	}

	payload, err := json.Marshal(base)
	if err != nil {
		return err
	}

	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, payload); token.Wait() && token.Error() != nil {
		return err
	}

	return nil
}

func (e emq) channelStatusUp(channelId string, status ChannelStatus) error {

	type Status struct {
		Id         string        `json:"id"`
		Status     ChannelStatus `json:"status"`
		StartCount int           `json:"startCount"`
		StartTime  int64         `json:"startTime"`
	}
	var tags []Status
	tags = append(tags, Status{
		Id:        channelId,
		Status:    status,
		StartTime: time.Now().Unix(),
	})

	base := baseResWithData{
		Operate:   CmdChannelStatusUp,
		OperateId: cmd[CmdChannelStatusUp],
		Code:      0,
		Data: map[string][]Status{
			"channels": tags,
		},
	}

	payload, err := json.Marshal(base)
	if err != nil {
		return err
	}

	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, payload); token.Wait() && token.Error() != nil {
		return err
	}

	return nil
}

func (e emq) tagUp(channelId string, value string, quality byte) error {
	type TagValue struct {
		Id string `json:"id"`
		V  string `json:"v"`
		Q  byte   `json:"Q"`
		Ts int64  `json:"ts"`
	}

	var tags []TagValue
	tags = append(tags, TagValue{
		Id: channelId,
		V:  value,
		Q:  quality,
		Ts: time.Now().Unix(),
	})
	base := baseResWithData{
		Operate:   CmdTagUp,
		OperateId: cmd[CmdTagUp],
		Code:      0,
		Data:      tags,
	}

	payload, err := json.Marshal(base)
	if err != nil {
		return err
	}

	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, payload); token.Wait() && token.Error() != nil {
		return err
	}

	return nil
}