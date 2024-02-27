package base

type SafeString struct {
	ptr *string
}

func NewSafeString(s *string) *SafeString {
	return &SafeString{ptr: s}
}

func (s *SafeString) String() string {
	if s.ptr == nil {
		return "<nil>"
	}
	return *s.ptr
}
