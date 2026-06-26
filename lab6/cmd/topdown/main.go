package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Lan-lann/go-compiler-lab/lab6/internal/topdown"
)

func main() {
	inputExpr := flag.String("input", "1+2-3+4", "指定要计算的简单表达式")
	// treeOutput := flag.String("tree", "output/syntax_tree.png", "指定语法树输出路径")
	flag.Parse()

	result, err := topdown.ParseExpression(*inputExpr)
	// result, tree, err := topdown.ParseExpression(*inputExpr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "解析失败：", err)
		os.Exit(1)
	}

	fmt.Printf("表达式结果: %d\n", result)

	// if err :=  topdown.DrawAST(tree, *treeOutput); err != nil {
	// 	fmt.Fprintln(os.Stderr, "绘制语法树失败：", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("语法树已生成: %s\n", *treeOutput)
}
