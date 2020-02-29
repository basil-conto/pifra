package pifra

import (
	"strconv"

	"github.com/mohae/deepcopy"
)

var boundNameIndex int
var freshNameIndex int

func generateFreshName(namePrefix string) string {
	name := namePrefix + "_" + strconv.Itoa(freshNameIndex)
	freshNameIndex = freshNameIndex + 1
	return name
}

func generateBoundName(namePrefix string) string {
	name := namePrefix + "_" + strconv.Itoa(boundNameIndex)
	boundNameIndex = boundNameIndex + 1
	return name
}

func substituteName(elem Element, oldName Name, newName Name) {
	subName(elem, oldName, newName)
}

func subName(elem Element, oldName Name, newName Name) {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		if outElem.Channel == oldName {
			outElem.Channel = newName
		}
		if outElem.Output == oldName {
			outElem.Output = newName
		}
		subName(outElem.Next, oldName, newName)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		if inpElem.Channel == oldName {
			inpElem.Channel = newName
		}
		if inpElem.Input == oldName {
			inpElem.Input = newName
		}
		subName(inpElem.Next, oldName, newName)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		if matchElem.NameL == oldName {
			matchElem.NameL = newName
		}
		if matchElem.NameR == oldName {
			matchElem.NameR = newName
		}
		subName(matchElem.Next, oldName, newName)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		subName(resElem.Next, oldName, newName)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		subName(sumElem.ProcessL, oldName, newName)
		subName(sumElem.ProcessR, oldName, newName)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		subName(parElem.ProcessL, oldName, newName)
		subName(parElem.ProcessR, oldName, newName)
	case ElemTypProcess:
		procElem := elem.(*ElemProcess)
		for i, param := range procElem.Parameters {
			if param == oldName {
				procElem.Parameters[i] = newName
			}
		}
	case ElemTypOutOutput:
		outOutput := elem.(*ElemOutOutput)
		if outOutput.Output == oldName {
			outOutput.Output = newName
		}
		subName(outOutput.Next, oldName, newName)
	case ElemTypInpInput:
		inpInput := elem.(*ElemInpInput)
		if inpInput.Input == oldName {
			inpInput.Input = newName
		}
		subName(inpInput.Next, oldName, newName)
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		subName(rootElem.Next, oldName, newName)
	}
}

func open(elem Element, boundName Name) Element {
	if boundName.Type != Bound {
		return nil
	}
	elemCopy := deepcopy.Copy(elem)
	elem = elemCopy.(Element)
	freshName := Name{
		Name: generateFreshName("fn"),
		Type: Fresh,
	}
	subName(elem, boundName, freshName)
	return elem
}

func close(elem Element, freshName Name) Element {
	if freshName.Type != Fresh {
		return nil
	}
	elemCopy := deepcopy.Copy(elem)
	elem = elemCopy.(Element)
	boundName := Name{
		Name: generateBoundName(freshName.Name),
		Type: Bound,
	}
	subName(elem, freshName, boundName)
	return elem
}

// InitRootAst performs alpha-conversion and adds a root element to the AST as the head,
// for use in the interpreter.
func InitRootAst(elem Element) Element {
	DoAlphaConversion(elem)
	return &ElemRoot{
		Next: elem,
	}
}

// DoAlphaConversion renames bound names to names appropriate to their scope.
func DoAlphaConversion(elem Element) {
	doAlphaConversion(elem)
}

func doAlphaConversion(elem Element) {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
	case ElemTypOutput:
		doAlphaConversion(elem.(*ElemOutput).Next)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		boundName := inpElem.Input.Name
		newName := generateBoundName(boundName)
		inpElem.Input = Name{
			Name: newName,
			Type: Bound,
		}
		subBoundNames(inpElem.Next, boundName, newName)
		doAlphaConversion(inpElem.Next)
	case ElemTypMatch:
		doAlphaConversion(elem.(*ElemMatch).Next)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		boundName := resElem.Restrict.Name
		newName := generateBoundName(boundName)
		resElem.Restrict = Name{
			Name: newName,
			Type: Bound,
		}
		subBoundNames(resElem.Next, boundName, newName)
		doAlphaConversion(resElem.Next)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		doAlphaConversion(sumElem.ProcessL)
		doAlphaConversion(sumElem.ProcessR)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		doAlphaConversion(parElem.ProcessL)
		doAlphaConversion(parElem.ProcessR)
	case ElemTypProcess:
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		doAlphaConversion(rootElem.Next)
	}
}

