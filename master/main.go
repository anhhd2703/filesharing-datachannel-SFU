package main

import (
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
	c, err := sdk.NewClient(e, addr, "master")
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
		if dc.Label() == "data1" {
			go func() {
				log.Errorf("-------------OnMessage------------------")
				// util.Detach(dc)
				util.DetachGoRoutine()
			}()
			dc.OnOpen(func() {
				dc.OnMessage(onMessage)
			})
		}
	}

	select {}
}
func onMessage(msg webrtc.DataChannelMessage) {
	log.Errorf("-------------OnMessage------------------=%v", string(msg.Data))
}
