package stuber

import (
	"bytes"
	"sync"
	"sync/atomic"
)

// BidiResult holds matching stubs for bidirectional streaming.
type BidiResult struct {
	searcher      *searcher
	lookup        *searcherLookup
	reserveQuery  Query
	inputBuf      [1]map[string]any
	matchingStubs []*Stub      // Stubs that match the current message pattern
	messageCount  atomic.Int32 // Number of messages processed so far
	mu            sync.Mutex   // Thread safety for concurrent access
}

// Next processes the next message in the bidirectional stream and returns the matching stub.
func (br *BidiResult) Next(messageData map[string]any) (*Stub, error) {
	br.mu.Lock()
	defer br.mu.Unlock()

	if messageData == nil {
		return nil, ErrStubNotFound
	}

	matched, messageIndex := br.ensureMatchingStubs(messageData)
	if !matched {
		return nil, ErrStubNotFound
	}

	itemQuery := br.queryForMessage(messageData)
	isFirstMessage := messageIndex == 0

	for {
		bestStub, bestIndex := br.selectBestStub(itemQuery, messageIndex)
		if bestStub == nil {
			return nil, ErrStubNotFound
		}

		if reservedStub, reserved := br.tryReserveAndFinalize(bestStub, itemQuery, isFirstMessage); reserved {
			return reservedStub, nil
		}
		// Stub exhausted (Times), remove and try next
		br.removeStubFromMatchingByIndex(bestIndex)
	}
}

// GetMessageIndex returns the current message index in the bidirectional stream.
func (br *BidiResult) GetMessageIndex() int {
	return int(br.messageCount.Load())
}

func (br *BidiResult) ensureMatchingStubs(messageData map[string]any) (bool, int) {
	messageIndex := 0

	if len(br.matchingStubs) == 0 {
		allStubs, err := br.lookup.LookupServiceAvailable(br.reserveQuery.Service, br.reserveQuery.Method)
		if err != nil {
			return false, messageIndex
		}

		for stub := range allStubs {
			if matchBidiStubMessage(stub, messageData) {
				br.matchingStubs = append(br.matchingStubs, stub)
			}
		}
	} else {
		messageIndex = int(br.messageCount.Add(1))
		br.filterMatchingStubs(messageData)
	}

	return len(br.matchingStubs) > 0, messageIndex
}

func (br *BidiResult) filterMatchingStubs(messageData map[string]any) {
	filtered := br.matchingStubs[:0]

	for _, stub := range br.matchingStubs {
		if matchBidiStubMessage(stub, messageData) {
			filtered = append(filtered, stub)
		}
	}

	br.matchingStubs = filtered
}

func (br *BidiResult) selectBestStub(itemQuery Query, messageIndex int) (*Stub, int) {
	var (
		bestStub  *Stub
		bestRank  float64
		bestIndex = -1
	)

	for i, stub := range br.matchingStubs {
		rank := scoreBidiStubMessage(itemQuery, stub, messageIndex)
		priorityBonus := float64(stub.Priority) * PriorityMultiplier
		totalRank := rank + priorityBonus

		if bestStub == nil || totalRank > bestRank {
			bestStub = stub
			bestRank = totalRank
			bestIndex = i

			continue
		}

		if totalRank == bestRank && bytes.Compare(stub.ID[:], bestStub.ID[:]) < 0 {
			bestStub = stub
			bestIndex = i
		}
	}

	return bestStub, bestIndex
}

func (br *BidiResult) tryReserveAndFinalize(bestStub *Stub, itemQuery Query, isFirstMessage bool) (*Stub, bool) {
	matchNumber, reserved := br.searcher.tryReserve(itemQuery, bestStub)
	if !reserved {
		return nil, false
	}

	if !bestStub.IsClientStream() && isFirstMessage {
		br.matchingStubs = nil
		br.messageCount.Store(0)
	}

	return effectiveDelayStub(bestStub, matchNumber), true
}

func (br *BidiResult) queryForMessage(messageData map[string]any) Query {
	br.inputBuf[0] = messageData

	return br.reserveQuery
}

func (br *BidiResult) removeStubFromMatchingByIndex(index int) {
	if index < 0 || index >= len(br.matchingStubs) {
		return
	}

	last := len(br.matchingStubs) - 1
	br.matchingStubs[index] = br.matchingStubs[last]
	br.matchingStubs[last] = nil
	br.matchingStubs = br.matchingStubs[:last]
}

// findBidi retrieves a BidiResult for bidirectional streaming with the given QueryBidi.
// For bidirectional streaming, each message is treated as a separate unary request.
func (s *searcher) findBidi(query QueryBidi) (*BidiResult, error) {
	// Check if the QueryBidi has an ID field
	if query.ID != nil {
		// For ID-based queries, we can't use bidirectional streaming - fallback to regular search
		return s.searchByIDBidi(query)
	}

	if err := s.ensureServiceMethodExists(query.Service, query.Method); err != nil {
		return nil, err
	}

	return s.newBidiResult(query, nil, s.lookup(query.Session)), nil
}

// searchByIDBidi handles ID-based queries for bidirectional streaming.
// Since we can't use bidirectional streaming for ID-based queries, we fallback to regular search.
func (s *searcher) searchByIDBidi(query QueryBidi) (*BidiResult, error) {
	if err := s.ensureServiceMethodExists(query.Service, query.Method); err != nil {
		return nil, err
	}

	lookup, found := s.lookupVisibleByID(query.Session, *query.ID)
	if found == nil {
		// Return an error if the Stub value is not found
		return nil, ErrServiceNotFound
	}

	// tryReserve not called here - BidiResult will call it when GetNextStub
	return s.newBidiResult(query, []*Stub{found}, lookup), nil
}

func (s *searcher) newBidiResult(query QueryBidi, matchingStubs []*Stub, lookup *searcherLookup) *BidiResult {
	result := &BidiResult{
		searcher:      s,
		lookup:        lookup,
		matchingStubs: matchingStubs,
	}

	result.reserveQuery = Query{
		Service:       query.Service,
		Method:        query.Method,
		Session:       query.Session,
		StrictService: query.StrictService,
		Headers:       query.Headers,
		Input:         result.inputBuf[:],
		toggles:       query.toggles,
	}

	return result
}
