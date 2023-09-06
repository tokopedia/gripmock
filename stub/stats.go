package stub

import (
	"sync"
)

var sttmx = sync.Mutex{}

// below represent map[servicename][methodname][]expectations
type statsMapping map[string]map[string]map[string]int

var statsStorage = statsMapping{}

func updateStats(stub *findStubPayload, status string) {
	statsStorage.updateStats(stub, status)
}

func (sm *statsMapping) updateStats(stub *findStubPayload, status string) {
	sttmx.Lock()
	defer sttmx.Unlock()

	if (*sm)[stub.Service] == nil {
		(*sm)[stub.Service] = make(map[string]map[string]int)
	}
	if (*sm)[stub.Service][stub.Method] == nil {
		(*sm)[stub.Service][stub.Method] = make(map[string]int)
	}
	(*sm)[stub.Service][stub.Method][status] = (*sm)[stub.Service][stub.Method][status] + 1
}

func allStats() statsMapping {
	sttmx.Lock()
	defer sttmx.Unlock()
	return statsStorage
}

func clearStats() {
	sttmx.Lock()
	defer sttmx.Unlock()

	statsStorage = statsMapping{}
}
