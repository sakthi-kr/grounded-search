package authorization

import "testing"

func TestEvaluatePolicyOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userID   string
		user     User
		document Document
		want     Result
	}{
		{
			name:   "deleted document denies public enabled user",
			userID: "alice",
			user:   enabledUser(),
			document: Document{
				Status: DocumentDeleted,
				Public: true,
			},
			want: Result{Allowed: false, Reason: ReasonDocumentDeleted},
		},
		{
			name:   "disabled document denies explicit allow",
			userID: "alice",
			user:   enabledUser(),
			document: Document{
				Status:    DocumentDisabled,
				UserRules: map[string]Decision{"alice": DecisionAllow},
			},
			want: Result{Allowed: false, Reason: ReasonDocumentDisabled},
		},
		{
			name:   "unknown user denied",
			userID: "alice",
			user:   User{Exists: false},
			document: Document{
				Status: DocumentActive,
				Public: true,
			},
			want: Result{Allowed: false, Reason: ReasonUnknownUser},
		},
		{
			name:   "disabled user denied",
			userID: "alice",
			user:   User{Exists: true, Enabled: false},
			document: Document{
				Status: DocumentActive,
				Public: true,
			},
			want: Result{Allowed: false, Reason: ReasonUserDisabled},
		},
		{
			name:   "explicit user deny overrides public",
			userID: "alice",
			user:   enabledUser(),
			document: Document{
				Status:    DocumentActive,
				Public:    true,
				UserRules: map[string]Decision{"alice": DecisionDeny},
			},
			want: Result{Allowed: false, Reason: ReasonExplicitUserDeny},
		},
		{
			name:   "group deny overrides explicit user allow",
			userID: "alice",
			user: User{
				Exists:   true,
				Enabled:  true,
				GroupIDs: set("engineering"),
			},
			document: Document{
				Status: DocumentActive,
				UserRules: map[string]Decision{
					"alice": DecisionAllow,
				},
				GroupRules: map[string]Decision{
					"engineering": DecisionDeny,
				},
			},
			want: Result{Allowed: false, Reason: ReasonMatchingGroupDeny},
		},
		{
			name:   "public document allowed",
			userID: "alice",
			user:   enabledUser(),
			document: Document{
				Status: DocumentActive,
				Public: true,
			},
			want: Result{Allowed: true, Reason: ReasonPublicDocument},
		},
		{
			name:   "explicit user allow",
			userID: "alice",
			user:   enabledUser(),
			document: Document{
				Status: DocumentActive,
				UserRules: map[string]Decision{
					"alice": DecisionAllow,
				},
			},
			want: Result{Allowed: true, Reason: ReasonExplicitUserAllow},
		},
		{
			name:   "group allow",
			userID: "alice",
			user: User{
				Exists:   true,
				Enabled:  true,
				GroupIDs: set("engineering"),
			},
			document: Document{
				Status: DocumentActive,
				GroupRules: map[string]Decision{
					"engineering": DecisionAllow,
				},
			},
			want: Result{Allowed: true, Reason: ReasonMatchingGroupAllow},
		},
		{
			name:   "default deny",
			userID: "alice",
			user:   enabledUser(),
			document: Document{
				Status: DocumentActive,
			},
			want: Result{Allowed: false, Reason: ReasonDefaultDeny},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := Evaluate(test.userID, test.user, test.document)
			if got != test.want {
				t.Fatalf("Evaluate() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestEvaluateDoesNotMutateInputs(t *testing.T) {
	t.Parallel()

	user := User{
		Exists:   true,
		Enabled:  true,
		GroupIDs: set("engineering"),
	}
	document := Document{
		Status: DocumentActive,
		UserRules: map[string]Decision{
			"alice": DecisionAllow,
		},
		GroupRules: map[string]Decision{
			"engineering": DecisionAllow,
		},
	}

	_ = Evaluate("alice", user, document)

	if len(user.GroupIDs) != 1 {
		t.Fatalf("user GroupIDs mutated: %+v", user.GroupIDs)
	}
	if len(document.UserRules) != 1 {
		t.Fatalf("document UserRules mutated: %+v", document.UserRules)
	}
	if len(document.GroupRules) != 1 {
		t.Fatalf("document GroupRules mutated: %+v", document.GroupRules)
	}
}

func enabledUser() User {
	return User{
		Exists:   true,
		Enabled:  true,
		GroupIDs: map[string]struct{}{},
	}
}

func set(values ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
