package randomy_test

import (
	"bytes"
	"encoding/hex"
	"runtime"
	"sync"
	"testing"

	"github.com/core-coin/go-randomy"
)

var testPairs = [][][]byte{
	// randomX
	{
		[]byte("test key 000"),
		[]byte("This is a test"),
		[]byte("0e9ab13e4b337c866e17445621be8d180c0a18a325932078dc33f7ac69ff17e8"),
	},
}

func TestAllocCache(t *testing.T) {
	cache, _ := randomy.AllocCache(randomy.GetFlags())
	randomy.InitCache(cache, []byte("123"))
	randomy.ReleaseCache(cache)
}

func TestAllocDataset(t *testing.T) {
	t.Log("warning: cannot use GetFlags() only, very slow!. After using FlagJIT, really fast!")

	ds, err := randomy.AllocDataset(randomy.FlagJIT)
	if err != nil {
		panic(err)
	}
	cache, err := randomy.AllocCache(randomy.FlagJIT)
	if err != nil {
		panic(err)
	}

	seed := make([]byte, 32)
	randomy.InitCache(cache, seed)
	t.Log("rxCache initialization finished")

	count := randomy.DatasetItemCount()
	t.Log("dataset count:", count/1024/1024, "mb")
	randomy.InitDataset(ds, cache, 0, count)
	t.Log(randomy.GetDatasetMemory(ds))

	randomy.ReleaseDataset(ds)
	randomy.ReleaseCache(cache)
}

func TestCreateVM(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var tp = testPairs[0]
	cache, _ := randomy.AllocCache(randomy.GetFlags())
	t.Log("alloc cache mem finished")
	seed := tp[0]
	randomy.InitCache(cache, seed)
	t.Log("cache initialization finished")

	ds, _ := randomy.AllocDataset(randomy.GetFlags())
	t.Log("alloc dataset mem finished")
	count := randomy.DatasetItemCount()
	t.Log("dataset count:", count)
	var wg sync.WaitGroup
	var workerNum = uint32(runtime.NumCPU())
	t.Log("Here though using GetFlags(), but we use multi-core to accelerate it")
	for i := uint32(0); i < workerNum; i++ {
		wg.Add(1)
		a := (count * i) / workerNum
		b := (count * (i + 1)) / workerNum
		go func() {
			defer wg.Done()
			randomy.InitDataset(ds, cache, a, b-a)
		}()
	}
	wg.Wait()
	t.Log("dataset initialization finished") // too slow when one thread
	vm, _ := randomy.CreateVM(cache, ds, randomy.GetFlags())

	var hashCorrect = make([]byte, hex.DecodedLen(len(tp[2])))
	_, err := hex.Decode(hashCorrect, tp[2])
	if err != nil {
		t.Log(err)
	}

	hash := randomy.CalculateHash(vm, tp[1])
	if !bytes.Equal(hash, hashCorrect) {
		t.Logf("answer is incorrect: %x, %x", hash, hashCorrect)
		t.Fail()
	}
}

// go test -v -run=^$ -benchtime=1m  -timeout 20m -bench=.
func BenchmarkCalculateHashDefault(b *testing.B) {
	cache, _ := randomy.AllocCache(randomy.GetFlags())
	ds, _ := randomy.AllocDataset(randomy.GetFlags())
	randomy.InitCache(cache, []byte("123"))
	count := randomy.DatasetItemCount()
	var wg sync.WaitGroup
	var workerNum = uint32(runtime.NumCPU())
	for i := uint32(0); i < workerNum; i++ {
		wg.Add(1)
		a := (count * i) / workerNum
		b := (count * (i + 1)) / workerNum
		go func() {
			defer wg.Done()
			randomy.InitDataset(ds, cache, a, b-a)
		}()
	}
	wg.Wait()
	vm, _ := randomy.CreateVM(cache, ds, randomy.GetFlags())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomy.CalculateHash(vm, []byte("123"))
	}

	randomy.DestroyVM(vm)
}

