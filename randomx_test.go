package randomx

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"runtime"
	"sync"
	"testing"
	"time"
)

var testPairs = [][][]byte{
	// randomX
	{
		[]byte("test key 000"),
		[]byte("This is a test"),
		[]byte("46b49051978dcce1cd534a4066035184afb16a0591b43522466e10cc2496717e"),
	},
}

func TestAllocCache(t *testing.T) {
	cache, _ := AllocCache(FlagDefault)
	InitCache(cache, []byte("123"))
	ReleaseCache(cache)
}

func TestAllocDataset(t *testing.T) {
	t.Log("warning: cannot use FlagDefault only, very slow!. After using FlagJIT, really fast!")

	ds, err := AllocDataset(FlagDefault)
	if err != nil {
		panic(err)
	}
	cache, err := AllocCache(FlagDefault)
	if err != nil {
		panic(err)
	}

	seed := make([]byte, 32)
	InitCache(cache, seed)
	t.Log("rxCache initialization finished")

	count := DatasetItemCount()
	t.Log("dataset count:", count/1024/1024, "mb")
	InitDataset(ds, cache, 0, count)
	t.Log(GetDatasetMemory(ds))

	ReleaseDataset(ds)
	ReleaseCache(cache)
}

func TestCreateVM(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var tp = testPairs[0]
	cache, _ := AllocCache(FlagDefault)
	t.Log("alloc cache mem finished")
	seed := tp[0]
	InitCache(cache, seed)
	t.Log("cache initialization finished")

	ds, _ := AllocDataset(FlagDefault)
	t.Log("alloc dataset mem finished")
	count := DatasetItemCount()
	t.Log("dataset count:", count)
	var wg sync.WaitGroup
	var workerNum = uint32(runtime.NumCPU())
	t.Log("Here though using FlagDefault, but we use multi-core to accelerate it")
	for i := uint32(0); i < workerNum; i++ {
		wg.Add(1)
		a := (count * i) / workerNum
		b := (count * (i + 1)) / workerNum
		go func() {
			defer wg.Done()
			InitDataset(ds, cache, a, b-a)
		}()
	}
	wg.Wait()
	t.Log("dataset initialization finished") // too slow when one thread
	vm, _ := CreateVM(cache, ds, FlagJIT, FlagHardAES, FlagFullMEM)

	var hashCorrect = make([]byte, hex.DecodedLen(len(tp[2])))
	_, err := hex.Decode(hashCorrect, tp[2])
	if err != nil {
		t.Log(err)
	}

	if bytes.Compare(CalculateHash(vm, tp[1]), hashCorrect) != 0 {
		t.Fail()
	}
}

func TestNewRxVM(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	start := time.Now()
	pair := testPairs[0]
	workerNum := uint32(runtime.NumCPU())

	seed := pair[0]
	t.Log("Here we use FlagJIT, really fast!")
	dataset, _ := NewRxDataset(FlagJIT)
	if dataset.GoInit(seed, workerNum) == false {
		log.Fatal("failed to init dataset")
	}
	//defer dataset.Close()
	fmt.Println("Finished generating dataset in", time.Since(start).Seconds(), "sec")

	vm, _ := NewRxVM(dataset, FlagFullMEM, FlagHardAES, FlagJIT, FlagSecure)
	//defer vm.Close()

	blob := pair[1]
	hash := vm.CalcHash(blob)

	var hashCorrect = make([]byte, hex.DecodedLen(len(pair[2])))
	_, err := hex.Decode(hashCorrect, pair[2])
	if err != nil {
		t.Log(err)
	}

	if bytes.Compare(hash, hashCorrect) != 0 {
		t.Logf("%x", hash)
		t.Fail()
	}
}

