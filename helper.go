package dnet

// 作为通知用的 channel， make(chan struct{}, 1)
func SendNotifyChan(ch chan struct{}) {
	select {
	case <-ch:
	default:
	}
	ch <- struct{}{}
}
