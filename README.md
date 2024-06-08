# transitive-strength-backend
backend for the transitive strength app
We use Topological sort to solve this problem.
This is the reverse of DFS.
One way of doing it is by building the graph then taking the first 0-indegree element from queue, then removing it, thereby reducing all it's edges and neighbour's in-degree by 1. Then recursing through the process until it becomes 0 and that cascades into 0 and so on.

Everytime you take a node out that's in-degree 0 you put into result. Then final output should have same nodes as the no. of vertices.


```bash
curl -X POST http://localhost:8080/api/submit_pairs -H "Content-Type: application/json" --data "@input_pairs.json"

```
