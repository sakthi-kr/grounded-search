package authorization

import "testing"

func FuzzDenyOverrides(f *testing.F) {
	f.Add(true, true, false, false, false)
	f.Add(true, true, true, true, true)
	f.Add(false, false, true, true, true)

	f.Fuzz(func(
		t *testing.T,
		userExists bool,
		userEnabled bool,
		documentPublic bool,
		userDeny bool,
		groupDeny bool,
	) {
		userRules := map[string]Decision{}
		groupRules := map[string]Decision{}

		if userDeny {
			userRules["alice"] = DecisionDeny
		} else {
			userRules["alice"] = DecisionAllow
		}

		if groupDeny {
			groupRules["engineering"] = DecisionDeny
		} else {
			groupRules["engineering"] = DecisionAllow
		}

		result := Evaluate(
			"alice",
			User{
				Exists:   userExists,
				Enabled:  userEnabled,
				GroupIDs: set("engineering"),
			},
			Document{
				Status:     DocumentActive,
				Public:     documentPublic,
				UserRules:  userRules,
				GroupRules: groupRules,
			},
		)

		if !userExists && result.Allowed {
			t.Fatal("unknown user was allowed")
		}
		if userExists && !userEnabled && result.Allowed {
			t.Fatal("disabled user was allowed")
		}
		if userExists && userEnabled && userDeny && result.Allowed {
			t.Fatal("explicit user deny did not override allow conditions")
		}
		if userExists && userEnabled && groupDeny && result.Allowed {
			t.Fatal("matching group deny did not override allow conditions")
		}
	})
}
