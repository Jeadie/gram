package main

type SearchResult struct {
	rowRef *Row
	startI uint
	rowI   uint
}

// SearchRows concurrently searches for a given query string in a slice of Rows.
// It returns a channel of search results.
func SearchRows(rows []Row, q string) <-chan SearchResult {
	results := make(chan SearchResult)

	go func() {
		defer close(results)

		for i := range rows {
			y := rows[i].RenderIndexOf(q, 0)

			// Handle multiple search terms in one Row.
			for y != -1 {
				results <- SearchResult{
					rowRef: &rows[i],
					startI: uint(y),
					rowI:   uint(i),
				}
				y = rows[i].RenderIndexOf(q, y+1)
			}
		}
	}()
	return results

}
