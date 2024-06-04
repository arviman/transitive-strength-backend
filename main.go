package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Pair struct {
	From string `json:"from" validate:"required"`
	To   string `json:"to" validate:"required"`
}

type Pairs struct {
	Pairs []Pair `json:"pairs" validate:"required,dive"`
}

var validate *validator.Validate

// curl -X POST http://localhost:8080/api/submit_pairs -H "Content-Type: application/json" -d '{"pairs":[{"from":"BTC","to":"ETH"},{"from":"ETH","to":"XRP"}]}'
func main() {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Next()
	})

	validate = validator.New()

	r.POST("/api/submit_pairs", submitPairs)

	r.Run(":8080")
}

func submitPairs(c *gin.Context) {
	var pairs Pairs

	if err := c.ShouldBindJSON(&pairs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(pairs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	graph, inDegree := buildGraph(pairs.Pairs)
	sorted, err := topologicalSort(graph, inDegree)

	if err != nil {
		mostOutgoing, leastIncoming := findCycleBreakers(graph, inDegree)
		c.JSON(http.StatusOK, gin.H{
			"message": "Cycle detected. Break cycle by removing edge",
			"from":    mostOutgoing,
			"to":      leastIncoming,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sorted": sorted})
}

func buildGraph(pairs []Pair) (map[string][]string, map[string]int) {
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, pair := range pairs {
		graph[pair.From] = append(graph[pair.From], pair.To)
		inDegree[pair.To]++
		if _, exists := inDegree[pair.From]; !exists {
			inDegree[pair.From] = 0
		}
	}

	return graph, inDegree
}

func topologicalSort(graph map[string][]string, inDegree map[string]int) ([]string, error) {
	var queue []string
	var sorted []string
	inDegreeCopy := make(map[string]int)
	for k, v := range inDegree {
		inDegreeCopy[k] = v
		if v == 0 {
			queue = append(queue, k)
		}
	}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		sorted = append(sorted, node)

		for _, neighbor := range graph[node] {
			inDegreeCopy[neighbor]--
			if inDegreeCopy[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(sorted) == len(graph) {
		return sorted, nil
	}

	return nil, fmt.Errorf("cycle detected")
}

func findCycleBreakers(graph map[string][]string, inDegree map[string]int) (string, string) {
	maxOutgoingNode := ""
	maxOutgoingCount := 0
	minIncomingNode := ""
	minIncomingCount := int(^uint(0) >> 1) // Max int value

	for node, neighbors := range graph {
		if len(neighbors) > maxOutgoingCount {
			maxOutgoingNode = node
			maxOutgoingCount = len(neighbors)
		}
	}

	for node, degree := range inDegree {
		if degree < minIncomingCount {
			minIncomingNode = node
			minIncomingCount = degree
		}
	}

	return maxOutgoingNode, minIncomingNode
}