func BenchmarkCalculateHashJIT(b *testing.B) {
	cache, _ := randomy.AllocCache(randomy.GetFlags(), randomy.FlagJIT)
	ds, _ := randomy.AllocDataset(randomy.GetFlags(), randomy.FlagJIT)
	randomy.InitCache(cache, []byte("123"))
	count := randomy.DatasetItemCount()
	var wg sync.WaitGroup
	var workerNum = uint32(runtime.NumCPU())
	for i := uint32(0); i < workerNum; i++ {
		wg.Add(1)
		a := (count * i) / workerNum
		b := (count * (i + 1)) / workerNum
		go func() {
			defer wg.Done()
			randomy.InitDataset(ds, cache, a, b-a)
		}()
	}
	wg.Wait()
	vm, _ := randomy.CreateVM(cache, ds, randomy.GetFlags())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomy.CalculateHash(vm, []byte("123"))
	}

	randomy.DestroyVM(vm)
}

func BenchmarkCalculateHashFullMEM(b *testing.B) {
	cache, _ := randomy.AllocCache(randomy.GetFlags(), randomy.FlagFullMEM)
	ds, _ := randomy.AllocDataset(randomy.GetFlags(), randomy.FlagFullMEM)
	randomy.InitCache(cache, []byte("123"))
	count := randomy.DatasetItemCount()
	var wg sync.WaitGroup
	var workerNum = uint32(runtime.NumCPU())
	for i := uint32(0); i < workerNum; i++ {
		wg.Add(1)
		a := (count * i) / workerNum
		b := (count * (i + 1)) / workerNum
		go func() {
			defer wg.Done()
			randomy.InitDataset(ds, cache, a, b-a)
		}()
	}
	wg.Wait()
	vm, _ := randomy.CreateVM(cache, ds, randomy.GetFlags())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomy.CalculateHash(vm, []byte("123"))
	}

	randomy.DestroyVM(vm)
}

func BenchmarkCalculateHashHardAES(b *testing.B) {
	cache, _ := randomy.AllocCache(randomy.GetFlags(), randomy.FlagHardAES)
	ds, _ := randomy.AllocDataset(randomy.GetFlags(), randomy.FlagHardAES)
	randomy.InitCache(cache, []byte("123"))
	count := randomy.DatasetItemCount()
	var wg sync.WaitGroup
	var workerNum = uint32(runtime.NumCPU())
	for i := uint32(0); i < workerNum; i++ {
		wg.Add(1)
		a := (count * i) / workerNum
		b := (count * (i + 1)) / workerNum
		go func() {
			defer wg.Done()
			randomy.InitDataset(ds, cache, a, b-a)
		}()
	}
	wg.Wait()
	vm, _ := randomy.CreateVM(cache, ds, randomy.GetFlags())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomy.CalculateHash(vm, []byte("123"))
	}

	randomy.DestroyVM(vm)
}

func BenchmarkCalculateHashAll(b *testing.B) {
	cache, _ := randomy.AllocCache(randomy.GetFlags(), randomy.FlagArgon2, randomy.FlagArgon2AVX2, randomy.FlagArgon2SSSE3, randomy.FlagFullMEM, randomy.FlagHardAES, randomy.FlagJIT) // without lagePage to avoid panic
	ds, _ := randomy.AllocDataset(randomy.GetFlags(), randomy.FlagArgon2, randomy.FlagArgon2AVX2, randomy.FlagArgon2SSSE3, randomy.FlagFullMEM, randomy.FlagHardAES, randomy.FlagJIT)
	randomy.InitCache(cache, []byte("123"))
	count := randomy.DatasetItemCount()
	var wg sync.WaitGroup
	var workerNum = uint32(runtime.NumCPU())
	for i := uint32(0); i < workerNum; i++ {
		wg.Add(1)
		a := (count * i) / workerNum
		b := (count * (i + 1)) / workerNum
		go func() {
			defer wg.Done()
			randomy.InitDataset(ds, cache, a, b-a)
		}()
	}
	wg.Wait()
	vm, _ := randomy.CreateVM(cache, ds, randomy.GetFlags())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomy.CalculateHash(vm, []byte("123"))
	}

	randomy.DestroyVM(vm)
}
