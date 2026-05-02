package dfa

import (
	mapset "github.com/deckarep/golang-set/v2"
)

func (dfa *DFA) NewStatesSet() []mapset.Set[string] {
	// 初始化状态集合
	nonFinalStateSet := mapset.NewSet[string]()
	finalStateSet := mapset.NewSet(dfa.FinalStates.ToSlice()...)
	// 初始化非终态集合
	for _, state := range dfa.States.ToSlice() {
		if !finalStateSet.Contains(state) {
			nonFinalStateSet.Add(state)
		}
	}

	StatesSetList := []mapset.Set[string]{nonFinalStateSet, finalStateSet}
	// 循环遍历
	for {
		// 通过状态集合长度是否变化判断划分是否结束
		length := len(StatesSetList)

		// 开始划分
		for i, stateSet := range StatesSetList {
			// 是否进行了划分
			flag := false
			// 若集合只有一个元素，则无需划分
			if stateSet.Cardinality() == 1 {
				continue
			}
			// 遍历输入字符
			for _, ch := range dfa.Alphabet.ToSlice() {

				// 依据集合长度生成空转移集合
				NoneStatesSetList := make([]mapset.Set[string], length)
				for i := range NoneStatesSetList {
					NoneStatesSetList[i] = mapset.NewSet[string]()
				}

				// 遍历集合里每一个元素
				for _, state := range stateSet.ToSlice() {
					nextState := dfa.Transitions[Pair{state, ch}]
					// 依据后继填充空转移集合
					for i, statesSet := range StatesSetList {
						if statesSet.Contains(nextState) {
							NoneStatesSetList[i].Add(state)
						}
					}
				}

				newStatesSetList := []mapset.Set[string]{}
				// 去除空集合，只留下可以到达的集合
				for _, stateset := range NoneStatesSetList {
					if stateset.Cardinality() > 0 {
						newStatesSetList = append(newStatesSetList, stateset)
					}
				}
				// 若集合内元素存在多个转移集合，则需要划分
				if len(newStatesSetList) >= 2 {
					// 将原集合分裂，直接从原集合位置插入
					StatesSetList = append(StatesSetList[:i], append(newStatesSetList, StatesSetList[i+1:]...)...)
					// 进行了划分
					flag = true
					break
				}

			}
			// 若进行了划分，则本次遍历结束，重新开始遍历
			if flag {
				break
			}
		}
		// 若状态集合无法继续划分，则停止
		if len(StatesSetList) == length {
			break
		}

	}

	return StatesSetList
}

func (dfa *DFA) Minimize() *DFA {
	// 初始化新 DFA
	newDFA := &DFA{
		States:      mapset.NewSet[string](),
		Alphabet:    mapset.NewSet[string](),
		Transitions: map[Pair]string{},
		StartStates: "",
		FinalStates: mapset.NewSet[string](),
		StateMap:    map[string]string{},
	}

	newDFA.Alphabet = dfa.Alphabet
	// 得到新状态集合
	statesSetList := dfa.NewStatesSet()

	stateList := []mapset.Set[string]{}

	// 向新 DFA 添加状态
	//	找初态
	for _, set := range statesSetList {
		if set.Contains(dfa.StartStates) {
			setByName := getStateName(setToKey(set), &newDFA.StateMap)
			newDFA.StartStates = setByName
		}
	}

	//	找中间状态
	for _, set := range statesSetList {
		setByName := getStateName(setToKey(set), &newDFA.StateMap)
		newDFA.States.Add(setByName)
		stateList = append(stateList, set)
	}

	// 找终态
	for _, set := range statesSetList {
		for _, state := range dfa.FinalStates.ToSlice() {
			if set.Contains(state) {
				setByName := getStateName(setToKey(set), &newDFA.StateMap)
				newDFA.FinalStates.Add(setByName)
			}
		}
	}

	// 依据原 DFA 确定转移函数
	for input, toState := range dfa.Transitions {
		// 获取当前状态和输入字符
		fromState, ch := input.First, input.Second
		var newFromState mapset.Set[string]
		var newToState mapset.Set[string]

		// 判断原 DFA 状态属于新 DFA 的哪个状态
		for _, state := range stateList {

			if state.Contains(fromState) {
				newFromState = state
			}

			if state.Contains(toState) {
				newToState = state
			}
		}

		// 获取状态短名称
		newFromStateByName := getStateName(setToKey(newFromState), &newDFA.StateMap)
		newToStateByName := getStateName(setToKey(newToState), &newDFA.StateMap)

		// 建立转移函数
		newDFA.Transitions[Pair{newFromStateByName, ch}] = newToStateByName
	}

	return newDFA
}
