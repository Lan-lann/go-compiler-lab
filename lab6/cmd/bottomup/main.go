package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Lan-lann/go-compiler-lab/lab6/internal/bottomup"
)

func main() {
	inputExpr := flag.String("input", "3+4-5", "指定要计算的简单表达式")
	flag.Parse()

	result, err := bottomup.ParseExpression(*inputExpr)
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
