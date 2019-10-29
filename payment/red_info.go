package payment

import (
	"encoding/xml"
	"fmt"
	"github.com/sqeven/weapp/util"
	"time"
)

const (
	redInfoAPI = "/mmpaymkttransfers/gethbinfo"
)

// RedPacket params
type RedInfoPacket struct {
	// required
	MchBillno string `xml:"mch_billno"` // 商户订单号
	MchID     string `xml:"mch_id"`     // 商户号
	AppID     string `xml:"appid"`      // 公众账号appid
	BillType  string `xml:"bill_type"`  // MCHT:通过商户订单号获取红包信息。
}

type redinfopacket struct {
	XMLName xml.Name `xml:"xml"`
	RedInfoPacket
	NonceStr string `xml:"nonce_str"`
	Sign     string `xml:"sign"` // 签名
}

type redInfoResponse struct {
	response
	MchBillno    string `xml:"mch_billno"`    // 商户订单号
	MchID        string `xml:"mch_id"`        // 微信支付分配的商户号
	DetailId     string `xml:"detail_id"`     // 使用API发放现金红包时返回的红包单号
	Status       string `xml:"status"`        // SENDING:发放中  SENT:已发放待领取 FAILED：发放失败 RECEIVED:已领取 RFUND_ING:退款中 REFUND:已退款
	SendType     string `xml:"send_type"`     // API:通过API接口发放 UPLOAD:通过上传文件方式发放 ACTIVITY:通过活动方式发放
	HbType       string `xml:"hb_type"`       // GROUP:裂变红包 NORMAL:普通红包
	TotalNum     int    `xml:"total_num"`     // 红包个数
	TotalAmount  int    `xml:"total_amount"`  // 红包总金额（单位分）
	Reason       string `xml:"reason"`        // 发送失败原因
	SendTime     string `xml:"send_time"`     // 红包发送时间
	RefundTime   string `xml:"refund_time"`   // 红包的退款时间（如果其未领取的退款）
	RefundAmount int    `xml:"refund_amount"` // 红包退款金额
	Wishing      string `xml:"wishing"`       // 祝福语
	Remark       string `xml:"remark"`        // 活动描述，低版本微信可见
	ActName      string `xml:"act_name"`      // 发红包的活动名称
	Hblist       hbList `xml:"hblist"`        // 裂变红包的领取列表
}

type hbList struct {
	HbInfo hbInfo `xml:"hbinfo"`
}

type hbInfo struct {
	Openid  string `xml:"openid"`   // 领取红包的openid
	Amount  int    `xml:"amount"`   // 领取金额
	RcvTime string `xml:"rcv_time"` // 领取红包的时间
}

// 请求成功返回数据
type RedInfoResponse struct {
	redInfoResponse
	Datetime time.Time
}

// 请求前准备
func (t *RedInfoPacket) prepare(key string) (redinfopacket, error) {
	traRedInfo := redinfopacket{
		RedInfoPacket: *t,
		NonceStr:      util.RandomString(32),
	}

	signData := map[string]string{
		"mch_billno": traRedInfo.MchBillno,
		"mch_id":     traRedInfo.MchID,
		"appid":      traRedInfo.AppID,
		"bill_type":  traRedInfo.BillType,
		"nonce_str":  traRedInfo.NonceStr,
	}

	sign, err := util.SignByMD5(signData, key)
	if err != nil {
		return traRedInfo, err
	}
	traRedInfo.Sign = sign

	return traRedInfo, nil
}

func (t RedInfoPacket) SendRedInfoPacket(key string, certPath, keyPath string) (res RedInfoResponse, err error) {
	reqData, err := t.prepare(key)
	if err != nil {
		return
	}

	resData, err := util.TSLPostXML(baseURL+redInfoAPI, reqData, certPath, keyPath)
	if err != nil {
		return
	}

	fmt.Println(string(resData))

	var ret redInfoResponse
	if err = xml.Unmarshal(resData, &ret); err != nil {
		return
	}

	if err = ret.Check(); err != nil {
		return
	}

	res.redInfoResponse = ret
	res.Datetime, err = time.Parse(transferTimeFormat, ret.SendTime)

	return
}
