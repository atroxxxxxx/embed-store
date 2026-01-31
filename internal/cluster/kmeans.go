package cluster

import (
	"errors"
	"math/rand"
	"time"
)

type ClusterConfig struct {
	Clusters int
	Iters    int
	Workers  int
}

var (
	ErrEmptyDataset       = errors.New("empty dataset")
	ErrInvalidClusterSize = errors.New("clusters must be > 0")
	ErrInvalidVectorDims  = errors.New("invalid vector dims")
)

func kMeans(vectors [][]float32, cfg ClusterConfig) ([]int32, error) {
	if len(vectors) == 0 {
		return nil, ErrEmptyDataset
	}
	if cfg.Clusters <= 0 {
		return nil, ErrInvalidClusterSize
	}
	if cfg.Iters <= 0 {
		cfg.Iters = 10
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}

	dim := len(vectors[0])
	for i := 1; i < len(vectors); i++ {
		if len(vectors[i]) != dim {
			return nil, ErrInvalidVectorDims
		}
	}

	clusterCount := cfg.Clusters
	if clusterCount > len(vectors) {
		clusterCount = len(vectors)
	}

	centroids := initCentroidsRandom(vectors, clusterCount)

	for iter := 0; iter < cfg.Iters; iter++ {
		assignments := assignClusters(vectors, centroids, cfg.Workers)
		centroids = recomputeCentroids(vectors, assignments, clusterCount, dim, cfg.Workers, centroids)

		// если хочешь, можно вернуть assignments после последней итерации
		if iter == cfg.Iters-1 {
			return assignments, nil
		}
	}

	return nil, nil
}

func initCentroidsRandom(vectors [][]float32, count int) [][]float32 {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	centroids := make([][]float32, 0, count)

	used := make(map[int]struct{}, count)
	for len(centroids) < count {
		idx := rnd.Intn(len(vectors))
		if _, ok := used[idx]; ok {
			continue
		}
		used[idx] = struct{}{}

		centroid := make([]float32, len(vectors[idx]))
		copy(centroid, vectors[idx])
		centroids = append(centroids, centroid)
	}

	return centroids
}

type assignJob struct {
	start int
	end   int
}

func assignClusters(vectors [][]float32, centroids [][]float32, workers int) []int32 {
	n := len(vectors)
	assignments := make([]int32, n)

	jobs := make(chan assignJob)
	done := make(chan struct{}, workers)

	for w := 0; w < workers; w++ {
		go func() {
			for job := range jobs {
				for i := job.start; i < job.end; i++ {
					assignments[i] = int32(findNearestCentroid(vectors[i], centroids))
				}
			}
			done <- struct{}{}
		}()
	}

	chunkSize := (n + workers - 1) / workers
	for start := 0; start < n; start += chunkSize {
		end := start + chunkSize
		if end > n {
			end = n
		}
		jobs <- assignJob{start: start, end: end}
	}
	close(jobs)

	for w := 0; w < workers; w++ {
		<-done
	}

	return assignments
}

func findNearestCentroid(vec []float32, centroids [][]float32) int {
	bestIndex := 0
	bestDist := squareDistance(vec, centroids[0])

	for c := 1; c < len(centroids); c++ {
		d := squareDistance(vec, centroids[c])
		if d < bestDist {
			bestDist = d
			bestIndex = c
		}
	}
	return bestIndex
}

type recomputeJob struct {
	start int
	end   int
}

func recomputeCentroids(
	vectors [][]float32,
	assignments []int32,
	clusterCount int,
	dim int,
	workers int,
	prevCentroids [][]float32,
) [][]float32 {
	// localSums[worker][cluster][dim]
	localSums := make([][][]float32, workers)
	localCounts := make([][]int, workers)

	for w := 0; w < workers; w++ {
		localSums[w] = make([][]float32, clusterCount)
		for c := 0; c < clusterCount; c++ {
			localSums[w][c] = make([]float32, dim)
		}
		localCounts[w] = make([]int, clusterCount)
	}

	jobs := make(chan recomputeJob)
	done := make(chan struct{}, workers)

	for w := 0; w < workers; w++ {
		workerID := w
		go func() {
			for job := range jobs {
				for i := job.start; i < job.end; i++ {
					clusterID := int(assignments[i])
					localCounts[workerID][clusterID]++

					sum := localSums[workerID][clusterID]
					vec := vectors[i]
					for d := 0; d < dim; d++ {
						sum[d] += vec[d]
					}
				}
			}
			done <- struct{}{}
		}()
	}

	n := len(vectors)
	chunkSize := (n + workers - 1) / workers
	for start := 0; start < n; start += chunkSize {
		end := start + chunkSize
		if end > n {
			end = n
		}
		jobs <- recomputeJob{start: start, end: end}
	}
	close(jobs)

	for w := 0; w < workers; w++ {
		<-done
	}

	totalCounts := make([]int, clusterCount)
	totalSums := make([][]float32, clusterCount)
	for c := 0; c < clusterCount; c++ {
		totalSums[c] = make([]float32, dim)
	}

	for w := 0; w < workers; w++ {
		for c := 0; c < clusterCount; c++ {
			totalCounts[c] += localCounts[w][c]
			for d := 0; d < dim; d++ {
				totalSums[c][d] += localSums[w][c][d]
			}
		}
	}

	centroids := make([][]float32, clusterCount)
	for c := 0; c < clusterCount; c++ {
		centroids[c] = make([]float32, dim)

		if totalCounts[c] == 0 {
			copy(centroids[c], prevCentroids[c])
			continue
		}

		inv := 1 / float32(totalCounts[c])
		for d := 0; d < dim; d++ {
			centroids[c][d] = totalSums[c][d] * inv
		}
	}

	return centroids
}
