package consul

type Action int8

var (
	ActionGet    Action = iota + 1
	ActionCreate Action
	ActionDelete Action
	ActionUpdate Action
)

var action_to_string = map[Action]string{
	ActionGet:    "get",
	ActionCreate: "create",
	ActionDelete: "delete",
	ActionUpdate: "update",
}

var string_to_action = map[string]Action{
	"get":    ActionGet,
	"create": ActionCreate,
	"delete": ActionDelete,
	"update": ActionUpdate,
}

func (a Action) String() string {
	if s, ok := action_to_string[a]; ok {
		return s
	} else {
		return ""
	}
}
