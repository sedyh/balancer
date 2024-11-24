package data

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

const week = 168 * time.Hour

type Progress = func(progress string, percent string, estimate string)

func SlogProgress(name string) Progress {
	return func(progress string, percent string, estimate string) {
		slog.Info(
			name,
			"progress", progress,
			"percent", percent,
			"estimate", estimate,
		)
	}
}

type ProgressReader struct {
	reader    io.Reader
	tick      time.Duration
	last      time.Time
	estimated time.Time
	started   time.Time
	progress  Progress
	size      int
	count     int
}

func NewProgressReader(reader io.Reader, size int, progress Progress) io.Reader {
	return &ProgressReader{
		reader:   reader,
		size:     size,
		progress: progress,
		tick:     5 * time.Second,
		started:  time.Now(),
	}
}

func (r *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.count += n

	if time.Since(r.last) < r.tick {
		return n, err
	}
	r.last = time.Now()

	if r.size == 0 {
		r.progress(bytes(r.count), "", "")
		return n, err
	}

	ratio := float64(r.count) / float64(r.size)
	past := float64(time.Since(r.started))
	if r.count > 0. {
		total := time.Duration(past / ratio)
		if total < week {
			r.estimated = r.started.Add(total)
		}
	}
	r.progress(
		fmt.Sprintf("%s/%s", bytes(r.count), bytes(r.size)),
		fmt.Sprintf("%d%%", r.percent()),
		fmt.Sprint(time.Until(r.estimated).Round(time.Second)),
	)
	return n, err
}

func (r *ProgressReader) percent() int {
	if r.count == 0 {
		return 0
	}
	if r.count >= r.size {
		return 100
	}
	return int(100.0 / (float64(r.size) / float64(r.count)))
}

func bytes(size int) string {
	return strings.ReplaceAll(humanize.IBytes(uint64(size)), " ", "")
}
