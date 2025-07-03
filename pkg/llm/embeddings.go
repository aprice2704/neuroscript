// filename: pkg/llm/embeddings.go
package llm

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- Mock Embeddings ---

// GenerateEmbedding creates a mock deterministic embedding.
// Moved from interpreter_c.go
func (i tool.Runtime) GenerateEmbedding(text string) ([]float32, error) {
	// Ensure embeddingDim is valid
	if i.embeddingDim <= 0 {
		return nil, fmt.Errorf("embedding dimension must be positive (is %d)", i.embeddingDim)
	}

	embedding := make([]float32, i.embeddingDim)
	var seed int64
	for _, r := range text {
		// Simple hash combining character codes
		seed = (seed*31 + int64(r)) & 0xFFFFFFFF // Use bitwise AND for potential overflow safety
	}
	// Ensure seed is non-negative if needed by rand.NewSource
	if seed < 0 {
		seed = -seed
	}

	rng := rand.New(rand.NewSource(seed)) // Use the derived seed

	norm := float32(0.0)
	for d := 0; d < i.embeddingDim; d++ {
		// Generate values in [-1, 1]
		val := rng.Float32()*2.0 - 1.0
		embedding[d] = val
		norm += val * val
	}

	// Normalize the vector
	norm = float32(math.Sqrt(float64(norm)))
	// Avoid division by zero or very small numbers
	if norm > 1e-6 {
		for d := range embedding {
			embedding[d] /= norm
		}
	} else if i.embeddingDim > 0 {
		// Handle zero vector case - set first element to 1? Or return error?
		// Setting first element to 1 ensures non-zero magnitude for cosine similarity.
		embedding[0] = 1.0
	}

	return embedding, nil
}

// cosineSimilarity calculates similarity between two vectors.
// Moved from interpreter_c.go
func cosineSimilarity(v1, v2 []float32) (float64, error) {
	if len(v1) == 0 || len(v2) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
	}
	if len(v1) != len(v2) {
		return 0, fmt.Errorf("vector dimensions mismatch (%d vs %d)", len(v1), len(v2))
	}

	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0
	for i := range v1 {
		dotProduct += float64(v1[i] * v2[i])
		norm1 += float64(v1[i] * v1[i])
		norm2 += float64(v2[i] * v2[i])
	}

	mag1 := math.Sqrt(norm1)
	mag2 := math.Sqrt(norm2)

	// Handle cases where one or both vectors have zero magnitude
	if mag1 < 1e-9 || mag2 < 1e-9 {
		// If both are near zero magnitude, they are considered identical (similarity 1)
		if mag1 < 1e-9 && mag2 < 1e-9 {
			return 1.0, nil
		}
		// If only one is near zero, they are orthogonal (similarity 0)
		return 0.0, nil
	}

	// Calculate similarity and clamp to [-1, 1] range due to potential floating point inaccuracies
	similarity := dotProduct / (mag1 * mag2)
	if similarity > 1.0 {
		similarity = 1.0
	} else if similarity < -1.0 {
		similarity = -1.0
	}

	return similarity, nil
}
