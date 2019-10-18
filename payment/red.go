package payment

import (
	"encoding/xml"
	"github.com/sqeven/weapp/util"
	"strconv"
	"time"
)

const (
	redAPI = "/mmpaymkttransfers/sendredpack"
)

// RedPacket params
type RedPacket struct {
	// required
	MchBillno   string `xml:"mch_billno"`   // 商户订单号
	MchID       string `xml:"mch_id"`       // 商户号
	AppID       string `xml:"wxappid"`      // 公众账号appid
	SendName    string `xml:"send_name"`    // 商户名称
	Openid      string `xml:"re_openid"`    // 用户openid
	TotalAmount int    `xml:"total_amount"` // 付款金额
	TotalNum    int    `xml:"total_num"`    // 红包发放总人数
	Wishing     string `xml:"wishing"`      // 红包祝福语
	IP          string `xml:"client_ip,omitempty"`
	ActName     string `xml:"act_name"`            // 活动名称
	Remark      string `xml:"remark"`              // 备注
	SceneId     string `xml:"scene_id"`            // 场景id 发放红包使用场景，红包金额大于200或者小于1元时必传
	RiskInfo    string `xml:"risk_info,omitempty"` // 活动信息
}

type redpacket struct {
	XMLName xml.Name `xml:"xml"`
	RedPacket
	NonceStr string `xml:"nonce_str"`
	Sign     string `xml:"sign"` // 签名
}

type redResponse struct {
	response
	MchBillno   string `xml:"mch_billno"` // 商户订单号
	MchID       string `xml:"mchid"`
	AppID       string `xml:"wxappid"` // 公众账号appid
	ReOpenid    string `xml:"re_openid"`
	TotalAmount int    `xml:"total_amount"`
	SendListid  string `xml:"send_listid"` // 微信单号
}

// 发放成功返回数据
type RedResponse struct {
	redResponse
	Datetime time.Time
}

// 请求前准备
func (t *RedPacket) prepare(key string) (redpacket, error) {
	traRed := redpacket{
		RedPacket: *t,
		NonceStr:  util.RandomString(32),
	}

	signData := map[string]string{
		"mch_billno":   traRed.MchBillno,
		"mch_id":       traRed.MchID,
		"wxappid":      traRed.AppID,
		"send_name":    traRed.SendName,
		"re_openid":    traRed.Openid,
		"total_amount": strconv.Itoa(traRed.TotalAmount),
		"total_num":    strconv.Itoa(traRed.TotalNum),
		"wishing":      traRed.Wishing,
		"act_name":     traRed.ActName,
		"remark":       traRed.Remark,
		"scene_id":     traRed.SceneId,
		"nonce_str":    traRed.NonceStr,
	}

	if t.IP == "" {
		ip, err := util.FetchIP()
		if err != nil {
			return traRed, err
		}

		traRed.IP = ip.String()
	}
	signData["client_ip"] = traRed.IP

	if t.RiskInfo != "" {
		signData["risk_info"] = traRed.RiskInfo
	}

	sign, err := util.SignByMD5(signData, key)
	if err != nil {
		return traRed, err
	}
	traRed.Sign = sign

	return traRed, nil
}

func (t RedPacket) SendRedPacket(key string, certPath, keyPath string) (res RedResponse, err error) {
	reqData, err := t.prepare(key)
	if err != nil {
		return
	}

	resData, err := util.TSLPostXML(baseURL+redAPI, reqData, certPath, keyPath)
	if err != nil {
		return
	}

	var ret redResponse
	if err = xml.Unmarshal(resData, &ret); err != nil {
		return
	}

	if err = ret.Check(); err != nil {
		return
	}

	res.redResponse = ret
	res.Datetime, err = time.Parse(transferTimeFormat, time.Now().Format(transferTimeFormat))

	return
}
