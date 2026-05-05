## MacOS

### DuckDB

#### Statically linked DuckDB extensions

If you want to build a `emeddingsdb-server` binary (or any other tool that uses this package as a library) for MacOS with support for DuckDB _and_ that has been signed and notarized you will need to compile a custom `libduckdb_bundle.a` library with both the JSON and VSS extensions statically linked. Then you will need to use specify that custom library when building the `emeddingsdb-server` binary. This is because the default behaviour for DuckDB is to load (and cache) extensions on the fly and those extensions will have been signed by someone other than the "team" (you) that notarized the `emeddingsdb-server` binary.

After a fair amount of trial and error this is what I managed to get working. It _should_ work for you but you know how these things end up changing when you're not looking.

_Note: There are [known problems with this process using recent releases of DuckDB](https://github.com/duckdb/duckdb-spatial/issues/794). I am trying to figure them out._

First install both `duckdb` and `vcpkg` from source:

```
$> git clone https://github.com/duckdb/duckdb.git /usr/local/src/duckdb
$> git clone https://github.com/microsoft/vcpkg.git /usr/local/src/vcpkg

$> cd /usr/local/src/duckdb
```

Now copy the `vss.cmake` config file in to the root directory:

```
$> cp .github/config/extensions/vss.cmake ./vss_config.cmake
```

Now edit it to remove the `DONT_LINK` instruction. For example:

```
duckdb_extension_load(vss
        LOAD_TESTS
        GIT_URL https://github.com/duckdb/duckdb-vss
        GIT_TAG c8a4efe05003d8ef6eaad34f5521cf50126c9967
        TEST_DIR test/sql
        APPLY_PATCHES
    )
```

Ensure the following environment variables are set:

```
$> printenv

GEN=ninja
BUILD_VSS=1
BUILD_JSON=1
EXTENSION_CONFIGS=vss_config.cmake
VCPKG_TOOLCHAIN_PATH=/usr/local/src/vcpkg/scripts/buildsystems/vcpkg.cmake
VCPKG_ROOT=/usr/local/src/vcpkg
```

Note the use of the `BUILD_JSON` environment variable. This will bundle the JSON extension which is necessary to use the VSS extension.

Now build the command line tool so you can verify that the VSS (and JSON) extensions are statically linked:

```
$> make

... stuff happens

$> du -h /usr/local/src/duckdb/build/release/duckdb
 43M	/usr/local/src/duckdb/build/release/duckdb
```

Once built, check the installed (and loaded) extensions:

```
$> /usr/local/src/duckdb/build/release/duckdb

DuckDB v1.5.0-dev5476 (Development Version, 1c62e11b82)
Enter ".help" for usage hints.

memory D SELECT extension_name, loaded, installed, install_mode FROM duckdb_extensions() WHERE installed = true;
┌────────────────┬─────────┬───────────┬───────────────────┐
│ extension_name │ loaded  │ installed │   install_mode    │
│    varchar     │ boolean │  boolean  │      varchar      │
├────────────────┼─────────┼───────────┼───────────────────┤
│ core_functions │ true    │ true      │ STATICALLY_LINKED │
│ json           │ true    │ true      │ STATICALLY_LINKED │
│ parquet        │ true    │ true      │ STATICALLY_LINKED │
│ shell          │ true    │ true      │ STATICALLY_LINKED │
│ vss            │ true    │ true      │ STATICALLY_LINKED │
└────────────────┴─────────┴───────────┴───────────────────┘
```

Assuming that the `vss` extension is installed and loaded build DuckDB again as a library:

```
$> make bundle-library

... stuff happens

$> du -h /usr/local/src/duckdb/build/release/libduckdb_bundle.a
 79M	/usr/local/src/duckdb/build/release/libduckdb_bundle.a
```

Apply additional MacOS hoop-jumping, appending the `generated_extension_loader.cpp.o` file to the `libduckdb_bundle.a` file::

```
$> find /usr/local/src/duckdb/build/release -name "generated_extension_loader.cpp.o"
/usr/local/src/duckdb/build/release/extension/CMakeFiles/duckdb_generated_extension_loader.dir/__/codegen/src/generated_extension_loader.cpp.o

$> ar rcs /usr/local/src/duckdb/build/release/libduckdb_bundle.a /usr/local/src/duckdb/build/release/extension/CMakeFiles/duckdb_generated_extension_loader.dir/__/codegen/src/generated_extension_loader.cpp.o
```

Finally rebuild the `embeddingsdb-server` with the customized DuckDB library using the handy `server-bundle` Makefile target (in this repo):

```
$> cd /usr/local/src/go-embeddingsdb
$> mkdir work
$> cp /usr/local/src/duckdb/build/release/libduckdb_bundle.a ./work/

$> make server-bundle
CGO_ENABLED=1 CPPFLAGS="-DDUCKDB_STATIC_BUILD" CGO_LDFLAGS="-L./work -lduckdb_bundle -lc++" \
	go build -tags=duckdb,duckdb_use_static_lib -mod vendor -ldflags="-s -w" \
	-o bin/embeddingsdb-server cmd/server/main.go
```

_Note: You don't have to copy `libduckdb_bundle.a` in to a local `work` folder but this way you don't have remember where it is or what happened to it the next time you clean up your `/usr/local/src` directory. The `work` directory is explicitly excluded from Git checkins in this repository._

### Bleve

#### Bundling with the `libfaiss` libraries (and friends)

If you want to build a `emeddingsdb-server` binary (or any other tool that uses this package as a library) for MacOS with support for Bleve _and_ that has been signed and notarized you will need to distribute your binary bundled copies of the `libfaiss` and `libomp` libraries.

_Remember these are the `libfaiss` libraries that you built to work with Bleve. See [database/README.md#bleve](../database/README.md#bleve) for details._

What follows is not a complete script to build a package installer. Rather it is just the stuff you'll need to make sure is included in whatever package installer script you write to make sure the relevant dependencies are bundled and installed.

First make sure your `.build/pkgroot` tree has the relevant folders where tools and libraries can be found.

```
mkdir -p .build/pkgroot/usr/local/bin
mkdir -p .build/pkgroot/usr/local/lib
```

After building the `embeddingsdb-server` binary use the `codesign` tool to sign it with your Apple developer key.

Then copy the binary in to `.build/pkgroot/usr/local/bin/`, copy the `libfaiss` and `libomp` libraries in to `.build/pkgroot/usr/local/lib/`, change their paths and codesign each (library) with your developer key. Finally make sure the `--install-location` in your `pkg-build` statement is "/". For example:

```
mv ./bin/embeddingsdb-server .build/pkgroot/usr/local/bin/

cp /usr/local/lib/libfaiss_c.dylib .build/pkgroot/usr/local/lib/
cp /usr/local/lib/libfaiss.dylib .build/pkgroot/usr/local/lib/
cp /opt/homebrew/opt/llvm/lib/libomp.dylib .build/pkgroot/usr/local/lib/

install_name_tool -change "/opt/homebrew/opt/llvm/lib/libomp.dylib" "/usr/local/lib/libomp.dylib" .build/pkgroot/usr/local/lib/libfaiss.dylib

codesign --force -s "Developer ID Application: San Francisco International Airport (JB8GZN32RY)" .build/pkgroot/usr/local/lib/libomp.dylib
codesign --force -s "Developer ID Application: San Francisco International Airport (JB8GZN32RY)" .build/pkgroot/usr/local/lib/libfaiss_c.dylib
codesign --force -s "Developer ID Application: San Francisco International Airport (JB8GZN32RY)" .build/pkgroot/usr/local/lib/libfaiss.dylib

pkgbuild \
    --root .build/pkgroot \
    --identifier org.sfomuseum.embeddingsdb-server \
    --version ${VERSION} \
    --install-location /\
    embeddingsdb-server-${ARCH}-${VERSION}.pkg
```

Run the `productbuild` tool and any other steps relevant to building your package installer.