package tplink

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Action int

const (
	OFF Action = iota
	ON
)

const (
	DISABLED = iota
	ENABLED
)

type TimeOption int

const (
	NONE TimeOption = iota
	SUNRISE
	SUNSET
)

type Days struct {
	Sunday    bool
	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
	Saturday  bool
}

func (d Days) String() string {
	days := []string{"0", "0", "0", "0", "0", "0", "0"}
	if d.Sunday {
		days[time.Sunday] = "1"
	}

	if d.Monday {
		days[time.Monday] = "1"
	}

	if d.Tuesday {
		days[time.Tuesday] = "1"
	}

	if d.Wednesday {
		days[time.Wednesday] = "1"
	}

	if d.Thursday {
		days[time.Thursday] = "1"
	}

	if d.Friday {
		days[time.Friday] = "1"
	}

	if d.Saturday {
		days[time.Saturday] = "1"
	}
	return fmt.Sprintf("[%s]", strings.Join(days, ","))
}

const (
	// --- Plug HS100 and HS110 ---

	// System Commands
	GET_INFO     = `{"system":{"get_sysinfo":{}}}`
	REBOOT       = `{"system":{"reboot":{"delay":1}}}`
	RESET        = `{"system":{"reset":{"delay":1}}}`
	SET_ALIAS    = `{"system":{"set_dev_alias":{"alias":"%s"}}}`
	TURN_LED_ON  = `{"system":{"set_led_off":{"off":0}}}`
	TURN_LED_OFF = `{"system":{"set_led_off":{"off":1}}}`
	TURN_ON      = `{"system":{"set_relay_state":{"state":1}}}`
	TURN_OFF     = `{"system":{"set_relay_state":{"state":0}}}`
	// WLAN Commands
	SCAN_WIFI = `{"netif":{"get_scaninfo":{"refresh":1}}}`
	SET_WIFI  = `{"netif":{"set_stainfo":{"ssid":"%s","password":"%s","key_type":%d}}}`
	// Cloud Commands
	GET_CLOUD_INFO = `{"cnCloud":{"get_info":null}}`
	SET_CLOUD_URL  = `{"cnCloud":{"set_server_url":{"server":"%s"}}}`
	CLOUD_BIND     = `{"cnCloud":{"bind":{"username":"%s", "password":"%s"}}}`
	CLOUD_UNBIND   = `{"cnCloud":{"unbind":null}}`
	// Time Commands
	GET_TIME     = `{"time":{"get_time":{}}}`
	GET_TIMEZONE = `{"time":{"get_timezone":null}}`
	SET_TIMEZONE = `{"time":{"set_timezone":{"year":%d,"month":%d,"mday":%d,"hour":%d,"min":%d,"sec":%d,"index":%d}}}`
	// Schedule Commands
	GET_NEXT_SCHEDULE_ACTION = `{"schedule":{"get_next_action":null}}`
	GET_SCHEDULE_RULES_LIST  = `{"schedule":{"get_rules":null}}`
	ADD_SCHEDULE_RULE        = `{"schedule":{"add_rule":{"stime_opt":%d,"wday":%s,"smin":%d,"enable":%d,"repeat":%d,"etime_opt":-1,"name":"%s","eact":-1,"month":%d,"sact":%d,"year":%d,"longitude":0,"day":%d,"force":0,"latitude":0,"emin":0},"set_overall_enable":{"enable":1}}}`
	EDIT_SCHEDULE_RULE       = `{"schedule":{"edit_rule":{"stime_opt":%d,"wday":%s,"smin":%d,"enable":%d,"repeat":%d,"etime_opt":-1,"id":"%s","name":"%s","eact":-1,"month":%d,"sact":%d,"year":%d,"longitude":0,"day":%d,"force":0,"latitude":0,"emin":0}}}`
	DELETE_SCHEDULE_RULE     = `{"schedule":{"delete_rule":{"id":"%s"}}}`
	DELETE_ALL_SCHEDULE_RULE = `{"schedule":{"delete_all_rules":null,"erase_runtime_stat":null}}`

	//  --- HS110 only ---

	// EMeter Energy Usage Statistics Commands
	GET_METER         = `{"system":{"get_sysinfo":{}}, "emeter":{"get_realtime":{},"get_vgain_igain":{}}}`
	GET_DAILY_STATS   = `{"emeter":{"get_daystat":{"month":%d,"year":%d}}}`
	GET_MONTHLY_STATS = `{"emeter":{"get_monthstat":{"year":%d}}}`
	ERASE_ALL_STATS   = `{"emeter":{"erase_emeter_stat":null}}`
)

