# Building
This section contains the instruction on how to build the library and associated wrappers. We first explain how to build the Go library and then further explain the wrapper specific building process. As the Python wrapper is the only wrapper at the moment, this only consists of this wrapper language for now.

## Building the Go library
To build the Go library, you need the dependencies for your system installed. We will go over the needed dependencies for Linux. Afterwards, we explain the basic commands to build the library.

### Dependencies
#### Linux
To build the Go shared library using Linux you need the following dependencies:

- [Go](https://go.dev/doc/install) 1.18 or later
- [Gcc](https://gcc.gnu.org/)
- [GNU Make](https://www.gnu.org/software/make/)
- Dependencies for the Python wrapper if you want to build that as well

### Commands
Before we can begin building the wrapper code, we need to build the Go code as a shared library. This section will tell you how to do so.

To build the shared library for the current platform issue the following command in the root directory:

```bash
make
```

The shared library will be output in `lib/`.

#### Cleaning
To clean build the library and wrapper, issue the following command in the root directory:

```bash
make clean
```

### Note on releases
Releases are build with the go tag "release" (add flag "-tags=release") to bundle the discovery JSON files and embed them in the shared library. See the [make_release](https://codeberg.org/eduVPN/eduvpn-common/src/branch/main/make_release.sh) script on how we bundle the files. A full command without the Makefile to build this library is:

```bash
go build -o lib/libeduvpn_common-${VERSION}.so -tags=release -buildmode=c-shared ./exports
```

## Python wrapper

To build the python wrapper issue the following command (in the root directory of the eduvpn-common project):

```bash
make -C wrappers/python
```

This uses the makefile in `wrappers/python/Makefile` to build the python file into a wheel placed in `wrappers/python/dist/eduvpncommon-[version]-py3-none-[platform].whl`. Where version is the version of the library and platform is your current platform. 

The wheel can be installed with `pip`:

```bash
pip install ./wrappers/python/dist/eduvpncommon-[version]-py3-none-[platform].whl
```

## Notes on building for release

To build for release, make sure you obtain the tarball artifacts in the release (ending with `.tar.xz`) at <https://codeberg.org/eduVPN/eduvpn-common/releases>.

These are signed with minisign and gpg keys, make sure to verify these signatures using the public keys available here: <https://codeberg.org/eduVPN/eduvpn-common/src/branch/main/keys>, they are also available externally:
- <https://app.eduvpn.org/linux/v4/deb/app+linux@eduvpn.org.asc>
- <https://git.sr.ht/~jwijenbergh/python3-eduvpn-common.rpm/tree/main/item/SOURCES/minisign-CA9409316AC93C07.pub>

To build for release, make sure to extract the tarball, and then add `-tags=release` to the `GOFLAGS` environment variable:

```bash
GOFLAGS="-tags=release" make
```

To upload the releases to Codeberg, run:
```bash
./make_release.sh
./upload_release.sh
```

For pre-releases:
```bash
./make_release.sh -p
./upload_release.sh -p
```

### Package formats

We support the following additional package formats: RPM (Linux, Fedora) and Deb (Linux, Debian derivatives)

#### Linux: RPM
The RPM files can be found on a [SourceHut Repo](https://git.sr.ht/~jwijenbergh/python3-eduvpn-common.rpm).These are then build with [builder.rpm](https://codeberg.org/eduVPN/builder.rpm).

#### Linux: Deb
The RPM files can be found on a [SourceHut Repo](https://git.sr.ht/~jwijenbergh/python3-eduvpn-common.deb). These are then build with [nbuilder.deb](https://codeberg.org/eduVPN/nbuilder.deb).
Proceed the build like normally.
