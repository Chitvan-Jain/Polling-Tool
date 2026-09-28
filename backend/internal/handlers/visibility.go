package handlers

import "github.com/chitvanjain/polling-tool/internal/models"

func canSeeResults(poll models.Poll, hasVoted bool) bool {
	switch poll.ResultVisibility {
	case models.ResultVisibilityAfterVoting:
		return hasVoted
	case models.ResultVisibilityAfterEnd:
		return poll.Status == models.PollStatusClosed
	case models.ResultVisibilityPrivate:
		return false
	default:
		return true
	}
}