type Device struct {
	IPAddress string
	Info      Info
}

type Response struct {
	System struct {
		*Info    `json:"get_sysinfo"`
		SetAlias struct {
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"set_dev_alias"`
		SetState struct {
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"set_relay_state"`
	}

	CNCloud struct {
		Info struct {
			Cloud
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"get_info"`
		SetServerUrl struct {
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"set_server_url"`
		Bind struct {
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"bind"`
		Unbind struct {
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"unbind"`
	} `json:"cnCloud"`

	Time struct {
		GetTime struct {
			Year         int    `json:"year"`
			Month        int    `json:"month"`
			Day          int    `json:"mday"`
			Hour         int    `json:"hour"`
			Minutes      int    `json:"min"`
			Seconds      int    `json:"sec"`
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"get_time"`

		GetTimeZone struct {
			Index        int    `json:"index"`
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"get_timezone"`

		SetTimeZone struct {
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"set_timezone"`
	}

	Schedule struct {
		GetNextAction struct {
			NextAction
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"get_next_action"`
		Rule struct {
			List         []Rule `json:"rule_list"`
			Enable       int    `json:"enable"`
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"get_rules"`
		AddRule struct {
			ID           string `json:"id"`
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"add_rule"`
		EditRule struct {
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"edit_rule"`
		DeleteRule struct {
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"delete_rule"`
		DeleteAllRules struct {
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"delete_all_rules"`
	} `json:"schedule"`

	NetIf struct {
		GetScanInfo struct {
			List         []AP   `json:"ap_list"`
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"get_scaninfo"`

		SetWifi struct {
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"set_stainfo"`
	} `json:"netif"`

	EMeter struct {
		*Meter         `json:"get_realtime"`
		MonthlyStats   *MonthyStats `json:"get_monthstat"`
		DailyStats     *DailyStats  `json:"get_daystat"`
		EraseMeterStat struct {
			ErrorCode    int    `json:"err_code"`
			ErrorMessage string `json:"err_msg"`
		} `json:"erase_emeter_stat"`
	} `json:"emeter"`
}

type Info struct {
	SoftwareVersion string  `json:"sw_ver"`      // Software version
	HardwareVersion string  `json:"hw_ver"`      // Hardware version
	HardwareID      string  `json:"hwId"`        // Hardware ID
	Type            string  `json:"type"`        // Type
	Model           string  `json:"model"`       // Model
	MacAddr         string  `json:"mac"`         // Mac Address
	DeviceID        string  `json:"deviceId"`    // Device ID
	FirmwareID      string  `json:"fwId"`        // Firmware ID
	OEMID           string  `json:"oemId"`       // OEM ID
	Alias           string  `json:"alias"`       // Description. e.g "Basement light"
	IconHash        string  `json:"icon_hash"`   // hash for custom picture
	State           int     `json:"relay_state"` // State:  0 = OFF; 1 = ON
	ActiveMode      string  `json:"active_mode"` // "schedule" for schedule mode
	Feature         string  `json:"feature"`     // "TIM:ENE" (Timer, Energy Monitor)
	Updating        int     `json:"updating"`    // 0 = not updating
	RSSI            int     `json:"rssi"`        // Signal Strength Indicator in dBm (e.g. -35)
	LedOff          int     `json:"led_off"`     // 0 = Led ON (default); 1 = Led OFF
	Latitude        float64 `json:"latitude"`    // Optional Geolocation information
	Longitude       float64 `json:"longitude"`   // Optional Geolocation information
}

func (i Info) IsOn() bool {
	return i.State == 1
}

func (i Info) IsLedOn() bool {
	return i.LedOff == 0
}

type Cloud struct {
	Username string `json:"username"`
	Server   string `json:"server"`
	Binded   int    `json:"binded"`
}

func (c Cloud) isBinded() bool {
	return c.Binded == 1
}

type Meter struct {
	Current float64 `json:"current"`
	Voltage float64 `json:"voltage"`
	Power   float64 `json:"power"`
	Total   float64 `json:"total"`
}

type MonthyStats struct {
	MonthlyUsageList []*MonthlyUsage `json:"month_list"`
}

type MonthlyUsage struct {
	Year   int
	Month  int
	Energy float64
}

type DailyStats struct {
	DailyUsageList []*DailyUsage `json:"day_list"`
}

type DailyUsage struct {
	Year   int
	Month  int
	Day    int
	Energy float64
}

type AP struct {
	SSID    string `json:"ssid"`
	KeyType int    `json:"key_type"`
}

type Rule struct {
	Id       string     `json:"id"`
	Name     string     `json:"name"`
	Enable   int        `json:"enable"`
	Minutes  int        `json:"smin"`
	Repeat   int        `json:"repeat"`
	Action   Action     `json:"sact"`
	WeekDays []Action   `json:"wday"`
	Year     int        `json:"year"` // when repeat = 0, year, month and day will be provided
	Month    int        `json:"month"`
	Day      int        `json:"day"`
	TimeOpt  TimeOption `json:"stime_opt"` // If set, means that this rule will run on sunrise or sunset
}

type NextAction struct {
	RuleID              string `json:"id"`
	Type                int    `json:"type"` // ???
	ScheduledTimeSecond int    `json:"schd_time"`
	Action              Action `json:"action"`
}

func (r Rule) IsEnabled() bool {
	return r.Enable == 1
}

func decrypt(request []byte) string {
	result := make([]byte, len(request))
	key := byte(0xAB)
	for i, c := range request {
		var a = key ^ uint8(c)
		key = uint8(c)
		result[i] = a
	}
	return string(result)
}

func encrypt(s string) []byte {
	request := []byte(s)
	key := byte(0xAB)
	result := make([]byte, 4+len(request))
	result[0] = 0x0
	result[1] = 0x0
	result[2] = 0x0
	result[3] = 0x0
	for i, c := range request {
		var a = key ^ uint8(c)
		key = uint8(a)
		result[i+4] = a
	}
	return result[4:]
}

func exec(ip string, cmd string, timeout time.Duration) (string, error) {
	data := encrypt(cmd)
	port := 9999
	conn, err := net.DialTimeout("udp4", ip+":"+strconv.Itoa(port), timeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(timeout))

	_, err = conn.Write(data)
	if err != nil {
		return "", err
	}

	rData := make([]byte, 1500)
	rLen, err := bufio.NewReader(conn).Read(rData)
	if err != nil {
		return "", err
	}

	return decrypt(rData[:rLen]), nil
}

func Scan(timeout time.Duration) ([]Device, error) {
	devices := []Device{}

	broadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:9999")
	if err != nil {
		return nil, err
	}

	fromAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8755")
	if err != nil {
		return nil, err
	}

	sock, err := net.ListenUDP("udp", fromAddr)
	defer sock.Close()
	if err != nil {
		return nil, err
	}
	sock.SetReadBuffer(2048)

	cmd := encrypt(GET_INFO)
	_, err = sock.WriteToUDP(cmd, broadcastAddr)
	if err != nil {
		return nil, err
	}

	err = sock.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil, err
	}

	for {
		buff := make([]byte, 2048)
		rlen, addr, err := sock.ReadFromUDP(buff)
		if err != nil {
			break
		}

		data := decrypt(buff[:rlen])

		r := Response{}
		if err := json.Unmarshal([]byte(data), &r); err != nil {
			return nil, err
		}

		devices = append(devices, Device{
			IPAddress: addr.IP.String(),
			Info:      *r.System.Info,
		})
	}

	return devices, err
}
