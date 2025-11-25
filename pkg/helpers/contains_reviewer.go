package helpers

// ContainsReviewer проверяет, содержится ли ревьювер в списке
func ContainsReviewer(reviewers []string, userID string) bool {
	for _, reviewerID := range reviewers {
		if reviewerID == userID {
			return true
		}
	}
	return false
}
