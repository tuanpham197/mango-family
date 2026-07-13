// Package subscriber — cầu nối pubsub → wshub (research D8).
package subscriber

import (
	"household-finance/api/component/pubsub"
	"household-finance/api/component/wshub"
)

// Start chạy goroutine chuyển mọi message pubsub thành WS broadcast theo hộ.
func Start(ps pubsub.PubSub, hub *wshub.Hub) {
	ch := ps.Subscribe()
	go func() {
		for msg := range ch {
			hub.Broadcast(msg.HouseholdID, msg.Topic)
		}
	}()
}
