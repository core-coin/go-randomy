# Go-randomy

This is a binding for Random-Y based algorithms.

Do NOT use go mod import this.

For better go.mod experience, like directly import go-randomy dep through `go get` or `go build`, check the https://github.com/core-coin/RandomY and https://github.com/core-coin/go-randomy and their GitHub actions.

## Algorithms

- random-y
- random-x
- random-xl
- random-wow
- random-arq
- random-yada

## Build

### Windows

1. Download and install the msys2, then open and install the following components:

Take msys2's pacman for example

```bash
pacman -Syu
pacman -S git mingw64/mingw-w64-x86_64-go mingw64/mingw-w64-x86_64-gcc mingw64/mingw-w64-x86_64-cmake mingw64/mingw-w64-x86_64-make
```

2. Clone this repo to your project folder

```bash
cd MyProject
git clone https://github.com/core-coin/go-randomy
```

3. Run `./build.sh` to auto compile official random-y code

```bash
# clone and compile RandomY source code into librandomy
./build random-y # random-y can be replaced with random-xl random-arq random-wow
```

4. You can use the package as your internal one.

Directly using it with `import "github.com/MyProject/go-randomy"` and then `randomy.AllocCache()` etc.

### Linux

1. Download the latest go from [web](https://golang.org/dl/) and then install it following [this instructions](https://golang.org/doc/install#tarball).

```bash
sudo apt update && sudo apt upgrade
sudo apt install git cmake make gcc build-essential
```

2. Clone this repo to your project folder.

```
cd MyProject
git clone https://github.com/core-coin/go-randomy
```

3. Run `go generate` to auto compile official random-y code.

```bash
# clone and compile RandomY source code into librandomy
./build random-y # random-y can be replaced with random-xl random-arq random-wow
```

4. You can using the package as your internal one.

Directly using it with `import "github.com/myname/my-project/go-randomy"` and then start the functions like `randomy.AllocCache()` etc.

## License

[The 3-Clause BSD License](LICENSE)
