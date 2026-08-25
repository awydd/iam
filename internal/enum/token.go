package enum

type TokenType uint8

const (
	TokenTypeRefresh TokenType = 1 + iota
)

var tokenTypeNames = map[TokenType]string{
	TokenTypeRefresh: "refresh",
}

var tokenTypeValues = map[string]TokenType{
	"refresh": TokenTypeRefresh,
}

func (t TokenType) Valid() bool    { return enumValid(tokenTypeNames, t) }
func (t TokenType) String() string { return enumString(tokenTypeNames, t) }

func (t TokenType) MarshalJSON() ([]byte, error) {
	return enumMarshalJSON(tokenTypeNames, t)
}

func (t *TokenType) UnmarshalJSON(data []byte) error {
	return enumUnmarshalJSON(tokenTypeValues, data, t)
}

func (t TokenType) MarshalText() ([]byte, error) {
	return enumMarshalText(tokenTypeNames, t)
}

func (t *TokenType) UnmarshalText(data []byte) error {
	return enumUnmarshalText(tokenTypeValues, data, t)
}

func TokenTypeOptions() []Option {
	return enumOptions(tokenTypeNames)
}
