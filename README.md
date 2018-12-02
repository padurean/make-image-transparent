## Make image transparent

Detects the background color of an opaque image by looking at the color of the 1st pixel, then makes transparent (sets the alpha channel value to 0 for) all the pixels which have the same color as the detected background one (within some tolerance values - see `colorTolerance` and `colorToleranceUniform` variables in [main.go](./main.go#L219)). Saves the output as *PNG*.

### Supported file types:

*jpeg*, *jpg*, *png*, *bmp*, *tiff*, *gif* and *webp*.

### Build

Implemented in [golang](https://golang.org/). To build an executable for your operating system run `go build`.

### Example:

```
/make-image-transparent sample--yellow-on-red--jpg.jpg
```

It also accepts a second (boolean) argument (`true` | `false`). Example:

```
/make-image-transparent sample--grey-on-white--jpg.jpg true
```

If `true` is specified => the image data will also be encoded to a Base64 string and decoded back (this is done just as an example on how to that, in case one needs to work with Base64 encoded images).
Unfortunately this is not supported for *webp* images as the used library only supports decoding *webp* image data from Base64, but it doesn't also support encoding it back to Base64.
