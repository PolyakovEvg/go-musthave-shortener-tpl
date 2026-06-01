package audit

import "time"

type Action string

const (
	ActionShorten Action = "shorten"
	ActionFollow  Action = "follow"
)

type AuditEvent struct {
	Ts     int64  `json:"ts"`
	Action Action `json:"action"`
	UserID string `json:"user_id,omitempty"`
	URL    string `json:"url"`
}

func NewAuditEvent(action Action, userID, url string) AuditEvent {
	return AuditEvent{
		Ts:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}
}
