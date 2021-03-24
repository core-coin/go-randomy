package randomx

import (
	"encoding/binary"
	"fmt"
	"sync"

	"golang.org/x/crypto/sha3"
)

const mixBytes = 128 // Width of mix

var (
	emptyVMError = fmt.Errorf("RandomX virtual machine does not exist and cannot create a hash")
)

func RandomX(vm *RandxVm, mutex *sync.Mutex, hash []byte, nonce uint64) ([]byte, error) {
	// Combine header+nonce into a 64 byte seed
	seed := make([]byte, 40)
	copy(seed, hash)
	binary.LittleEndian.PutUint64(seed[32:], nonce)

	seed = SHA3_512(seed)

	randXhash, err := randomxhash(vm, mutex, seed)
	if err != nil {
		return []byte{}, err
	}
	return randXhash, nil
}

func randomxhash(vm *RandxVm, mutex *sync.Mutex, buf []byte) (ret []byte, err error) {
	if vm == nil {
		return []byte{}, emptyVMError
	}
	mutex.Lock()
	ret = vm.Hash(buf)
	mutex.Unlock()
	return
}

type RandxVm struct {
	cache   Cache
	dataset Dataset
	vm      VM
}

func NewRandxVm(key []byte) (ret *RandxVm, err error) {
	cache, err := AllocCache(GetFlags())
	if nil != err {
		return
	}
	InitCache(cache, key)

	dataset, err := AllocDataset(GetFlags())
	if nil != err {
		return
	}
	InitDataset(dataset, cache, 0, DatasetItemCount()) // todo: multi core acceleration

	vm, err := CreateVM(cache, dataset, GetFlags())
	if nil != err {
		return
	}

	ret = &RandxVm{
		cache:   cache,
		dataset: dataset,
		vm:      vm,
	}

	return
}

func NewRandomXVMWithKeyAndMutex() (*RandxVm, *sync.Mutex) {
	key := []byte{53, 54, 55, 56, 57}
	vm, err := NewRandxVm(key)
	if nil != err {
		panic(err)
	}
	return vm, new(sync.Mutex)
}

func (this *RandxVm) Close() {
	DestroyVM(this.vm)
	ReleaseDataset(this.dataset)
	ReleaseCache(this.cache)
}

func (this *RandxVm) Hash(buf []byte) (ret []byte) {
	return CalculateHash(this.vm, buf)
}

// fnv is an algorithm inspired by the FNV hash, which in some cases is used as
// a non-associative substitute for XOR. Note that we multiply the prime with
// the full 32-bit input, in contrast with the FNV-1 spec which multiplies the
// prime with one byte (octet) in turn.
func fnv(a, b uint32) uint32 {
	return a*0x01000193 ^ b
}

//NIPS implementation of SHA3-512
func SHA3_512(data ...[]byte) []byte {
	d := sha3.New512()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}
