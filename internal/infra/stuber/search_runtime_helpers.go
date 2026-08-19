package stuber

// searchOptimized performs ultra-fast search with minimal allocations.
func (s *searcher) searchOptimized(query Query) (*Result, error) {
	candidates, err := s.resolveSearchCandidates(query)
	if err != nil {
		return nil, err
	}

	return s.processStrategy.Process(query, candidates)
}

func (s *searcher) resolveSearchCandidates(query Query) ([]*Stub, error) {
	lookup := s.lookup(query.Session)

	stubs, err := lookup.LookupServiceAvailable(query.Service, query.Method)
	if err == nil {
		return collectStubs(stubs), nil
	}

	if query.StrictService {
		return nil, ErrStubNotFound
	}

	if !lookup.HasMethodAvailable(query.Method) {
		return nil, ErrStubNotFound
	}

	candidates := collectStubs(lookup.LookupMethodAvailable(query.Method))
	if len(candidates) == 0 {
		return nil, ErrStubNotFound
	}

	return candidates, nil
}

// processStubs processes the collected stubs with ultra-fast paths.
func (s *searcher) processStubs(query Query, stubs []*Stub) (*Result, error) {
	if len(stubs) == 0 {
		return nil, ErrStubNotFound
	}

	if len(stubs) == 1 {
		return s.processSingleStub(query, stubs[0])
	}

	// Parallel processing for multiple stubs
	if len(stubs) >= parallelProcessingThreshold {
		return s.processStubsParallel(query, stubs)
	}

	// Single-threaded processing for small sets
	return s.processStubsSequential(query, stubs)
}

// processStubsSequential processes stubs sequentially (original logic).
func (s *searcher) processStubsSequential(query Query, stubs []*Stub) (*Result, error) {
	matches, best := s.collectSequentialCandidates(query, stubs)
	sortRankedMatches(matches)

	return s.resultFromRankedAndSimilar(query, matches, best.stub)
}

func (s *searcher) processSingleStub(query Query, stub *Stub) (*Result, error) {
	if s.matcher.Match(query, stub) {
		if matchNumber, reserved := s.tryReserve(query, stub); reserved {
			return &Result{found: effectiveDelayStub(stub, matchNumber)}, nil
		}
	}

	return &Result{similar: stub}, nil
}

func (s *searcher) collectSequentialCandidates(query Query, stubs []*Stub) ([]rankedMatch, similarCandidate) {
	var (
		matches []rankedMatch
		best    similarCandidate
	)

	for _, stub := range stubs {
		ranked, matched := s.evaluateRankedMatch(query, stub)

		candidate := s.buildSimilarCandidateFromRanked(stub, ranked)
		if matched {
			matches = append(matches, ranked)
		}

		if betterSimilar(best, candidate) {
			best = candidate
		}
	}

	return matches, best
}

func (s *searcher) reserveFirstRankedMatch(query Query, matches []rankedMatch) *Stub {
	for _, match := range matches {
		if matchNumber, reserved := s.tryReserve(query, match.stub); reserved {
			return effectiveDelayStub(match.stub, matchNumber)
		}
	}

	return nil
}

func (s *searcher) resultFromRankedAndSimilar(query Query, matches []rankedMatch, similar *Stub) (*Result, error) {
	if found := s.reserveFirstRankedMatch(query, matches); found != nil {
		return &Result{found: found}, nil
	}

	if similar != nil {
		return &Result{similar: similar}, nil
	}

	return nil, ErrStubNotFound
}

// processStubsParallel processes stubs in parallel using goroutines.
func (s *searcher) processStubsParallel(query Query, stubs []*Stub) (*Result, error) {
	const chunkSize = 50

	numChunks := (len(stubs) + chunkSize - 1) / chunkSize
	results := make(chan chunkOutcome, numChunks)

	for i := range numChunks {
		start := i * chunkSize

		end := min(start+chunkSize, len(stubs))
		go func(chunkStubs []*Stub) {
			results <- s.processChunk(query, chunkStubs)
		}(stubs[start:end])
	}

	bestMatches, mostSimilar := collectChunkResults(results, numChunks)
	sortRankedMatches(bestMatches)

	return s.resultFromRankedAndSimilar(query, bestMatches, mostSimilar.stub)
}

func (s *searcher) processChunk(query Query, chunkStubs []*Stub) chunkOutcome {
	var (
		bestMatch   rankedMatch
		mostSimilar scoredStub
	)

	for _, stub := range chunkStubs {
		ranked, matched := s.evaluateRankedMatch(query, stub)
		if matched && betterRanked(bestMatch, ranked) {
			bestMatch = ranked
		}

		mostSimilar = pickHigherScore(mostSimilar, scoredStub{stub: stub, score: ranked.totalScore})
	}

	return chunkOutcome{bestMatch: bestMatch, mostSimilar: mostSimilar}
}

func collectChunkResults(results chan chunkOutcome, numChunks int) ([]rankedMatch, scoredStub) {
	var (
		bestMatches = make([]rankedMatch, 0, numChunks)
		bestSimilar scoredStub
	)

	for range numChunks {
		result := <-results
		if result.bestMatch.stub != nil {
			bestMatches = append(bestMatches, result.bestMatch)
		}

		bestSimilar = pickHigherScore(bestSimilar, result.mostSimilar)
	}

	return bestMatches, bestSimilar
}
