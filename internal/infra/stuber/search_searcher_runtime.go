package stuber

import "github.com/google/uuid"

// upsert inserts the given stub values into the searcher. If a stub value
// already exists with the same key, it is updated.
//
// The function returns a slice of UUIDs representing the keys of the
// inserted or updated values.
func (s *searcher) upsert(values ...*Stub) []uuid.UUID {
	return s.storage.upsert(values...)
}

// del deletes the stub values with the given UUIDs from the searcher.
//
// Returns the number of stub values that were successfully deleted.
func (s *searcher) del(ids ...uuid.UUID) int {
	if len(ids) == 0 {
		return 0
	}

	idSet := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.stubCallCount {
		if _, ok := idSet[key.id]; ok {
			delete(s.stubCallCount, key)
		}
	}

	return s.storage.del(ids...)
}

// delBySession deletes all stubs belonging to the given session.
func (s *searcher) delBySession(session string) int {
	if session == "" {
		return 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.stubCallCount {
		if key.session == session {
			delete(s.stubCallCount, key)
		}
	}

	return s.storage.delBySession(session)
}

// findByID retrieves the stub value associated with the given ID from the
// searcher.
//
// Returns a pointer to the Stub struct associated with the given ID, or nil
// if not found.
func (s *searcher) findByID(id uuid.UUID) *Stub {
	return s.storage.findByID(id)
}

// findBy retrieves all Stub values that match the given service and method
// from the searcher, sorted by score in descending order.
func (s *searcher) findBy(service, method string) ([]*Stub, error) {
	seq, err := s.storage.findAll(service, method)
	if err != nil {
		return nil, s.wrap(err)
	}

	return collectStubs(seq), nil
}

// clear resets the searcher.
//
// It clears the stubUsed map and calls the storage clear method.
func (s *searcher) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear the stubCallCount map.
	s.stubCallCount = make(map[callCountKey]int)

	// Clear lookup cache.
	s.lookupMu.Lock()
	s.lookupCache = make(map[string]*searcherLookup)
	s.lookupMu.Unlock()

	// Clear the storage.
	s.storage.clear()
}

// all returns all Stub values stored in the searcher.
//
// Returns:
// - []*Stub: The Stub values stored in the searcher.
func (s *searcher) all() []*Stub {
	return collectStubs(s.storage.values())
}

func (s *searcher) sessions() []string {
	return s.storage.sessionsList()
}

// used returns all Stub values that have been used by the searcher.
//
// Returns:
// - []*Stub: The Stub values that have been used by the searcher.
func (s *searcher) used() []*Stub {
	s.mu.RLock()
	defer s.mu.RUnlock()

	usedIDs := s.collectUsedIDs()

	seq := func(yield func(uuid.UUID) bool) {
		for id := range usedIDs {
			if !yield(id) {
				return
			}
		}
	}

	return collectStubs(s.storage.findByIDs(seq))
}

// unused returns all Stub values that have not been used by the searcher (in any session).
func (s *searcher) unused() []*Stub {
	s.mu.RLock()
	usedIDs := s.collectUsedIDs()
	s.mu.RUnlock()

	// Collect unused stubs in a single pass
	var unused []*Stub

	for stub := range s.storage.values() {
		if _, exists := usedIDs[stub.ID]; !exists {
			unused = append(unused, stub)
		}
	}

	return unused
}

// find retrieves the Stub value associated with the given Query from the searcher.
//
// Parameters:
// - query: The Query used to search for a Stub value.
//
// Returns:
// - *Result: The Result containing the found Stub value (if any), or nil.
// - error: An error if the search fails.
func (s *searcher) find(query Query) (*Result, error) {
	if query.ID != nil {
		return s.searchByID(query)
	}

	return s.searchOptimized(query)
}

// searchByID retrieves the Stub value associated with the given ID from the searcher.
func (s *searcher) searchByID(query Query) (*Result, error) {
	if err := s.ensureServiceMethodExists(query.Service, query.Method); err != nil {
		return nil, err
	}

	_, found := s.lookupVisibleByID(query.Session, *query.ID)
	if found != nil {
		if matchNumber, reserved := s.tryReserve(query, found); reserved {
			return &Result{found: effectiveDelayStub(found, matchNumber)}, nil
		}
	}

	return nil, ErrServiceNotFound
}

// tryReserve atomically checks if the stub can be used (under Times limit) and increments the count.
// When query.Session is set, the count is per-session (parallel test isolation).
// Returns the one-based match number and true if the reservation succeeded.
// Internal requests do not consume the counter and return match number zero.
func (s *searcher) tryReserve(query Query, stub *Stub) (int, bool) {
	if stub == nil {
		return 0, false
	}

	if query.RequestInternal() {
		return 0, true
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := callCountKey{id: stub.ID, session: query.Session}

	times := stub.EffectiveTimes()
	if times > 0 && s.stubCallCount[key] >= times {
		return 0, false
	}

	s.stubCallCount[key]++

	return s.stubCallCount[key], true
}
