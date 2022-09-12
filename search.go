package main

type SearchResult struct {
	rowRef *Row
	startI uint
	rowI   uint
}

// SearchRows for a given search query q. Returns a channel of search results
func SearchRows(rows []Row, q string) chan SearchResult {
	results := make(chan SearchResult)

	go func(out chan SearchResult, rows []Row, q string) {
		defer close(results)

		for i, r := range rows {
			y := r.RenderIndexOf(q, 0)

			// Handle multiple search terms in one Row.
			for y != -1 {
				results <- SearchResult{
					rowRef: &r,
					startI: uint(y),
					rowI:   uint(i),
				}
				y = r.RenderIndexOf(q, y+1)
			}
		}

	}(results, rows, q)
	return results
}
