# Go bindings for UndoDB Live Recorder

The _UndoDB Live Recorder_ allows recording of a running process for later playback in the [Undo](https://undo.io) debugger.

There are two packages provided here: **`undolr`** and **`undoex`**.

`undolr` enables starting, stopping and saving recordings from directly within a Go program.

The `undoex` package allows insertion of annotations in to a recording, whether started via `undolr` or by running the program under Live Recorder. Note that the `undoex` package can be safely used when recording is not in use.

Delve support for Undo is current available in a [fork of Delve](https://github.com/undoio/delve). The intention is for support to eventually be merged in to the upstream project.

## Building

Both packages use cgo to work with external libraries. These libraries (and the associated header files) can be obtained from [Undo](https://undo.io).

If the libraries and headers are not accessible via the standard paths then additional environment variables must be set to allow `go` to find them when building:
```sh
CGO_LDFLAGS=-L <path_to_undolr_libraries> -L <path_to_undoex_libraries>
CGO_CFLAGS=-I <path_to_undolr_headers> -I <path_to_undoex_headers>
```

In addition, the libraries will need to be on the library path:
```sh
LD_LIBRARY_PATH=<path_to_undolr_libraries>:<path_to_undoex_libraries>
```
The `LD_LIBRARY_PATH` will also need to be set when run, and the target system will need the relevant library.

## Usage

The following snippet will start recording and insert an annotation. It then stops the recording and saves it in the background.

```go
import (
	"go.undo.io/bindings/undoex"
	"go.undo.io/bindings/undolr"
	"time"
)

func Example_Annotation() {
	// Start a recording
	err := undolr.Start()
	if err != nil {
		panic(err)
	}

	// Add an annotation in the recording
	err = undoex.AnnotationAddInt("Example", "example annotation", 42)
	if err != nil {
		panic(err)
	}

	// Stop the recording process, defering Discard to clean up
	recContext, err := undolr.Stop()
	if err != nil {
		panic(err)
	}
	defer recContext.Discard()

	// We now have a recording context and we can save it as we please.
	// We save it via a goroutine here, so we could do other things while
	// the save continues in the background if we wanted.
	ch := make(chan error, 1)

	go recContext.SaveBackground("recording.undolr", ch)

	select {
	case err = <-ch:
		if err != nil {
			panic(err)
		}
	case <-time.After(time.Second * 30):
		panic("Save hadn't completed after 30 seconds")
	}
}
```

Further examples can be found within each package.
