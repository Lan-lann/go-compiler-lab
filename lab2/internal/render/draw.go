package render

import (
	"context"

	"github.com/Lan-lann/go-compiler-lab/lab2/internal/dfa"
	"github.com/Lan-lann/go-compiler-lab/lab2/internal/nfa"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// 绘制 DFA 图像
func DrawGraphDFA(dfa *dfa.DFA, path string) {
	ctx := context.Background()
	g, err := graphviz.New(ctx)
	if err != nil {
		panic(err)
	}

	graph, err := g.Graph()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			panic(err)
		}
		g.Close()
	}()

	graph.SetRankDir(cgraph.LRRank)
	nodeMap := make(map[string]*cgraph.Node)

	for _, node := range dfa.States.ToSlice() {
		n, err := graph.CreateNodeByName(node)
		if err != nil {
			panic(err)
		}
		n.SetShape(cgraph.CircleShape)
		if dfa.FinalStates.Contains(node) {
			n.SetShape(cgraph.DoubleCircleShape)
		}

		nodeMap[node] = n
	}

	for formState, ToState := range dfa.Transitions {
		formNode := nodeMap[formState.First]
		ch := formState.Second
		ToNode := nodeMap[ToState]

		e, err := graph.CreateEdgeByName(ch, formNode, ToNode)
		if err != nil {
			panic(err)
		}
		e.SetLabel(ch)
	}

	if err := g.RenderFilename(ctx, graph, graphviz.PNG, path); err != nil {
		panic(err)
	}
}

// 绘制 NFA 图像
func DrawGraphNFA(nfa *nfa.NFA, path string) {
	ctx := context.Background()
	g, err := graphviz.New(ctx)
	if err != nil {
		panic(err)
	}

	graph, err := g.Graph()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			panic(err)
		}
		g.Close()
	}()

	graph.SetRankDir(cgraph.LRRank)

	nodeMap := make(map[string]*cgraph.Node)

	for _, node := range nfa.States.ToSlice() {
		n, err := graph.CreateNodeByName(node)
		if err != nil {
			panic(err)
		}
		n.SetShape(cgraph.CircleShape)
		if nfa.FinalStates.Contains(node) {
			n.SetShape(cgraph.DoubleCircleShape)
		}

		nodeMap[node] = n
	}

	for formState, ToState := range nfa.Transitions {
		formNode := nodeMap[formState.First]
		ch := formState.Second
		for _, toState := range ToState.ToSlice() {
			ToNode := nodeMap[toState]

			e, err := graph.CreateEdgeByName(ch, formNode, ToNode)
			if err != nil {
				panic(err)
			}
			e.SetLabel(ch)
		}

	}

	if err := g.RenderFilename(ctx, graph, graphviz.PNG, path); err != nil {
		panic(err)
	}
}
