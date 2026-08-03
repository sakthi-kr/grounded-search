package authorization

// Decision is an ACL rule decision.
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
)

// DocumentStatus is the lifecycle state of a document.
type DocumentStatus string

const (
	DocumentActive   DocumentStatus = "active"
	DocumentDisabled DocumentStatus = "disabled"
	DocumentDeleted  DocumentStatus = "deleted"
)

// User describes the identity state required for authorisation.
type User struct {
	Exists   bool
	Enabled  bool
	GroupIDs map[string]struct{}
}

// Document describes the document state and attached ACL rules.
type Document struct {
	Status     DocumentStatus
	Public     bool
	UserRules  map[string]Decision
	GroupRules map[string]Decision
}

// Result is the complete, explainable authorisation result.
type Result struct {
	Allowed bool
	Reason  string
}
