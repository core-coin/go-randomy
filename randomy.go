package randomy

//#cgo CFLAGS: -I./randomx
//#cgo LDFLAGS: -lrandomx -lstdc++
//#cgo linux,amd64 LDFLAGS:-L${SRCDIR}/build/linux-x86_64 -lm
//#cgo linux,arm64 LDFLAGS:-L${SRCDIR}/build/linux-arm64 -lm
//#cgo darwin,amd64 LDFLAGS:-L${SRCDIR}/build/macos-x86_64 -lm
//#cgo windows,amd64 LDFLAGS:-L${SRCDIR}/build/windows-x86_64
//#include <stdlib.h>
//#include "randomy.h"
import "C"
import (
	"errors"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/cpu"
)

const RxHashSize = C.RANDOMX_HASH_SIZE

// All flags
const (
	FlagDefault     C.randomx_flags = 8 + 4 // for all default
	FlagLargePages  C.randomx_flags = 1     // for dataset & rxCache & vm
	FlagHardAES     C.randomx_flags = 2     // for vm
	FlagFullMEM     C.randomx_flags = 4     // for vm
	FlagJIT         C.randomx_flags = 8     // for vm & cache
	FlagSecure      C.randomx_flags = 16
	FlagArgon2SSSE3 C.randomx_flags = 32 // for cache
	FlagArgon2AVX2  C.randomx_flags = 64 // for cache
	FlagArgon2      C.randomx_flags = 96 // = avx2 + sse3
)

type Cache *C.randomx_cache

type Dataset *C.randomx_dataset

type VM *C.randomx_vm

func hasAES() bool {
	switch runtime.GOARCH {
	case "amd64":
		return cpu.X86.HasAES
	case "arm64":
		return cpu.ARM64.HasAES
	default:
		panic("unsupported arch")
	}
}

func GetFlags() C.randomx_flags {
	if hasAES() {
		return FlagDefault + FlagHardAES
	}
	return FlagDefault
}
func AllocCache(flags ...C.randomx_flags) (Cache, error) {
	var SumFlag = GetFlags()
	var cache *C.randomx_cache

	for _, flag := range flags {
		SumFlag = SumFlag | flag
	}

	cache = C.randomx_alloc_cache(SumFlag)
	if cache == nil {
		return nil, errors.New("failed to alloc mem for rxCache")
	}

	return cache, nil
}

func InitCache(cache Cache, seed []byte) {
	if len(seed) == 0 {
		panic("seed cannot be NULL")
	}

	C.randomx_init_cache(cache, unsafe.Pointer(&seed[0]), C.size_t(len(seed)))
}

func ReleaseCache(cache Cache) {
	C.randomx_release_cache(cache)
}

func AllocDataset(flags ...C.randomx_flags) (Dataset, error) {
	var SumFlag = GetFlags()
	for _, flag := range flags {
		SumFlag = SumFlag | flag
	}

	var dataset *C.randomx_dataset
	dataset = C.randomx_alloc_dataset(SumFlag)
	if dataset == nil {
		return nil, errors.New("failed to alloc mem for dataset")
	}

	return dataset, nil
}

func DatasetItemCount() uint32 {
	var length C.ulong
	length = C.randomx_dataset_item_count()
	return uint32(length)
}

func InitDataset(dataset Dataset, cache Cache, startItem uint32, itemCount uint32) {
	if dataset == nil {
		panic("alloc dataset mem is required")
	}

	if cache == nil {
		panic("alloc cache mem is required")
	}
	goroutinesCount := uint32(128)

	perGoroutine := itemCount / goroutinesCount
	remainder := itemCount % goroutinesCount
	var wg sync.WaitGroup

	for i := uint32(0); i < goroutinesCount; i++ {
		count := uint32(0)
		wg.Add(1)
		if i == goroutinesCount-1 {
			count = perGoroutine + remainder
		} else {
			count = perGoroutine
		}
		go func(start, end uint32) {
			C.randomx_init_dataset(dataset, cache, C.ulong(start), C.ulong(end))
			wg.Done()
		}(startItem, count)
		startItem += count
	}
	wg.Wait()
}

func GetDatasetMemory(dataset Dataset) unsafe.Pointer {
	return C.randomx_get_dataset_memory(dataset)
}

func ReleaseDataset(dataset Dataset) {
	C.randomx_release_dataset(dataset)
}

func CreateVM(cache Cache, dataset Dataset, flags ...C.randomx_flags) (VM, error) {
	var SumFlag = GetFlags()
	for _, flag := range flags {
		SumFlag = SumFlag | flag
	}

	if dataset == nil {
		panic("failed creating vm: using empty dataset")
	}

	vm := C.randomx_create_vm(SumFlag, cache, dataset)

	if vm == nil {
		return nil, errors.New("failed to create vm")
	}

	return vm, nil
}

func SetVMCache(vm VM, cache Cache) {
	C.randomx_vm_set_cache(vm, cache)
}

func SetVMDataset(vm VM, dataset Dataset) {
	C.randomx_vm_set_dataset(vm, dataset)
}

func DestroyVM(vm VM) {
	C.randomx_destroy_vm(vm)
}

func CalculateHash(vm VM, in []byte) []byte {
	if vm == nil {
		panic("failed hashing: using empty vm")
	}

	input := C.CBytes(in)
	output := C.CBytes(make([]byte, RxHashSize))
	C.randomx_calculate_hash(vm, input, C.size_t(len(in)), output)
	hash := C.GoBytes(output, RxHashSize)
	C.free(unsafe.Pointer(input))
	C.free(unsafe.Pointer(output))

	return hash
}

func CalculateHashFirst(vm VM, in []byte) {
	if vm == nil {
		panic("failed hashing: using empty vm")
	}

	input := C.CBytes(in)
	C.randomx_calculate_hash_first(vm, input, C.size_t(len(in)))
	C.free(unsafe.Pointer(input))
}

func CalculateHashNext(vm VM, in []byte) []byte {
	if vm == nil {
		panic("failed hashing: using empty vm")
	}

	input := C.CBytes(in)
	output := C.CBytes(make([]byte, RxHashSize))
	C.randomx_calculate_hash_next(vm, input, C.size_t(len(in)), output)
	hash := C.GoBytes(output, RxHashSize)
	C.free(unsafe.Pointer(input))
	C.free(unsafe.Pointer(output))

	return hash
}