func subBoundNames(elem Element, boundName string, newName string) {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		if outElem.Channel.Name == boundName {
			outElem.Channel = Name{
				Name: newName,
				Type: Bound,
			}
		}
		if outElem.Output.Name == boundName {
			outElem.Output = Name{
				Name: newName,
				Type: Bound,
			}
		}
		subBoundNames(outElem.Next, boundName, newName)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		if inpElem.Channel.Name == boundName {
			inpElem.Channel = Name{
				Name: newName,
				Type: Bound,
			}
		}
		if inpElem.Input.Name != boundName {
			subBoundNames(inpElem.Next, boundName, newName)
		}
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		if matchElem.NameL.Name == boundName {
			matchElem.NameL = Name{
				Name: newName,
				Type: Bound,
			}
		}
		if matchElem.NameR.Name == boundName {
			matchElem.NameR = Name{
				Name: newName,
				Type: Bound,
			}
		}
		subBoundNames(matchElem.Next, boundName, newName)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		if resElem.Restrict.Name != boundName {
			subBoundNames(resElem.Next, boundName, newName)
		}
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		subBoundNames(sumElem.ProcessL, boundName, newName)
		subBoundNames(sumElem.ProcessR, boundName, newName)
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		subBoundNames(parElem.ProcessL, boundName, newName)
		subBoundNames(parElem.ProcessR, boundName, newName)
	case ElemTypProcess:
		pcsElem := elem.(*ElemProcess)
		for i, param := range pcsElem.Parameters {
			if param.Name == boundName {
				pcsElem.Parameters[i] = Name{
					Name: newName,
					Type: Bound,
				}
			}
		}
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		subBoundNames(rootElem.Next, boundName, newName)
	}
}

// PrettyPrintAst returns a string containing the pi-calculus syntax of the AST.
func PrettyPrintAst(elem Element) string {
	return prettyPrintAcc(elem, "")
}

func prettyPrintAcc(elem Element, str string) string {
	elemTyp := elem.Type()
	switch elemTyp {
	case ElemTypNil:
		str = str + "0"
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		str = str + outElem.Channel.Name + "'<" + outElem.Output.Name + ">."
		return prettyPrintAcc(outElem.Next, str)
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		str = str + inpElem.Channel.Name + "(" + inpElem.Input.Name + ")."
		return prettyPrintAcc(inpElem.Next, str)
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		str = str + "[" + matchElem.NameL.Name + "=" + matchElem.NameL.Name + "]"
		return prettyPrintAcc(matchElem.Next, str)
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		str = str + "$" + resElem.Restrict.Name + "."
		return prettyPrintAcc(resElem.Next, str)
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		left := prettyPrintAcc(sumElem.ProcessL, "")
		right := prettyPrintAcc(sumElem.ProcessR, "")
		str = str + "(" + left + " + " + right + ")"
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		left := prettyPrintAcc(parElem.ProcessL, "")
		right := prettyPrintAcc(parElem.ProcessR, "")
		str = str + "(" + left + " | " + right + ")"
	case ElemTypProcess:
		pcsElem := elem.(*ElemProcess)
		if len(pcsElem.Parameters) == 0 {
			str = str + pcsElem.Name
		} else {
			params := "("
			for i, param := range pcsElem.Parameters {
				if i == len(pcsElem.Parameters)-1 {
					params = params + param.Name + ")"
				} else {
					params = params + param.Name + ", "
				}
			}
			str = str + pcsElem.Name + params
		}
	case ElemTypOutOutput:
		outOutput := elem.(*ElemOutOutput)
		str = str + "'<" + outOutput.Output.Name + ">."
		return prettyPrintAcc(outOutput.Next, str)
	case ElemTypInpInput:
		inpInput := elem.(*ElemInpInput)
		str = str + "(" + inpInput.Input.Name + ")."
		return prettyPrintAcc(inpInput.Next, str)
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		return prettyPrintAcc(rootElem.Next, str)
	}
	return str
}

