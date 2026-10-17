package handshake

type KeyUpdateRequest uint8

const (
	KeyUpdateNotRequested KeyUpdateRequest = 0
	KeyUpdateRequested    KeyUpdateRequest = 1
)
