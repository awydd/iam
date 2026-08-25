package enum

type KeypairAlgorithm uint8

const (
	KeypairAlgoRS256 KeypairAlgorithm = 1 + iota
	KeypairAlgoES256
)

var keypairAlgorithmNames = map[KeypairAlgorithm]string{
	KeypairAlgoRS256: "RS256",
	KeypairAlgoES256: "ES256",
}

var keypairAlgorithmValues = map[string]KeypairAlgorithm{
	"RS256": KeypairAlgoRS256,
	"ES256": KeypairAlgoES256,
}

func (a KeypairAlgorithm) Valid() bool    { return enumValid(keypairAlgorithmNames, a) }
func (a KeypairAlgorithm) String() string { return enumString(keypairAlgorithmNames, a) }

func (a KeypairAlgorithm) MarshalJSON() ([]byte, error) {
	return enumMarshalJSON(keypairAlgorithmNames, a)
}

func (a *KeypairAlgorithm) UnmarshalJSON(data []byte) error {
	return enumUnmarshalJSON(keypairAlgorithmValues, data, a)
}

func (a KeypairAlgorithm) MarshalText() ([]byte, error) {
	return enumMarshalText(keypairAlgorithmNames, a)
}

func (a *KeypairAlgorithm) UnmarshalText(data []byte) error {
	return enumUnmarshalText(keypairAlgorithmValues, data, a)
}

func KeypairAlgorithmOptions() []Option {
	return enumOptions(keypairAlgorithmNames)
}

type KeypairStatus uint8

const (
	KeypairStatusActive KeypairStatus = 1 + iota
	KeypairStatusGrace
	KeypairStatusRetired
)

var keypairStatusNames = map[KeypairStatus]string{
	KeypairStatusActive:  "active",
	KeypairStatusGrace:   "grace",
	KeypairStatusRetired: "retired",
}

var keypairStatusValues = map[string]KeypairStatus{
	"active":  KeypairStatusActive,
	"grace":   KeypairStatusGrace,
	"retired": KeypairStatusRetired,
}

func (s KeypairStatus) Valid() bool    { return enumValid(keypairStatusNames, s) }
func (s KeypairStatus) String() string { return enumString(keypairStatusNames, s) }

func (s KeypairStatus) MarshalJSON() ([]byte, error) {
	return enumMarshalJSON(keypairStatusNames, s)
}

func (s *KeypairStatus) UnmarshalJSON(data []byte) error {
	return enumUnmarshalJSON(keypairStatusValues, data, s)
}

func (s KeypairStatus) MarshalText() ([]byte, error) {
	return enumMarshalText(keypairStatusNames, s)
}

func (s *KeypairStatus) UnmarshalText(data []byte) error {
	return enumUnmarshalText(keypairStatusValues, data, s)
}

func KeypairStatusOptions() []Option {
	return enumOptions(keypairStatusNames)
}
