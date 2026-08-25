package enum

import "time"

const (
	ApplicationDefaultAccessTokenTTL  = 5 * time.Minute
	ApplicationMaxAccessTokenTTL      = 15 * time.Minute
	ApplicationDefaultRefreshTokenTTL = 7 * 24 * time.Hour
	ApplicationMaxRefreshTokenTTL     = 30 * 24 * time.Hour
)

type ApplicationStatus uint8

const (
	ApplicationStatusDisabled ApplicationStatus = 1 + iota
	ApplicationStatusActive
)

var applicationStatusNames = map[ApplicationStatus]string{
	ApplicationStatusDisabled: "disabled",
	ApplicationStatusActive:   "active",
}

var applicationStatusValues = map[string]ApplicationStatus{
	"disabled": ApplicationStatusDisabled,
	"active":   ApplicationStatusActive,
}

func (s ApplicationStatus) Valid() bool    { return enumValid(applicationStatusNames, s) }
func (s ApplicationStatus) String() string { return enumString(applicationStatusNames, s) }

func (s ApplicationStatus) MarshalJSON() ([]byte, error) {
	return enumMarshalJSON(applicationStatusNames, s)
}

func (s *ApplicationStatus) UnmarshalJSON(data []byte) error {
	return enumUnmarshalJSON(applicationStatusValues, data, s)
}

func (s ApplicationStatus) MarshalText() ([]byte, error) {
	return enumMarshalText(applicationStatusNames, s)
}

func (s *ApplicationStatus) UnmarshalText(data []byte) error {
	return enumUnmarshalText(applicationStatusValues, data, s)
}

func ApplicationStatusOptions() []Option {
	return enumOptions(applicationStatusNames)
}

type ApplicationClientType uint8

const (
	ApplicationClientTypeConfidential ApplicationClientType = 1 + iota
	ApplicationClientTypePublic
)

var applicationClientTypeNames = map[ApplicationClientType]string{
	ApplicationClientTypeConfidential: "confidential",
	ApplicationClientTypePublic:       "public",
}

var applicationClientTypeValues = map[string]ApplicationClientType{
	"confidential": ApplicationClientTypeConfidential,
	"public":       ApplicationClientTypePublic,
}

func (t ApplicationClientType) Valid() bool { return enumValid(applicationClientTypeNames, t) }
func (t ApplicationClientType) String() string {
	return enumString(applicationClientTypeNames, t)
}

func (t ApplicationClientType) MarshalJSON() ([]byte, error) {
	return enumMarshalJSON(applicationClientTypeNames, t)
}

func (t *ApplicationClientType) UnmarshalJSON(data []byte) error {
	return enumUnmarshalJSON(applicationClientTypeValues, data, t)
}

func (t ApplicationClientType) MarshalText() ([]byte, error) {
	return enumMarshalText(applicationClientTypeNames, t)
}

func (t *ApplicationClientType) UnmarshalText(data []byte) error {
	return enumUnmarshalText(applicationClientTypeValues, data, t)
}

func ApplicationClientTypeOptions() []Option {
	return enumOptions(applicationClientTypeNames)
}
