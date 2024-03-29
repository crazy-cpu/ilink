package ilink

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/crazy-cpu/ilink/messagebus/mqtt"
	emqx "github.com/eclipse/paho.mqtt.golang"

	//emqx "github.com/eclipse/paho.mqtt.golang"
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

var channelCount = make(map[string]int) //记录通道的启动次数

type emq struct {
	operateId              int64
	pluginId               string
	client                 *mqtt.Client
	qos                    byte
	connAck                chan int64
	callbackSyncChannelTag func(cfg ChannelTagConfig, up chan<- TagUp) //需调用方在每次采集上报时将数据传入Up通道
	callbackDelChannel     func(channelId string)
	callbackDelAllChannel  func(channelId string)
	callbackTagWrite       func()
	callbackTagRead        func()
	upQueue                chan TagUp
}

type TagUp struct {
	ChannelId string
	TagId     string
	Value     string
	Quality   byte
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

func newMqtt(ip string, port int, auth ...string) (mqtt.Client, error) {
	var userName, password string
	if auth[0] != "" && auth[1] != "" {
		userName = auth[0]
		password = auth[1]
	}
	p := strconv.Itoa(port)
	C := mqtt.Client{
		Server:   "tcp://" + ip + ":" + p,
		UserName: userName,
		Password: password,
	}
	C.Connect()
	return C, nil
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

func newEmq(pluginId string, client *mqtt.Client, qos byte) error {
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
		upQueue:   make(chan TagUp, 1024),
	}

	return nil
}

func (e *emq) addCallbackSyncChannelTag(f func(tagConfig ChannelTagConfig, up chan<- TagUp)) {
	e.callbackSyncChannelTag = f
}

func (e *emq) addCallbackDelChannel(f func(channelId string)) {
	e.callbackDelChannel = f
}

func (e *emq) addCallbackDelAllChannel(f func(channelId string)) {
	e.callbackDelAllChannel = f
}

func (e *emq) addCallbackTagWrite(f func()) {
	e.callbackTagWrite = f
}

func (e *emq) addCallbackTagRead(f func()) {
	e.callbackTagRead = f
}

func getChannelTagConfig(b []byte) ChannelTagConfig {
	body := string(b)
	cfg := ChannelTagConfig{}

	cfg.ChannelConfig = make(map[string]string)
	cfg.Operate = gjson.Get(body, "operate").String()
	cfg.OperateId = gjson.Get(body, "operateId").Int()
	cfg.Version = gjson.Get(body, "version").String()

	cfg.ChannelId = gjson.Get(body, "data.channel.channelId").String()
	cfg.ChannelName = gjson.Get(body, "data.channel.channelName").String()
	cfg.Timeout = gjson.Get(body, "data.channel.timeout").Int()
	cfg.TimeWait = gjson.Get(body, "data.channel.timeWait").Int()
	cfg.Type = gjson.Get(body, "data.channel.type").String()

	channelCfg := gjson.Get(body, "data.channel.channelConfig").Map()
	for k, v := range channelCfg {
		cfg.ChannelConfig[k] = v.String()
	}

	tags := gjson.Get(body, "data.tags").Array()
	for _, t := range tags {
		cfg.Tags = append(cfg.Tags, Tag{
			TagId:      gjson.Get(t.String(), "tagId").String(),
			TagPeriod:  gjson.Get(t.String(), "tagPeriod").Int(),
			DataType:   gjson.Get(t.String(), "dataType").String(),
			Rw:         gjson.Get(t.String(), "rw").Int(),
			AreaType:   gjson.Get(t.String(), "areaType").String(),
			AreaParams: gjson.Get(t.String(), "areaParams").String(),
		})
	}

	return cfg

}

func (e *emq) up() {
	for {
		body := <-e.upQueue
		e.tagUp(body.ChannelId, body.TagId, body.Value, body.Quality)
	}
}
func (e *emq) commandsSubscribe() {
	subArgs := mqtt.SubArg{
		SubTopic: downTopic + "/" + e.pluginId,
		SubQos:   e.qos,
		SubCallback: func(client emqx.Client, message emqx.Message) {
			operate := gjson.Get(string(message.Payload()), "operate").String()
			operateId := gjson.Get(string(message.Payload()), "operateId").Int()
			switch command(operate) {
			case CmdSyncChannelTagStart:
				e.syncChannelTagStartResp(operateId)
			case CmdSyncChannelTag:
				//将通道和点位配置转换成合适的格式
				Config := getChannelTagConfig(message.Payload())
				go e.callbackSyncChannelTag(Config, e.upQueue)
				e.syncChannelTagRes(operateId)
			case CmdSyncChannelTagEnd:
				e.syncChannelTagEndResponse(operateId)
				channelId := gjson.Get(string(message.Payload()), "data.channelId").String()
				channelCount[channelId] += 1
				e.channelStatusUp(channelId, 1)
			case CmdConnectACK:
				e.connAck <- operateId
			case CmdDelChannel:
				channelId := gjson.Get(string(message.Payload()), "data.channelId").String()
				e.callbackDelChannel(channelId)
				e.deleteChannelRes(operateId)
				e.channelStatusUp(channelId, 0)
			case CmdDelAllChannel:
				channelId := gjson.Get(string(message.Payload()), "data.channelId").String()
				e.callbackDelAllChannel(channelId)
				e.deleteAllChannelRes(operateId)
			case CmdTagWrite:
				e.callbackTagWrite()
			case CmdTagRead:
				e.callbackTagRead()
			}
		},
	}
	e.client.Subscribe(subArgs)
	//e.client.Subscribe(downTopic+"/"+e.pluginId, e.qos, func(client emqx.Client, message emqx.Message) {
	//	operate := gjson.Get(string(message.Payload()), "operate").String()
	//	operateId := gjson.Get(string(message.Payload()), "operateId").Int()
	//	switch command(operate) {
	//	case CmdSyncChannelTagStart:
	//		e.syncChannelTagStartResp(operateId)
	//	case CmdSyncChannelTag:
	//		//将通道和点位配置转换成合适的格式
	//		Config := getChannelTagConfig(message.Payload())
	//		go e.callbackSyncChannelTag(Config, e.upQueue)
	//		e.syncChannelTagRes(operateId)
	//	case CmdSyncChannelTagEnd:
	//		e.syncChannelTagEndResponse(operateId)
	//		channelId := gjson.Get(string(message.Payload()), "data.channelId").String()
	//		channelCount[channelId] += 1
	//		e.channelStatusUp(channelId, 1)
	//	case CmdConnectACK:
	//		e.connAck <- operateId
	//	case CmdDelChannel:
	//		channelId := gjson.Get(string(message.Payload()), "data.channelId").String()
	//		e.callbackDelChannel(channelId)
	//		e.deleteChannelRes(operateId)
	//		e.channelStatusUp(channelId, 0)
	//	case CmdDelAllChannel:
	//		channelId := gjson.Get(string(message.Payload()), "data.channelId").String()
	//		e.callbackDelAllChannel(channelId)
	//		e.deleteAllChannelRes(operateId)
	//	case CmdTagWrite:
	//		e.callbackTagWrite()
	//	case CmdTagRead:
	//		e.callbackTagRead()
	//	}
	//
	//})

}

func (e *emq) heartBeat() error {
	var err error
	h := baseReq{
		Operate:   CmdHeartBeat,
		OperateId: atomic.AddInt64(&e.operateId, 1),
		Version:   version,
		Data: map[string]int64{
			"time":      time.Now().UnixMilli(),
			"runStatus": 2,
		},
	}
	payload, err := json.Marshal(h)
	if err != nil {
		return err
	}
	pubArgs := mqtt.PubArg{
		PubQos:      e.qos,
		PubTopic:    upTopic + "/" + e.pluginId,
		PubRetained: false,
		PubPayload:  payload,
	}
	if err = e.client.Publish(pubArgs); err != nil {
		return err
	}

	return nil
}

//connect 带响应超时，默认等待5S
func (e *emq) connect(ver string) error {
	var err error
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

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: body}); err != nil {
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

func (e *emq) syncChannelTagStart() error {
	var err error
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

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: body}); err != nil {
		return err
	}

	return nil
}

