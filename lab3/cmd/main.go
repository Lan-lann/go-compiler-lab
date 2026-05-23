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

		// 判断是否含有左递归
		if g.HaveLeftRecursion() {
			fmt.Println("含有左递归, 尝试改写")
			// 消除间接左递归
			g.FirstLetterSubstitution()
			// 消除直接左递归
			g.ParsingDirectLeftRecursion()
			// 消除不可到达式
			g.DelUnreachableProduction()
			// 判断是否为 LL1 文法
			if g.IsLL1() {
				fmt.Println("改写后是 LL1 文法")
				g.ShowSelectSet()
			} else {
				fmt.Println("改写后不是 LL1 文法")
			}
			return
		}

		// 判断是否含有左公共因子
		if g.HaveCommonFactor() {
			fmt.Println("含有左公共因子, 尝试改写")
			// 提取隐藏的左公共因子
			g.FirstLetterSubstitutionForCommonFactor()
			// 消除左公共因子
			g.ParsingCommonFactor()
			// 消除不可到达式
			g.DelUnreachableProduction()
			fmt.Println(g)

			if g.IsLL1() {
				fmt.Println("改写后是 LL1 文法")
				g.ShowSelectSet()
				return
			} else {
				fmt.Println("改写后不是 LL1 文法")
			}
			return
		}

		fmt.Println("无法进行改写")
	}
}
