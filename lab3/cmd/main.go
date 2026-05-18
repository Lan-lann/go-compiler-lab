package main

import (
	"flag"
	"fmt"

	"github.com/Lan-lann/go-compiler-lab/lab3/internal/grammar"
)

func main() {
	inputFile := flag.String("input", "input/g1.txt", "指定文法文件路径")
	flag.Parse()

	g, err := grammar.LoadGrammar(*inputFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(g)
	// 首先判断是否为 LL1 文法
	if g.IsLL1() {
		fmt.Println("是 LL1 文法")
		g.ShowSelectSet()
	} else { // 若不是 LL1 文法，尝试进行改写
		fmt.Println("不是 LL1 文法, 尝试改写")

		if g.HaveLeftRecursion() {
			fmt.Println("含有左递归, 尝试改写")

			g.FirstLetterSubstitution()
			g.ParsingDirectLeftRecursion()
			g.DelUnreachableProduction()

			fmt.Println(g)
			if g.IsLL1() {
				fmt.Println("改写后是 LL1 文法")
				g.ShowSelectSet()
			} else {
				fmt.Println("改写后不是 LL1 文法")
			}
		}

		if g.HaveCommonFactor() {
			fmt.Println("含有左公共因子, 尝试改写")

			g.FirstLetterSubstitutionForCommonFactor()
			g.ParsingCommonFactor()
			g.DelUnreachableProduction()
			fmt.Println(g)

			if g.IsLL1() {
				fmt.Println("改写后是 LL1 文法")
				g.ShowSelectSet()
				return
			} else {
				fmt.Println("改写后不是 LL1 文法")
			}
		}
	}
}
