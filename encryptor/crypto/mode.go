package crypto

// Mode are the supported modes for a cipher.
type Mode string

func (m Mode) String() string {
	return string(m)
}