// GetAllFreshNames returns all fresh names in the AST.
func GetAllFreshNames(elem Element) []string {
	visitedProcs := make(map[string]bool)

	var getAllFreshNamesAcc func(Element, []string) []string
	getAllFreshNamesAcc = func(elem Element, freshNames []string) []string {
		switch elem.Type() {
		case ElemTypNil:
		case ElemTypOutput:
			outElem := elem.(*ElemOutput)
			if outElem.Channel.Type == Fresh {
				freshNames = append(freshNames, outElem.Channel.Name)
			}
			if outElem.Output.Type == Fresh {
				freshNames = append(freshNames, outElem.Output.Name)
			}
			return getAllFreshNamesAcc(outElem.Next, freshNames)
		case ElemTypInput:
			inpElem := elem.(*ElemInput)
			if inpElem.Channel.Type == Fresh {
				freshNames = append(freshNames, inpElem.Channel.Name)
			}
			if inpElem.Input.Type == Fresh {
				freshNames = append(freshNames, inpElem.Input.Name)
			}
			return getAllFreshNamesAcc(inpElem.Next, freshNames)
		case ElemTypMatch:
			matchElem := elem.(*ElemMatch)
			if matchElem.NameL.Type == Fresh {
				freshNames = append(freshNames, matchElem.NameL.Name)
			}
			if matchElem.NameR.Type == Fresh {
				freshNames = append(freshNames, matchElem.NameR.Name)
			}
			return getAllFreshNamesAcc(matchElem.Next, freshNames)
		case ElemTypRestriction:
			resElem := elem.(*ElemRestriction)
			if resElem.Restrict.Type == Fresh {
				freshNames = append(freshNames, resElem.Restrict.Name)
			}
			return getAllFreshNamesAcc(resElem.Next, freshNames)
		case ElemTypSum:
			sumElem := elem.(*ElemSum)
			freshNames = getAllFreshNamesAcc(sumElem.ProcessL, freshNames)
			freshNames = getAllFreshNamesAcc(sumElem.ProcessR, freshNames)
		case ElemTypParallel:
			parElem := elem.(*ElemParallel)
			freshNames = getAllFreshNamesAcc(parElem.ProcessL, freshNames)
			freshNames = getAllFreshNamesAcc(parElem.ProcessR, freshNames)
		case ElemTypProcess:
			procElem := elem.(*ElemProcess)

			// Parameter checks.
			processName := procElem.Name
			if _, ok := DeclaredProcs[processName]; !ok {
				return freshNames
			}
			dp := DeclaredProcs[processName]
			if len(dp.Parameters) != len(procElem.Parameters) {
				return freshNames
			}

			// Do alpha conversion on declared process.
			// Restore original boundNameIndex because process is only used
			// for finding free names. Bound names are disregarded.
			proc := deepcopy.Copy(dp.Process).(Element)
			bni := boundNameIndex
			doAlphaConversion(proc)
			boundNameIndex = bni

			// Substitute parameter names to the new process.
			for i, oldName := range dp.Parameters {
				subName(proc, Name{
					Name: oldName,
				}, procElem.Parameters[i])
			}

			// Prevent infinitely looping processes.
			if !visitedProcs[processName] {
				visitedProcs[processName] = true
				// Find free names in declared process.
				freshNames = getAllFreshNamesAcc(proc, freshNames)
			}
		case ElemTypOutOutput:
			outOutput := elem.(*ElemOutOutput)
			if outOutput.Output.Type == Fresh {
				freshNames = append(freshNames, outOutput.Output.Name)
			}
			return getAllFreshNamesAcc(outOutput.Next, freshNames)
		case ElemTypInpInput:
			inpInput := elem.(*ElemInpInput)
			if inpInput.Input.Type == Fresh {
				freshNames = append(freshNames, inpInput.Input.Name)
			}
			return getAllFreshNamesAcc(inpInput.Next, freshNames)
		case ElemTypRoot:
			rootElem := elem.(*ElemRoot)
			return getAllFreshNamesAcc(rootElem.Next, freshNames)
		}
		return freshNames
	}

	return getAllFreshNamesAcc(elem, []string{})
}

func getElemSetType(elem Element) ElemSetType {
	switch elem.Type() {
	case ElemTypNil:
		nilElem := elem.(*ElemNil)
		return nilElem.SetType
	case ElemTypOutput:
		outElem := elem.(*ElemOutput)
		return outElem.SetType
	case ElemTypInput:
		inpElem := elem.(*ElemInput)
		return inpElem.SetType
	case ElemTypMatch:
		matchElem := elem.(*ElemMatch)
		return matchElem.SetType
	case ElemTypRestriction:
		resElem := elem.(*ElemRestriction)
		return resElem.SetType
	case ElemTypSum:
		sumElem := elem.(*ElemSum)
		return sumElem.SetType
	case ElemTypParallel:
		parElem := elem.(*ElemParallel)
		return parElem.SetType
	case ElemTypProcess:
		pcsElem := elem.(*ElemProcess)
		return pcsElem.SetType
	case ElemTypOutOutput:
		elemOutOut := elem.(*ElemOutOutput)
		return elemOutOut.SetType
	case ElemTypInpInput:
		elemInpInp := elem.(*ElemInpInput)
		return elemInpInp.SetType
	case ElemTypRoot:
		rootElem := elem.(*ElemRoot)
		return rootElem.SetType
	}
	return ElemSetReg
}
