package ilink

type command string

const (
	CmdHeartBeat              = command("HEART_BEAT")
	CmdHeartBeatRes           = command("HEART_BEAT_RES")
	CmdConnect                = command("CONNECT")
	CmdConnectACK             = command("CONNACK")
	CmdSyncChannelTagStart    = command("SYNC_CHANNEL_TAG_START")
	CmdSyncChannelTagStartRes = command("SYNC_CHANNEL_TAG_START_RES")
	CmdSyncChannelTag         = command("SYNC_CHANNEL_TAG")
	CmdSyncChannelTagRes      = command("SYNC_CHANNEL_TAG_RES")
	CmdSyncChannelTagEnd      = command("SYNC_CHANNEL_TAG_END")
	CmdSyncChannelTagEndRes   = command("SYNC_CHANNEL_TAG_END_RES")
	CmdDelChannel             = command("DEL_CHANNEL")
	CmdDelChannelRes          = command("DEL_CHANNEL_RES")
	CmdDelAllChannel          = command("DEL_ALL_CHANNEL")
	CmdDelAllChannelRes       = command("DEL_ALL_CHANNEL_RES")
	CmdGetChannelStatus       = command("GET_CHANNEL_STATUS")
	CmdGetChannelStatusRes    = command("GET_CHANNEL_STATUS_RES")
	CmdChannelStatusUp        = command("CHANNEL_STATUS_UP")
	CmdTagUp                  = command("TAG_UP")
	CmdTagUpRes               = command("TAG_UP_RES")
	CmdTagWrite               = command("TAG_WRITE")
	CmdTagWriteRes            = command("TAG_WRITE_RES")
	CmdTagRead                = command("TAG_READ")
	CmdTagReadRes             = command("TAG_READ_RES")
	CmdAttrRead               = command("ATTR_READ")
	CmdAttrReadRes            = command("ATTR_READ_RES")
	CmdAttrWrite              = command("ATTR_WRITE")
	CmdAttrWriteRes           = command("ATTR_WRITE_RES")
	CmdAttrUp                 = command("ATTR_UP")
	CmdAttrUpRes              = command("ATTR_UP_RES")
	CmdServiceDown            = command("SERVICE_DOWN")
	CmdServiceDownRes         = command("SERVICE_DOWN_RES")
	CmdEventUp                = command("EVENT_UP")
	CmdEventUpRes             = command("EVENT_UP_RES")
	CmdDevStatusUp            = command("DEV_STATUS_UP")
	CmdDevStatusUpRes         = command("DEV_STATUS_UP_RES")
	CmdGetDevStatus           = command("GET_DEV_STATUS")
	CmdGetDevStatusRes        = command("GET_DEV_STATUS_RES")
	CmdDriverStatusUp         = command("DRIVER_STATUS_UP")
	CmdPing                   = command("PING")
	CmdPong                   = command("PONG")
	CmdFetchSyncId            = command("FETCH_SYNC_ID")
	CmdFetchSyncIdRes         = command("FETCH_SYNC_ID")
	CmdSyncInterrupt          = command("SYNC_INTERRUPT")
	CmdSyncAsk                = command("SYNC_ASK")
	CmdSyncReady              = command("SYNC_READY")
	CmdSyncStart              = command("SYNC_START")
	CmdSyncAck                = command("SYNC_ACK")
	CmdSyncIng                = command("SYNC_ING")
	CmdSyncRecv               = command("SYNC_RECV")
	CmdSyncConflict           = command("SYNC_CONFLICT")
	CmdSyncConfirm            = command("SYNC_CONFIRM")
	CmdSyncEnd                = command("SYNC_END")
	CmdSyncOk                 = command("SYNC_OK")
)

var cmd map[command]int64

func init() {
	cmd = map[command]int64{
		"HEART_BEAT":                 1,
		"HEART_BEAT_RES":             2,
		"CONNECT":                    3,
		"CONNACK":                    4,
		"SYNC_CHANNEL_TAG_START":     7,
		"SYNC_CHANNEL_TAG_START_RES": 8,
		"SYNC_CHANNEL_TAG":           9,
		"SYNC_CHANNEL_TAG_RES":       10,
		"SYNC_CHANNEL_TAG_END":       11,
		"SYNC_CHANNEL_TAG_END_RES":   12,
		"DEL_CHANNEL":                13,
		"DEL_CHANNEL_RES":            14,
		"DEL_ALL_CHANNEL":            15,
		"DEL_ALL_CHANNEL_RES":        16,
		"GET_CHANNEL_STATUS":         17,
		"GET_CHANNEL_STATUS_RES":     18,
		"CHANNEL_STATUS_UP":          19,
		"TAG_UP":                     23,
		"TAG_UP_RES":                 24,
		"TAG_WRITE":                  25,
		"TAG_WRITE_RES":              26,
		"TAG_READ":                   27,
		"TAG_READ_RES":               28,
		"ATTR_READ":                  33,
		"ATTR_READ_RES":              34,
		"ATTR_WRITE":                 35,
		"ATTR_WRITE_RES":             36,
		"ATTR_UP":                    37,
		"ATTR_UP_RES":                38,
		"SERVICE_DOWN":               39,
		"SERVICE_DOWN_RES":           40,
		"EVENT_UP":                   41,
		"EVENT_UP_RES":               42,
		"DEV_STATUS_UP":              43,
		"DEV_STATUS_UP_RES":          44,
		"GET_DEV_STATUS":             45,
		"GET_DEV_STATUS_RES":         46,
		"DRIVER_STATUS_UP":           47,
		"PING":                       51,
		"PONG":                       52,
		"FETCH_SYNC_ID":              53,
		"FETCH_SYNC_ID_RES":          54,
		"SYNC_INTERRUPT":             55,
		"SYNC_ASK":                   57,
		"SYNC_READY":                 59,
		"SYNC_START":                 61,
		"SYNC_ACK":                   63,
		"SYNC_ING":                   65,
		"SYNC_RECV":                  67,
		"SYNC_CONFLICT":              69,
		"SYNC_CONFIRM":               71,
		"SYNC_END":                   73,
		"SYNC_OK":                    75,
	}
}
