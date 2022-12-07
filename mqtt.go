package ilink

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	emqx "github.com/eclipse/paho.mqtt.golang"
	"github.com/tidwall/gjson"
	"os"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	timeoutResp = 5
)

var (
	ErrOperateIdNotMatched    = errors.New("operateId不匹配")
	ErrTimeoutConnectResponse = errors.New("等待[CONNACK]超时")
	ErrPluginIdOrClientIsNull = errors.New("pluginId或client参数不能为空")
)

type emq struct {
	operateId int64
	pluginId  string
	client    emqx.Client
	qos       byte
	connAck   chan int64
}

type ChannelStatus uint8
type TagsQuality uint8

var (
	mq *emq
)

const (
	ChannelStatusStopped = ChannelStatus(0)
	ChannelStatusRunning = ChannelStatus(1)
	ChannelStatusError   = ChannelStatus(2)

	Quality = TagsQuality(0)
)

func newMqtt(ip string, port int, auth ...string) (emqx.Client, error) {
	p := strconv.Itoa(port)
	ops := emqx.NewClientOptions().AddBroker("tcp://" + ip + ":" + p).
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
	OperateId int64       `json:"operateId"`
	Version   string      `json:"version"`
	Data      interface{} `json:"data"`
}

type baseRes struct {
	Operate   command `json:"operate"`
	OperateId int64   `json:"operateId"`
	Code      int     `json:"code"`
}

type baseResWithData struct {
	Operate   command     `json:"operate"`
	OperateId int64       `json:"operateId"`
	Code      int         `json:"code"`
	Data      interface{} `json:"data"`
}

func newEmq(pluginId string, client emqx.Client, qos byte) error {
	if pluginId == "" || client == nil {
		return nil
	}

	if mq != nil {
		return nil
	}

	mq = &emq{
		operateId: 0,
		pluginId:  pluginId,
		client:    client,
		qos:       qos,
		connAck:   make(chan int64, 1),
	}

	return nil
}

func (e emq) commandsSubscribe(c chan subscribe) {
	e.client.Subscribe(downTopic+"/"+e.pluginId, e.qos, func(client emqx.Client, message emqx.Message) {
		operate := gjson.Get(string(message.Payload()), "operate").String()
		operateId := gjson.Get(string(message.Payload()), "operateId").Int()
		switch command(operate) {
		case CmdSyncChannelTagStart:
			e.syncChannelTagStartResp(operateId)
		case CmdSyncChannelTagEnd:
			e.syncChannelTagEndResponse(operateId)
		case CmdConnectACK:
			e.connAck <- operateId
		case CmdDelChannel, CmdDelAllChannel, CmdTagWrite, CmdTagRead:
			sub := subscribe{
				Operate:   command(operate),
				OperateId: operateId,
				Body:      message.Payload(),
			}
			c <- sub
		}

	})

}

func (e emq) heartBeat() error {
	h := baseReq{
		Operate:   CmdHeartBeat,
		OperateId: atomic.AddInt64(&e.operateId, 1),
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

//connect 带响应超时，默认等待5S
func (e emq) connect(ver string) error {
	if ver == "" {
		return fmt.Errorf("协议组件版本参数不能为空")
	}
	opId := atomic.AddInt64(&e.operateId, 1)

	connect := baseReq{
		Operate:   CmdConnect,
		OperateId: opId,
		Version:   version,
		Data: map[string]string{
			"pid":     strconv.Itoa(os.Getpid()),
			"version": ver,
		},
	}
	body, err := json.Marshal(connect)
	if err != nil {
		return err
	}
	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, body); token.Wait() && token.Error() != nil {
		return err
	}

	t := time.NewTimer(timeoutResp * time.Second)
	select {
	case Id := <-e.connAck:
		if Id != opId {
			return ErrOperateIdNotMatched
		}
		return nil
	case <-t.C:
		return ErrTimeoutConnectResponse
	}
	return nil
}

func (e emq) syncChannelTagStart() error {
	type req struct {
		Operate   command `json:"operate"`
		OperateId int64   `json:"operateId"`
		Version   string  `json:"version"`
	}

	r := req{
		Operate:   CmdSyncChannelTagStart,
		OperateId: atomic.AddInt64(&e.operateId, 1),
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

func (e emq) syncChannelTagStartResp(operateId int64) error {
	res := baseRes{
		Operate:   CmdSyncChannelTagStartRes,
		OperateId: operateId,
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

func (e emq) syncChannelTagEndResponse(operateId int64) error {
	base := baseRes{
		Operate:   CmdSyncChannelTagEndRes,
		OperateId: operateId,
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

func (e emq) syncChannelTagRes(operateId int64) error {
	base := baseRes{
		Operate:   CmdSyncChannelTagRes,
		OperateId: operateId,
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

func (e emq) deleteChannelRes(operateId int64) error {
	base := baseRes{
		Operate:   CmdDelChannelRes,
		OperateId: operateId,
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

func (e emq) deleteAllChannelRes(operateId int64) error {
	res := baseRes{
		Operate:   CmdDelAllChannelRes,
		OperateId: operateId,
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
		OperateId int64      `json:"operateId"`
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
		OperateId: atomic.AddInt64(&e.operateId, 1),
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
		OperateId: atomic.AddInt64(&e.operateId, 1),
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
		OperateId: atomic.AddInt64(&e.operateId, 1),
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
		OperateId: atomic.AddInt64(&e.operateId, 1),
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

func (e emq) tagUp(channelId string, TagId string, value string, quality byte) error {
	type TagValue struct {
		Id string `json:"id"`
		V  string `json:"v"`
		Q  byte   `json:"Q"`
		Ts int64  `json:"ts"`
	}

	type Data struct {
		ChannelId string     `json:"channelId"`
		Tags      []TagValue `json:"tags"`
	}

	type Res struct {
		Operate   command `json:"operate"`
		OperateId int64   `json:"operateId"`
		Version   string  `json:"version"`
		Data      Data    `json:"data"`
	}

	res := Res{
		Operate:   CmdTagUp,
		OperateId: atomic.AddInt64(&e.operateId, 1),
		Version:   "",
		Data: Data{
			ChannelId: channelId,
			Tags: []TagValue{
				{
					Id: TagId,
					V:  value,
					Q:  quality,
					Ts: time.Now().Unix(),
				},
			},
		},
	}
	payload, err := json.Marshal(res)
	if err != nil {
		return err
	}

	if token := e.client.Publish(upTopic+"/"+e.pluginId, e.qos, false, payload); token.Wait() && token.Error() != nil {
		return err
	}

	return nil
}
