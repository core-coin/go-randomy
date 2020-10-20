# go-randomx

This is a binding for Random-X based algorithms.

Do NOT use go mod import this.

**NOTICE**: For better go.mod experience, like direcly import go-randomx dep through `go get` or `go build`, check the https://github.com/core-coin/RandomY and https://github.com/core-coin/go-randomx and their github actions.

## Algorithms

- random-y
- random-x
- random-xl
- random-wow
- random-arq
- random-yada

## Build

### Windows

Firstly download and install the msys2, then open and install the following components:

Take msys2's pacman for example

```bash
pacman -Syu
pacman -S git mingw64/mingw-w64-x86_64-go mingw64/mingw-w64-x86_64-gcc mingw64/mingw-w64-x86_64-cmake mingw64/mingw-w64-x86_64-make
```

Secondly clone this repo to your project folder
```
cd MyProject
git clone https://github.com/core-coin/go-randomx
```

And then run `./build.sh` to auto compile official random-x code
```bash
# clone and compile RandomX source code into librandomx
./build random-x # random-x can be replaced with random-xl random-arq random-wow
```

Finally you can using the package as your internal one.

Directly using it with `import "github.com/MyProject/go-randomx"` and then `randomx.AllocCache()` etc.

### Linux

Take Ubuntu for example

Download the latest go from [here](https://golang.org/dl/) and then install it following [this instruction](https://golang.org/doc/install#tarball)

```bash
sudo apt update && sudo apt upgrade
sudo apt install git cmake make gcc build-essential
```

Secondly clone this repo to your project folder

```
cd MyProject
git clone https://github.com/core-coin/go-randomx
```

And then run `go generate` to auto compile official random-x code

```bash
# clone and compile RandomX source code into librandomx
./build random-x # random-x can be replaced with random-xl random-arq random-wow
```

Finally you can using the package as your internal one.

Directly using it with `import "github.com/myname/my-project/go-randomx"` and then start the functions like `randomx.AllocCache()` etc.

## License

[The 3-Clause BSD License](LICENSE)
