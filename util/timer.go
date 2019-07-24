package util

import "time"

//平衡二叉树 O(logN)
type Timer struct {
}

//指定时间调用
func (t *Timer) RunAt(timeAt time.Time, cb func()) uint16 {
	return t.addTimer(timeAt, cb, 0)
}

//等一段时间调用
func (t *Timer) RunAfter(delay time.Duration, cb func()) uint16 {
	now := time.Now().Add(delay)
	return t.RunAt(now, cb)
}

//以固定时间调用
func (t *Timer) RunEvery(interval time.Duration, cb func()) uint16 {
	now := time.Now().Add(interval)
	return t.addTimer(now, cb, interval)
}

func (t *Timer) addTimer(timeAt time.Time, cb func(), interval time.Duration) uint16 {
	return 0
}

//停用
func (t *Timer) Cancle(uint16) {}
