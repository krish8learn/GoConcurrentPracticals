package fixedpool

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io/fs"
	"os"
	"path/filepath"
)

type Resource struct {
	src  string
	dest string
}

// Center creates destFile which is the center of image encode in data.
func Center(srcFile, destFile string) error {
	file, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer file.Close()

	src, err := jpeg.Decode(file)
	if err != nil {
		return err
	}

	x, y := src.Bounds().Max.X, src.Bounds().Max.Y
	r := image.Rect(0, 0, x/2, y/2)
	dest := image.NewRGBA(r)
	draw.Draw(dest, dest.Bounds(), src, image.Point{x / 4, y / 4}, draw.Over)

	out, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, dest, nil)
}

// CenterDir calls Center on every image in srcDir. n is the maximal number of goroutines.
func CenterDir(ctx context.Context, srcDir, destDir string, n int) error {
	if err := os.Mkdir(destDir, 0750); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}

	matches, err := filepath.Glob(fmt.Sprintf("%s/*.jpg", srcDir))
	if err != nil {
		return err
	}

	// for _, src := range matches {
	// 	dest := fmt.Sprintf("%s/%s", destDir, filepath.Base(src)) // this is producer
	// 	if err := Center(src, dest); err != nil {                 // this is worker
	// 		return err
	// 	}
	// }

	in, out := make(chan Resource), make(chan error, len(matches))
	// keep the worker ready for processing
	for index := 0; index < n; index++ {
		go Worker(ctx, in, out)
	}

	go Producer(ctx, in, matches, destDir)

	// get the result from channels
	for range matches {
		select {
		case err := <-out:
			if err != nil {
				return err
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func Producer(ctx context.Context, in chan<- Resource, srcFiles []string, destDir string) {
	defer close(in)

	for _, src := range srcFiles {
		dest := fmt.Sprintf("%s/%s", destDir, filepath.Base(src))
		select {
		case in <- Resource{
			src:  src,
			dest: dest,
		}:
		case <-ctx.Done():
			return
		}
	}
}

func Worker(ctx context.Context, in <-chan Resource, out chan<- error) {
	for {
		select {
		case r, ok := <-in:
			if !ok {
				return
			} else {
				out <- Center(r.src, r.dest)
			}
		case <-ctx.Done():
			return
		}
	}
}
