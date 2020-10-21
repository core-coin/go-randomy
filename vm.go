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

func RandomX(vm *RandxVm, mutex *sync.Mutex, hash []byte, nonce uint64) ([]byte, []byte, error) {
	// Combine header+nonce into a 64 byte seed
	seed := make([]byte, 40)
	copy(seed, hash)
	binary.LittleEndian.PutUint64(seed[32:], nonce)

	seed = SHA3_512(seed)

	// Start the mix with replicated seed
	mix := make([]uint32, mixBytes/4)
	for i := 0; i < len(mix); i++ {
		mix[i] = binary.LittleEndian.Uint32(seed[i%16*4:])
	}

	// Compress mix
	for i := 0; i < len(mix); i += 4 {
		mix[i/4] = fnv(fnv(fnv(mix[i], mix[i+1]), mix[i+2]), mix[i+3])
	}
	mix = mix[:len(mix)/4]

	digest := make([]byte, 32)
	for i, val := range mix {
		binary.LittleEndian.PutUint32(digest[i*4:], val)
	}
	randXhash, err := randomxhash(vm, mutex, append(seed, digest...))
	if err != nil {
		return []byte{}, []byte{}, err
	}
	return digest, randXhash, nil
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
	cache, err := AllocCache(FlagDefault)
	if nil != err {
		return
	}
	InitCache(cache, key)

	dataset, err := AllocDataset(FlagDefault)
	if nil != err {
		return
	}
	InitDataset(dataset, cache, 0, DatasetItemCount()) // todo: multi core acceleration

	vm, err := CreateVM(cache, dataset, FlagJIT, FlagHardAES, FlagFullMEM)
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
