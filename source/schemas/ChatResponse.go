package schemas

import utils_time "dvah/utils/time"

type ChatResponse struct {
	ID      string              `json:"id"`
    Object  string              `json:"object"`
	Created utils_time.UnixTime `json:"created"`
    Choices []Choice            `json:"choices"`
    Usage   *Usage              `json:"usage,omitempty"`
}

