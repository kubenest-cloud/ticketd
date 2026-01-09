package web

const (
	pageSize = 20
)

// totalPages calculates the total number of pages needed for the given total count.
// It accounts for partial pages by rounding up.
// Returns 1 if total is 0 to avoid division by zero.
func totalPages(total int) int {
	if total == 0 {
		return 1
	}
	pages := total / pageSize
	if total%pageSize != 0 {
		pages++
	}
	return pages
}

// prevPage returns the previous page number, or 0 if there is no previous page.
// Used in templates to determine if a "Previous" link should be shown.
func prevPage(current int) int {
	if current > 1 {
		return current - 1
	}
	return 0
}

// nextPage returns the next page number, or 0 if there is no next page.
// Used in templates to determine if a "Next" link should be shown.
func nextPage(current, total int) int {
	if current < totalPages(total) {
		return current + 1
	}
	return 0
}
