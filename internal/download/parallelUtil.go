package download

import(
	"runtime"
)

func (h *DownloadHandler) calculateOptimalChunkSize(contentLength int64) int64{
	const (
		minChunkSize = 1024 * 1024   // 1MB minimum chunk size
		maxChunkSize = 10 * 1024 * 1024 // 10MB maximum chunk size
	)

	if contentLength < minChunkSize {
		h.CHUNK_SIZE = contentLength 
	} else {
		targetParts := contentLength / minChunkSize
		if targetParts < 4 {
			targetParts = 4 // making sure we have at least 4 chunks for worker distribution
		}
		h.CHUNK_SIZE = contentLength / targetParts

		// bounding it 
		if h.CHUNK_SIZE > maxChunkSize {
			h.CHUNK_SIZE = maxChunkSize 
		}
		if h.CHUNK_SIZE < minChunkSize {
			h.CHUNK_SIZE = minChunkSize
		}
	}
	return h.CHUNK_SIZE
}


func (h *DownloadHandler) calculateOptimalWorkerCount(contentLength int64) int{
	const (
		defaultWorkers = 4  
		maxWorkers     = 16 // preventing overload
	)


	// calculating the number of parts based on chunk size
	h.PartsCount = (contentLength + h.CHUNK_SIZE - 1) / h.CHUNK_SIZE

	// basing the workersCount based on cpu cores and check it later
	cpuCores := runtime.NumCPU()
	h.WORKERS_COUNT = cpuCores

	if int(h.PartsCount) < h.WORKERS_COUNT {
		h.WORKERS_COUNT = int(h.PartsCount)
	}

	// bounding
	if h.WORKERS_COUNT < defaultWorkers {
		h.WORKERS_COUNT = defaultWorkers
	}
	if h.WORKERS_COUNT > maxWorkers {
		h.WORKERS_COUNT = maxWorkers 
	}

	// we need at least one worker
	if h.WORKERS_COUNT < 1 {
		h.WORKERS_COUNT = 1
	}
	return h.WORKERS_COUNT
}