package enum

type UserStatus uint8

const (
	UserStatusPending UserStatus = iota + 1
	UserStatusActive
	UserStatusDisabled
)

var userStatusNames = map[UserStatus]string{
	UserStatusPending:  "pending",
	UserStatusActive:   "active",
	UserStatusDisabled: "disabled",
}

var userStatusValues = map[string]UserStatus{
	"pending":  UserStatusPending,
	"active":   UserStatusActive,
	"disabled": UserStatusDisabled,
}

func (s UserStatus) Valid() bool    { return enumValid(userStatusNames, s) }
func (s UserStatus) String() string { return enumString(userStatusNames, s) }

func (s UserStatus) MarshalJSON() ([]byte, error) {
	return enumMarshalJSON(userStatusNames, s)
}

func (s *UserStatus) UnmarshalJSON(data []byte) error {
	return enumUnmarshalJSON(userStatusValues, data, s)
}

func (s UserStatus) MarshalText() ([]byte, error) {
	return enumMarshalText(userStatusNames, s)
}

func (s *UserStatus) UnmarshalText(data []byte) error {
	return enumUnmarshalText(userStatusValues, data, s)
}

func UserStatusOptions() []Option {
	return enumOptions(userStatusNames)
}
