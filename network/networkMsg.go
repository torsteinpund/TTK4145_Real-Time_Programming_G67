package network

// import (
// 	"fmt"
// 	"reflect"
// 	"encoding/json"
// )

// type SimpleNetworkMsg struct {
// 	MsgType string		`json:"msgType"`
// 	MsgData []byte		`json:"msgData"`
// }

// func DecodeMessage(ch_RX RXChannels, msg SimpleNetworkMsg){

// 	rxTypes := reflect.TypeOf(ch_RX) // Get the type of the struct RXChannels
// 	rxValue := reflect.ValueOf(ch_RX) // Get the value of the struct RXChannels

// 	// Loop through all the fields in the struct RXChannels
// 	for i := 0; i < rxTypes.NumField(); i++ {
// 		ch := rxTypes.Field(i)
// 		chValue := rxValue.Field(i).Interface()
// 		// Get the type of the channel
// 		chType := reflect.TypeOf(chValue).Elem()
// 		typeName := ch.Tag.Get("addr")
// 		if msg.MsgType == typeName {
// 			// Create a new instance of the type of the channel
// 			value := reflect.New(chType)
// 			// Decode the JSON message and put it in the new instance
// 			err := json.Unmarshal(msg.MsgData, value.Interface())
// 			if err != nil {
// 				fmt.Println("Error decoding JSON 2:" + err.Error())
// 			}
// 			// Send the new instance to the channel
// 			reflect.Select([]reflect.SelectCase{{
// 				Dir:  reflect.SelectSend,
// 				Chan: reflect.ValueOf(chValue),
// 				Send: reflect.Indirect(value),
// 			}})
// 		}
// 	}

// }