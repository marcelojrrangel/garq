package worker

// CopyPayload represents the payload for a copy job.
type CopyPayload struct {
	Sources  []string `json:"sources"`
	Dest     string   `json:"dest"`
	Conflict string   `json:"conflict"`
}

// CompressPayload represents the payload for a compress job.
type CompressPayload struct {
	Sources  []string `json:"sources"`
	Dest     string   `json:"dest"`
	Conflict string   `json:"conflict"`
}

// ExtractPayload represents the payload for an extract job.
type ExtractPayload struct {
	Archive  string `json:"archive"`
	Dest     string `json:"dest"`
	Conflict string `json:"conflict"`
}

// MovePayload represents the payload for a move job.
type MovePayload struct {
	Sources  []string `json:"sources"`
	Dest     string   `json:"dest"`
	Conflict string   `json:"conflict"`
}

// DeletePayload represents the payload for a delete job.
type DeletePayload struct {
	Sources []string `json:"sources"`
}
