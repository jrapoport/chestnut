package tags

import "strings"

const (
	// TODO: Add support for chestnut & gorm struct field tags
	// ChestnutTag = "cn"
	// GORMTag = "gorm"

	// JSONTag is the default JSON struct tag to use.
	JSONTag = "json"

	// SecureOption is the tag option to enable sparse encryption of a struct field.
	SecureOption = "secure"

	// HashOption is the tag option to hash a struct field of type string. Defaults to SHA256.
	HashOption = "hash"

	jsonSeparator  = ","
	jsonNameIgnore = "-"
)

// Hash provides a type for supported hash function names
type Hash string

const (
	// HashNone is used to indicate no has function was found in the tag options.
	HashNone Hash = ""

	// HashSHA256 sets the HashOption to use sha256. This is the default.
	// TODO: support parsing this from the struct field tag hash option e.g. `...,hash=md5"`
	HashSHA256 = "sha256"
)

func (h Hash) String() string {
	return string(h)
}

// ParseJSONTag returns the name and options for a JSON struct field tag.
func ParseJSONTag(tag string) (name string, opts []string) {
	parts := strings.Split(tag, jsonSeparator)
	switch len(parts) {
	case 0:
		return "", []string{}
	case 1:
		return parts[0], []string{}
	default:
		if IgnoreField(parts[0]) {
			return parts[0], []string{}
		}
		return parts[0], parts[1:]
	}
}

// IgnoreField checks the name to see if field should be ignored.
func IgnoreField(name string) bool {
	return name == jsonNameIgnore
}

// HasOption checks to see if the tag options contain a specific option.
func HasOption(opts []string, opt string) bool {
	for _, s := range opts {
		if s == opt {
			return true
		}
	}
	return false
}

// HashName checks to see if the hash option is set. The struct field *MUST BE*
// type string and capable of holding the decoded hash as a string. If no hash option
// is found it will return HashNone. Defaults to HashSHA256 (sha256).
func HashName(opts []string) Hash {
	if HasOption(opts, HashOption) {
		return HashSHA256
	}
	return HashNone // do not hash
}

// IsSecure checks to see if the secure option is set.
func IsSecure(opts []string) bool {
	return HasOption(opts, SecureOption)
}
