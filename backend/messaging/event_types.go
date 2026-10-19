package messaging

type EventType string

const (
	EventTypeGroupCreated       EventType = "group.created"
	EventTypeGroupMemberAdded   EventType = "group.member_added"
	EventTypeGroupMemberRemoved EventType = "group.member_removed"

	EventTypeExpenseCreated EventType = "expense.created"
	EventTypeExpenseUpdated EventType = "expense.updated"
	EventTypeExpenseDeleted EventType = "expense.deleted"

	EventTypeSettlementCreated EventType = "settlement.created"
	EventTypeSettlementDeleted EventType = "settlement.deleted"
)

func (t EventType) String() string { return string(t) }
