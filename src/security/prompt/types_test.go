package prompt

import "testing"

func TestActionType_String(t *testing.T) {
	cases := map[ActionType]string{
		ActionNone:     "None",
		ActionModified: "Modified",
		ActionWarned:   "Warned",
		ActionBlocked:  "Blocked",
		ActionLogged:   "Logged",
		ActionReported: "Reported",
		ActionType(99): "Unknown",
	}
	for action, want := range cases {
		if got := action.String(); got != want {
			t.Errorf("ActionType(%d).String() = %q, want %q", action, got, want)
		}
	}
}