func TestCalculateHashFirst(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	start := time.Now()
	pair := testPairs[0]
	workerNum := uint32(runtime.NumCPU())

	seed := pair[0]
	dataset, _ := NewRxDataset(FlagJIT)
	if dataset.GoInit(seed, workerNum) == false {
		log.Fatal("failed to init dataset")
	}
	//defer dataset.Close()
	fmt.Println("Finished generating dataset in", time.Since(start).Seconds(), "sec")
	vm, _ := NewRxVM(dataset, FlagFullMEM, FlagHardAES, FlagJIT, FlagSecure)
	//defer vm.Close()

	targetBlob := make([]byte, 76)
	targetNonce := make([]byte, 4)
	binary.LittleEndian.PutUint32(targetNonce, 2333)
	copy(targetBlob[39:43], targetNonce)

	targetResult := vm.CalcHash(targetBlob)

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		vm, _ := NewRxVM(dataset, FlagFullMEM, FlagHardAES, FlagJIT, FlagSecure)

		wg.Add(1)
		blob := make([]byte, 76)
		vm.CalcHashFirst(blob)

		n := uint32(0)
		go func() {
			for {
				n++
				nonce := make([]byte, 4)
				binary.LittleEndian.PutUint32(nonce, n)
				copy(blob[39:43], nonce)
				result := vm.CalcHashNext(blob)
				if bytes.Compare(result, targetResult) == 0 {
					fmt.Println(n, "found")
					wg.Done()
				} else {
					//fmt.Println(n, "failed")
				}
			}
		}()
	}
	wg.Wait()

}

// go test -v -run=^$ -benchtime=1m  -timeout 20m -bench=.
func BenchmarkCalculateHashDefault(b *testing.B) {
	cache, _ := AllocCache(FlagDefault)
	ds, _ := AllocDataset(FlagDefault)
	InitCache(cache, []byte("123"))
	FastInitFullDataset(ds, cache, uint32(runtime.NumCPU()))
	vm, _ := CreateVM(cache, ds, FlagDefault)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CalculateHash(vm, []byte("123"))
	}

	DestroyVM(vm)
}

func BenchmarkCalculateHashJIT(b *testing.B) {
	cache, _ := AllocCache(FlagDefault | FlagJIT)
	ds, _ := AllocDataset(FlagDefault | FlagJIT)
	InitCache(cache, []byte("123"))
	FastInitFullDataset(ds, cache, uint32(runtime.NumCPU()))
	vm, _ := CreateVM(cache, ds, FlagDefault)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CalculateHash(vm, []byte("123"))
	}

	DestroyVM(vm)
}

func BenchmarkCalculateHashFullMEM(b *testing.B) {
	cache, _ := AllocCache(FlagDefault | FlagFullMEM)
	ds, _ := AllocDataset(FlagDefault | FlagFullMEM)
	InitCache(cache, []byte("123"))
	FastInitFullDataset(ds, cache, uint32(runtime.NumCPU()))
	vm, _ := CreateVM(cache, ds, FlagDefault)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CalculateHash(vm, []byte("123"))
	}

	DestroyVM(vm)
}

func BenchmarkCalculateHashHardAES(b *testing.B) {
	cache, _ := AllocCache(FlagDefault | FlagHardAES)
	ds, _ := AllocDataset(FlagDefault | FlagHardAES)
	InitCache(cache, []byte("123"))
	FastInitFullDataset(ds, cache, uint32(runtime.NumCPU()))
	vm, _ := CreateVM(cache, ds, FlagDefault)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CalculateHash(vm, []byte("123"))
	}

	DestroyVM(vm)
}

func BenchmarkCalculateHashAll(b *testing.B) {
	cache, _ := AllocCache(FlagDefault | FlagArgon2 | FlagArgon2AVX2 | FlagArgon2SSSE3 | FlagFullMEM | FlagHardAES | FlagJIT) // without lagePage to avoid panic
	ds, _ := AllocDataset(FlagDefault | FlagArgon2 | FlagArgon2AVX2 | FlagArgon2SSSE3 | FlagFullMEM | FlagHardAES | FlagJIT)
	InitCache(cache, []byte("123"))
	FastInitFullDataset(ds, cache, uint32(runtime.NumCPU()))
	vm, _ := CreateVM(cache, ds, FlagDefault)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CalculateHash(vm, []byte("123"))
	}

	DestroyVM(vm)
}
