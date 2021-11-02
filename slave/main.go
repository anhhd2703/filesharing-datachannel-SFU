package main

import (
	"encoding/json"
	"strconv"
	"webrtc/proto"
	"webrtc/util"

	log "github.com/pion/ion-log"
	sdk "github.com/pion/ion-sdk-go"
	"github.com/pion/webrtc/v3"
)

func main() {
	log.Init("debug")
	session := "test"
	addr := "localhost:50051"

	webrtcCfg := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			webrtc.ICEServer{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	config := sdk.Config{
		WebRTC: sdk.WebRTCTransportConfig{
			Configuration: webrtcCfg,
		},
	}
	// new sdk engine
	e := sdk.NewEngine(config)
	// create a new client from engine
	c, err := sdk.NewClient(e, addr, "slave")
	if err != nil {
		log.Errorf("err=%v", err)
		return
	}
	err = c.Join(session, nil)
	if err != nil {
		log.Errorf("err=%v", err)
		return
	}
	c.CreateDataChannel("data1")

	c.OnDataChannel = func(dc *webrtc.DataChannel) {
		log.Errorf("-------------------------------=%v", dc.Label())
		dc.OnOpen(func() {
			// buffArray := []byte{}
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				if msg.IsString == true {
					log.Errorf("-------------------------------=%v", string(msg.Data))
					number, _ := strconv.ParseUint(string(msg.Data), 10, 64)
					util.Combine(number)
					// util.WriteToFile(buffArray)
				} else {
					var data proto.Data
					json.Unmarshal(msg.Data, &data)
					// buffArray = append(buffArray, data.Buff...)
					util.WriteToLocal(&data)
				}

			})
		})
	}

	select {}
}
func onMessage(msg webrtc.DataChannelMessage) {

	if msg.IsString == true {
		log.Errorf("-------------------------------=%v", string(msg.Data))
		number, _ := strconv.ParseUint(string(msg.Data), 10, 64)
		util.Combine(number)
	} else {
		// util.WriteToLocal(msg.Data)
	}

}
