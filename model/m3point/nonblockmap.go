package m3point

import (
	"github.com/freddy33/qsm-go/m3util"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

type MurmurKey interface {
	MurmurHash() uint32
}

type MurmurHashMap interface {
	Has(key MurmurKey) bool
	Load(key MurmurKey) (value unsafe.Pointer, loaded bool)
	Store(key MurmurKey, value unsafe.Pointer) unsafe.Pointer
	LoadOrStore(key MurmurKey, value unsafe.Pointer) (actualValue unsafe.Pointer, inserted bool)
	Delete(key interface{}) unsafe.Pointer
	Size() int
	InitSize() int
}

type hashMapEntry struct {
	mKey  uint32
	key   MurmurKey
	value unsafe.Pointer
	next  *hashMapEntry
}

type NonBlockConcurrentMap struct {
	nbElements     int32
	nbEntries      int
	entries        []*hashMapEntry
	mHashConflicts int
}

type RangeContext struct {
	async         bool
	nbProc        int
	isListening   bool
	firstError    error
	collectErrors chan error
	done          chan int32
	Wg            *sync.WaitGroup
	canceled      bool
	logger        m3util.Logger
	nbDone        int32
}

/***************************************************************/
// RangeContext Functions
/***************************************************************/

func MakeRangeContext(async bool, nbProc int, logger m3util.Logger) *RangeContext {
	if nbProc <= 0 {
		nbProc = runtime.NumCPU()
	}
	res := &RangeContext{
		async:         async,
		nbProc:        nbProc,
		isListening:   false,
		firstError:    nil,
		collectErrors: make(chan error),
		done:          make(chan int32),
		Wg:            &sync.WaitGroup{},
		canceled:      false,
		logger:        logger,
		nbDone:        int32(0),
	}
	return res

}

func (ec *RangeContext) SetLogger(logger m3util.Logger) {
	ec.logger = logger
}

func (ec *RangeContext) IsAsync() bool {
	return ec.async
}

func (ec *RangeContext) Wait() {
	ec.Wg.Wait()
}

func (ec *RangeContext) SendError(err error) {
	ec.collectErrors <- err
}

func (ec *RangeContext) ErrorChan() <-chan error {
	return ec.collectErrors
}

func (ec *RangeContext) Done() <-chan int32 {
	return ec.done
}

func (ec *RangeContext) incrementDone() {
	atomic.AddInt32(&ec.nbDone, 1)
}

func (ec *RangeContext) Reset() {
	ec.canceled = false
	ec.firstError = nil
	ec.isListening = false
	close(ec.done)
	close(ec.collectErrors)
	ec.collectErrors = make(chan error)
	ec.done = make(chan int32)
	ec.Wg = &sync.WaitGroup{}
}

func (ec *RangeContext) Close() {
	close(ec.done)
	close(ec.collectErrors)
}

func (ec *RangeContext) GetFirstError() error {
	return ec.firstError
}

func (ec *RangeContext) IsCancel() bool {
	return ec.canceled
}

func (ec *RangeContext) Cancel() {
	ec.canceled = true
	close(ec.done)
}

func (ec *RangeContext) ReceivedError(err error) {
	if ec.firstError == nil {
		ec.firstError = err
	}
	ec.logger.Error(err)
}

func (ec *RangeContext) ReceivedDone(nbDone int32, ok bool) {
	if !ok {
		// All done
		if ec.logger.IsTrace() {
			ec.logger.Tracef("In listen finished")
		}
	} else {
		if ec.logger.IsTrace() {
			ec.logger.Tracef("In listen done %d", nbDone)
		}
		atomic.AddInt32(&ec.nbDone, nbDone)
	}
}

func (ec *RangeContext) stoppedListening() {
	ec.isListening = false
}

func (ec *RangeContext) Listen() {
	if ec.isListening {
		return
	}
	defer ec.stoppedListening()
	ec.isListening = true
	for {
		select {
		case err, ok := <-ec.collectErrors:
			if ok {
				ec.ReceivedError(err)
			}
		case nbDone, ok := <-ec.done:
			ec.ReceivedDone(nbDone, ok)
			return
		}
	}
}

/***************************************************************/
// NonBlockConcurrentMap Functions
/***************************************************************/

func MakeNonBlockConcurrentMap(initSize int) *NonBlockConcurrentMap {
	result := new(NonBlockConcurrentMap)
	result.entries = make([]*hashMapEntry, initSize)
	result.nbEntries = len(result.entries)
	result.nbElements = 0
	return result
}

func (n *NonBlockConcurrentMap) InitSize() int {
	return n.nbEntries
}

func (n *NonBlockConcurrentMap) Has(key MurmurKey) bool {
	hashIdx := murmurHashToInt(key.MurmurHash(), n.nbEntries)
	entry := n.entries[hashIdx]
	for {
		if entry == nil {
			return false
		}
		if entry.key == key {
			return true
		}
		entry = entry.next
	}
}

func (n *NonBlockConcurrentMap) Load(key MurmurKey) (unsafe.Pointer, bool) {
	hashIdx := murmurHashToInt(key.MurmurHash(), n.nbEntries)
	entry := n.entries[hashIdx]
	for {
		if entry == nil {
			return nil, false
		}
		if entry.key == key {
			return entry.value, true
		}
		entry = entry.next
	}
}

func (n *NonBlockConcurrentMap) Store(key MurmurKey, value unsafe.Pointer) unsafe.Pointer {
	oldValue, _ := n.internalPut(key, value, true)
	return oldValue
}

func (n *NonBlockConcurrentMap) LoadOrStore(key MurmurKey, value unsafe.Pointer) (actual unsafe.Pointer, inserted bool) {
	return n.internalPut(key, value, false)
}

func (n *NonBlockConcurrentMap) internalPut(key MurmurKey, value unsafe.Pointer, overrideValue bool) (actualOrOld unsafe.Pointer, inserted bool) {
	mKey := key.MurmurHash()
	hashIdx := murmurHashToInt(mKey, n.nbEntries)
	newEntry := &hashMapEntry{
		mKey:  mKey,
		key:   key,
		value: value,
		next:  nil,
	}
	tries := 0
	for {
		actualOrOld, inserted, success := n.internalPutWithHash(hashIdx, newEntry, overrideValue)
		if success {
			return actualOrOld, inserted
		}
		tries++
		if tries > 10 {
			Log.Errorf("Did not managed to insert %v after %d tries", key, tries)
			return unsafe.Pointer(nil), false
		}
	}
}

func (n *NonBlockConcurrentMap) internalPutWithHash(hashIdx int, newEntry *hashMapEntry, overrideValue bool) (actualOrOld unsafe.Pointer, inserted bool, success bool) {
	entry := n.entries[hashIdx]
	entryAddr := (*unsafe.Pointer)(unsafe.Pointer(&n.entries[hashIdx]))
	for {
		if entry == nil {
			success := atomic.CompareAndSwapPointer(entryAddr, unsafe.Pointer(nil), unsafe.Pointer(newEntry))
			if !success {
				return nil, false, false
			} else {
				atomic.AddInt32(&n.nbElements, 1)
				if overrideValue {
					return nil, true, true
				} else {
					return newEntry.value, true, true
				}
			}
		} else {
			if entry.mKey == newEntry.mKey {
				if entry.key == newEntry.key {
					if overrideValue {
						oldValue := entry.value
						success := atomic.CompareAndSwapPointer(&entry.value, entry.value, newEntry.value)
						if !success {
							return nil, false, false
						} else {
							return oldValue, true, true
						}
					} else {
						return entry.value, false, true
					}
				} else {
					n.mHashConflicts++
				}
			}
			entryAddr = (*unsafe.Pointer)(unsafe.Pointer(&entry.next))
			entry = entry.next
		}
	}
}

func (n *NonBlockConcurrentMap) Delete(key interface{}) unsafe.Pointer {
	panic("implement me")
}

func (n *NonBlockConcurrentMap) internalDelete(key interface{}) (unsafe.Pointer, bool) {
	panic("implement me")
}

func (n *NonBlockConcurrentMap) Size() int {
	return int(n.nbElements)
}

func (n *NonBlockConcurrentMap) Range(visit func(key MurmurKey, value unsafe.Pointer) bool, rc *RangeContext) {
	if rc == nil {
		Log.Fatalf("cannot execute range visit without range context")
		return
	}

	if !rc.isListening {
		go rc.Listen()
	}

	nbGoRoutines := rc.nbProc
	var elPerRoutines int
	if n.nbEntries < nbGoRoutines {
		nbGoRoutines = n.nbEntries
		elPerRoutines = 1
	} else {
		elPerRoutines = n.nbEntries / nbGoRoutines
	}
	Log.Debugf("Running range for %d entries with %d go routines and %d elements per routine",
		n.nbEntries, nbGoRoutines, elPerRoutines)
	for routineIdx := 0; routineIdx < nbGoRoutines; routineIdx++ {
		startIdx := elPerRoutines * routineIdx
		endIdx := startIdx + elPerRoutines
		if routineIdx == nbGoRoutines-1 {
			endIdx = n.nbEntries
		}
		rc.Wg.Add(1)
		go func() {
			defer rc.Wg.Done()
			i := startIdx
			entry := n.entries[i]
			for {
				select {
				case nbDone, ok := <-rc.done:
					rc.ReceivedDone(nbDone, ok)
					return
				default:
					if entry == nil {
						i++
						if i >= endIdx {
							return
						}
						entry = n.entries[i]
					} else {
						if visit(entry.key, entry.value) {
							rc.Cancel()
							return
						}
						entry = entry.next
						rc.incrementDone()
					}
				}
			}
		}()
	}

	if !rc.async {
		rc.Wait()
	}
}
