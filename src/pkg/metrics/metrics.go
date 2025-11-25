package metrics

import (
	"strconv"
	"sync/atomic"
	"time"
)

var (
	startTime        = time.Now()
	messagesSent     atomic.Uint64
	messagesReceived atomic.Uint64
	sessionsActive   atomic.Int64
)

func IncMessagesSent() {
	messagesSent.Add(1)
}

func IncMessagesReceived() {
	messagesReceived.Add(1)
}

func SetSessionsActive(n int64) {
	sessionsActive.Store(n)
}

func AddSessions(delta int64) {
	sessionsActive.Add(delta)
}

func UptimeSeconds() float64 {
	return time.Since(startTime).Seconds()
}

func Snapshot() (sent, received uint64, sessions int64, uptimeSeconds float64) {
	sent = messagesSent.Load()
	received = messagesReceived.Load()
	sessions = sessionsActive.Load()
	uptimeSeconds = UptimeSeconds()
	return
}

// Prometheus text exposition (very small set of gauges/counters)
func PrometheusText() string {
	sent, recv, sess, up := Snapshot()
	return "# HELP whatsapp_messages_sent_total Total outbound WhatsApp messages\n" +
		"# TYPE whatsapp_messages_sent_total counter\n" +
		"whatsapp_messages_sent_total " + formatUint(sent) + "\n" +
		"# HELP whatsapp_messages_received_total Total inbound WhatsApp messages\n" +
		"# TYPE whatsapp_messages_received_total counter\n" +
		"whatsapp_messages_received_total " + formatUint(recv) + "\n" +
		"# HELP whatsapp_sessions_active Number of active WhatsApp sessions\n" +
		"# TYPE whatsapp_sessions_active gauge\n" +
		"whatsapp_sessions_active " + formatInt(sess) + "\n" +
		"# HELP whatsapp_uptime_seconds Process uptime in seconds\n" +
		"# TYPE whatsapp_uptime_seconds gauge\n" +
		"whatsapp_uptime_seconds " + formatFloat(up) + "\n"
}

func formatUint(v uint64) string { return formatFloat(float64(v)) }
func formatInt(v int64) string   { return formatFloat(float64(v)) }

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
