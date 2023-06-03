package commands

type actionResult struct {
	resetSpamFilter bool
}

func (ar *actionResult) ResetSpamFilter() bool {
	if ar == nil {
		return false
	}
	return ar.resetSpamFilter
}
