package audit

type Observer interface {
	Send(event AuditEvent) error
}