func (e *emq) syncChannelTagStartResp(operateId int64) error {
	var err error
	res := baseRes{
		Operate:   CmdSyncChannelTagStartRes,
		OperateId: operateId,
	}
	body, err := json.Marshal(res)
	if err != nil {
		return err
	}

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: body}); err != nil {
		return err
	}
	return nil
}

func (e *emq) syncChannelTagEndResponse(operateId int64) error {
	var err error
	base := baseRes{
		Operate:   CmdSyncChannelTagEndRes,
		OperateId: operateId,
		Code:      0,
	}

	body, err := json.Marshal(base)
	if err != nil {
		return err
	}

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: body}); err != nil {
		return err
	}

	return nil
}

func (e *emq) syncChannelTagRes(operateId int64) error {
	var err error
	base := baseRes{
		Operate:   CmdSyncChannelTagRes,
		OperateId: operateId,
		Code:      0,
	}

	body, err := json.Marshal(base)
	if err != nil {
		return err
	}

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: body}); err != nil {
		return err
	}

	return nil
}

func (e *emq) deleteChannelRes(operateId int64) error {
	var err error
	base := baseRes{
		Operate:   CmdDelChannelRes,
		OperateId: operateId,
		Code:      0,
	}
	body, err := json.Marshal(base)
	if err != nil {
		return err
	}

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: body}); err != nil {
		return err
	}
	return nil
}

