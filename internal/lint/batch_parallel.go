package lint

import (
	"sync"

	"github.com/dotcommander/cclint/internal/discovery"
)

// lintBatchParallel lints files concurrently with a worker pool of
// min(ctx.Concurrency, len(files)) workers. Results are returned in input
// order so downstream output is byte-identical to the sequential path.
//
// The per-file pipeline is safe for concurrent use: the CUE validator
// serializes access behind its own mutex and every other dependency is
// read-only during linting.
func lintBatchParallel(ctx *LinterContext, files []discovery.File, linter ComponentLinter) []LintResult {
	workers := ctx.Concurrency
	if workers > len(files) {
		workers = len(files)
	}
	if workers < 1 {
		workers = 1
	}

	results := make([]LintResult, len(files))
	indexes := make(chan int)

	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			// Exit condition: indexes is closed by the producer below after
			// every index has been dispatched.
			defer wg.Done()
			for i := range indexes {
				results[i] = lintBatchFile(ctx, files[i], linter)
			}
		}()
	}
	for i := range files {
		indexes <- i
	}
	close(indexes) // releases every worker goroutine
	wg.Wait()

	return results
}
