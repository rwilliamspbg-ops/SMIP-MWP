package afxdp

import (
	"context"
	"runtime"
)

// SpawnPerCPUWorkers starts numWorkers goroutines, each locked to its
// OS thread, calls SetCurrentThreadAffinity(workerID) (no-op on unsupported
// builds), and runs the provided workerFunc(ctx, workerID). The caller is
// responsible for cancelling the context to stop workers.
func SpawnPerCPUWorkers(ctx context.Context, numWorkers int, workerFunc func(context.Context, int)) {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}
	for i := 0; i < numWorkers; i++ {
		id := i
		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			// attempt to set affinity; ignore errors here — it's advisory
			_ = SetCurrentThreadAffinity(id)
			workerFunc(ctx, id)
		}()
	}
}