func (e *emq) deleteAllChannelRes(operateId int64) error {
	var err error
	res := baseRes{
		Operate:   CmdDelAllChannelRes,
		OperateId: operateId,
		Code:      0,
	}
	body, err := json.Marshal(res)
	if err != nil {
		return err
	}

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: body}); err != nil {
		return err
	}
	return nil
}

func (e *emq) getChannelStatusRes(channelId string, stat ChannelStatus) error {
	var err error
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

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: payload}); err != nil {
		return err
	}
	return nil
}

type Tags struct {
	Id   string `json:"tags"`
	Code int    `json:"code"`
}

func (e *emq) tagWriteRes(channelId string) error {
	var (
		err  error
		tags []Tags
	)
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

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: payload}); err != nil {
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

func (e *emq) tagReadRes(channelId string, value string, quality byte) error {
	var (
		err  error
		tags []tagValue
	)
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

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: payload}); err != nil {
		return err
	}

	return nil
}

func (e *emq) channelStatusUp(channelId string, status ChannelStatus) error {

	type Status struct {
		Id         string        `json:"id"`
		Status     ChannelStatus `json:"status"`
		StartCount int           `json:"startCount"`
		StartTime  int64         `json:"startTime"`
	}
	var (
		err  error
		tags []Status
	)
	tags = append(tags, Status{
		Id:         channelId,
		Status:     status,
		StartCount: channelCount[channelId],
		StartTime:  time.Now().UnixMilli(),
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

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: payload}); err != nil {
		return err
	}

	return nil
}

func (e *emq) tagUp(channelId string, TagId string, value string, quality byte) error {
	var err error
	type TagValue struct {
		Id string `json:"id"`
		V  string `json:"v"`
		Q  byte   `json:"q"`
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
					Ts: time.Now().UnixMilli(),
				},
			},
		},
	}
	payload, err := json.Marshal(res)
	if err != nil {
		return err
	}

	if err = e.client.Publish(mqtt.PubArg{PubTopic: upTopic + "/" + e.pluginId, PubQos: e.qos, PubRetained: false, PubPayload: payload}); err != nil {
		return err
	}

	return nil
}
