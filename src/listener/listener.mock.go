package listener

import "github.com/lib/pq"

type MockListenerImpl struct {
	MockClose               func() error
	MockListen              func(channel string) error
	MockNotificationChannel func() <-chan *pq.Notification
	MockPing                func() error
	MockUnlisten            func(channel string) error
	MockUnlistenAll         func() error
}

func (m MockListenerImpl) Close() error {
	return m.MockClose()
}
func (m MockListenerImpl) Listen(channel string) error {
	return m.MockListen(channel)
}
func (m MockListenerImpl) NotificationChannel() <-chan *pq.Notification {
	return m.MockNotificationChannel()
}
func (m MockListenerImpl) Ping() error {
	return m.MockPing()
}
func (m MockListenerImpl) Unlisten(channel string) error {
	return m.MockUnlisten(channel)
}
func (m MockListenerImpl) UnlistenAll() error {
	return m.MockUnlistenAll()
}
