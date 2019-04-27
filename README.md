# Computers in a Nutshell

This is my entry for [Ludum Dare](https://ldjam.com/) 44, the theme is "Your Life is Currency".

You can find the game in [the downloads](https://github.com/gonutz/ld44/releases).

# Build

To build the game yourself you need [Git](https://git-scm.com/) and [the Go programming language](https://golang.org/) installed.

Currently the game runs on Windows only. It can however be cross-compiled from any OS supported by Go.

I am using Go 1.10 for this project so there is no module support and I have not tested this with later versions of Go. They probably work as well, if not please create an issue on [Github](https://github.com/gonutz/ld44).

Run `go get github.com/gonutz/ld44` to download the game. If you are on Windows, `go get` will build and install the game to your `%GOPATH%\bin` folder and you can run it as `ld44.exe`. However if you want to work with the code, go to your `%GOPATH%\src\github.com\gonutz\ld44` folder and run the `build.bat` build script.

All generated files are included in the repo but if you want to change any assets and rebuild them, there is a script `bake-resources.bat` that will do this using some extra Go tools. The script will `go get` these tools so if you do not yet have them, it will download and install them. If you do have them already, they will be installed from disk. If this causes any errors you might need to run a `go get -u` on the failing tool, it might have been updated in an incompatible way.
