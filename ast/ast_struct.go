package ast

import (
	"fmt"
	"strings"
)

type StatementType int
type OperationType int

const (
	PFTypeUndefined StatementType = iota
	PFTypeProcedure
	PFTypeFunction
)

const (
	OpPlus OperationType = iota
	OpMinus
	OpMul
	OpDiv
	OpEq  // =
	OpGt  // >
	OpLt  // <
	OpNe  // <>
	OpLe  // <=
	OpGe  // >=
	OpMod // % - деление по модулю
	OpOr
	OpAnd
)

type IUnary interface {
	UnaryMinus() interface{}
}

type INot interface {
	Not() interface{}
}

type IParams interface {
	Params() []Statement
}

type Statement interface{}

type GlobalVariables struct {
	Directive string
	Var       VarStatement
	Export    bool
}

type ModuleStatement struct {
	Name            string
	GlobalVariables map[string]GlobalVariables `json:"GlobalVariables,omitempty"`
	Body            []Statement
}

type VarStatement struct {
	Name string
	addStatementField
}

type FunctionOrProcedure struct {
	ExplicitVariables map[string]VarStatement
	Name              string
	Directive         string
	Body              []Statement
	Params            []ParamStatement
	Type              StatementType
	Export            bool
}

type ParamStatement struct {
	Default Statement `json:"Default,omitempty"`
	Name    string
	IsValue bool `json:"IsValue,omitempty"`
}

type addStatementField struct {
	unaryMinus bool
	unaryPlus  bool
	not        bool
}

type ExpStatement struct {
	Left      interface{}
	Right     interface{}
	Operation OperationType
	addStatementField
}

// type IfElseStatement struct {
// 	Expression Statement
// 	TrueBlock  []Statement
// }

type IfStatement struct {
	Expression  Statement
	TrueBlock   []Statement
	IfElseBlock []Statement
	ElseBlock   []Statement
}

type TryStatement struct {
	Body  []Statement
	Catch []Statement
}

type ThrowStatement struct {
	Param Statement
}

type UndefinedStatement struct{}

type ReturnStatement struct {
	Param Statement
}

type NewObjectStatement struct {
	Constructor string
	Param       []Statement
}

type CallChainStatement struct {
	Unit Statement
	Call Statement
	addStatementField
}

type MethodStatement struct {
	Name  string
	Param []Statement
	addStatementField
}

type BreakStatement struct {
}

type ContinueStatement struct {
}

type LoopStatement struct {
	For       Statement `json:"For,omitempty"`
	To        Statement `json:"To,omitempty"`
	In        Statement `json:"In,omitempty"`
	WhileExpr Statement `json:"WhileExpr,omitempty"`
	Body      []Statement
}

type TernaryStatement struct {
	Expression Statement
	TrueBlock  Statement
	ElseBlock  Statement
}

type ItemStatement struct {
	Item   Statement
	Object Statement
}

type GoToStatement struct {
	Label *GoToLabelStatement
}

type GoToLabelStatement struct {
	Name string
}

type BuiltinFunctionStatement struct {
	Name  string
	Param []Statement
}

var ierahiy = 0

func (p *ParamStatement) Fill(valueParam *Token, identifier Token) *ParamStatement {
	p.IsValue = valueParam != nil
	p.Name = identifier.literal
	return p
}

func (p *ParamStatement) DefaultValue(value Statement) *ParamStatement {
	if value == nil {
		p.Default = UndefinedStatement{}
	} else {
		p.Default = value
	}

	return p
}

func (e *ExpStatement) UnaryMinus() interface{} {
	e.unaryMinus = true
	return e
}

func (e *ExpStatement) Not() interface{} {
	e.not = true
	return e
}

func (e VarStatement) UnaryMinus() interface{} {
	e.unaryMinus = true
	return e
}

func (e VarStatement) Not() interface{} {
	e.not = true
	return e
}

func (e CallChainStatement) UnaryMinus() interface{} {
	e.unaryMinus = true
	return e
}

func (e CallChainStatement) Not() interface{} {
	e.not = true
	return e
}

// IsMethod вернет true в случаях Блокировка.Заблокировать() и false для Источник.Ссылка
func (e CallChainStatement) IsMethod() bool {
	_, ok := e.Unit.(MethodStatement)
	return ok
}

func (e MethodStatement) Not() interface{} {
	e.not = true
	return e
}

func (n NewObjectStatement) Params() []Statement {
	return n.Param
}

func (n MethodStatement) Params() []Statement {
	return n.Param
}

func (o OperationType) String() string {
	switch o {
	case OpPlus:
		return "+"
	case OpMinus:
		return "-"
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	case OpEq:
		return "="
	case OpGt:
		return ">"
	case OpLt:
		return "<"
	case OpNe:
		return "<>"
	case OpLe:
		return "<="
	case OpGe:
		return ">="
	case OpMod:
		return "%"
	case OpOr:
		return "ИЛИ"
	case OpAnd:
		return "И"
	default:
		return ""
	}
}

func (m ModuleStatement) Walk(callBack func(current *FunctionOrProcedure, statement *Statement)) {
	fmt.Println("Walk ModuleStatement", m.Name)
	StatementWalk(m.Body, callBack)
}

func StatementWalk(stm []Statement, callBack func(current *FunctionOrProcedure, statement *Statement)) {
	fmt.Println("StatementWalk:", len(stm))
	walkHelper(nil, stm, callBack)
}

func (m *ModuleStatement) Append(item Statement, yylex yyLexer) {
	switch v := item.(type) {
	case GlobalVariables:
		if len(m.Body) > 0 {
			yylex.Error("variable declarations must be placed at the beginning of the module")
			return
		}

		if m.GlobalVariables == nil {
			m.GlobalVariables = map[string]GlobalVariables{}
		}

		if _, ok := m.GlobalVariables[v.Var.Name]; ok {
			yylex.Error(fmt.Sprintf("%v: with the specified name %q", errVariableAlreadyDefined, v.Var.Name))
		} else {
			m.GlobalVariables[v.Var.Name] = v
		}
	case []GlobalVariables:
		for _, item := range v {
			m.Append(item, yylex)
		}
	case []Statement:
		m.Body = append(m.Body, v...)
	case *FunctionOrProcedure:
		// если предыдущее выражение не процедура функция, то это значит что какой-то умник вначале или в середине модуля вставил какие-то выражения, а это нельзя. 1С разрешает выражения только в конце модуля
		if len(m.Body) > 0 {
			if _, ok := m.Body[len(m.Body)-1].(*FunctionOrProcedure); !ok {
				yylex.Error("procedure and function definitions should be placed before the module body statements")
				return
			}
		}

		m.Body = append(m.Body, item)
	default:
		m.Body = append(m.Body, item)
	}
}

// func (m Statements) Walk(callBack func(statement *Statement)) {
// 	walkHelper(m, callBack)
// }

func walkHelper(parent *FunctionOrProcedure, statements []Statement, callBack func(current *FunctionOrProcedure, statement *Statement)) {
	ierahiy = ierahiy + 1
	otstup := strings.Repeat(" ", ierahiy)
	fmt.Printf("%swalkHelper parent: %T %v %v\n", otstup, parent, parent, ierahiy)
	fmt.Printf("%swalkHelper statements: %d %T %v\n", otstup, len(statements), statements, statements)
	//fmt.Printf("walkHelper current: %T %v\n", current, current)
	for i, item := range statements {
		fmt.Printf("%swalkHelper item: %d %T %v\n", otstup, i, item, item)
		switch v := item.(type) {
		case *IfStatement:
			fmt.Printf("%sIfStatement %v\n", otstup, v.Expression)
			walkHelper(parent, []Statement{v.Expression}, callBack)
			walkHelper(parent, v.TrueBlock, callBack)
			walkHelper(parent, v.ElseBlock, callBack)
			walkHelper(parent, v.IfElseBlock, callBack)
		case TryStatement:
			fmt.Printf("%sTryStatement\n", otstup)
			walkHelper(parent, v.Body, callBack)
			walkHelper(parent, v.Catch, callBack)
		case *LoopStatement:
			fmt.Printf("%sLoopStatement\n", otstup)
			walkHelper(parent, v.Body, callBack)
		case *FunctionOrProcedure:
			fmt.Printf("%sFunctionOrProcedure\n", otstup)
			walkHelper(v, v.Body, callBack)
			parent = v
		case MethodStatement:
			fmt.Printf("%sMethodStatement\n", otstup)
			walkHelper(parent, v.Param, callBack)
		//case CallChainStatement:
		//	walkHelper(parent, []Statement{v.Unit}, callBack)
		case *ExpStatement:
			fmt.Printf("%sExpStatement\n", otstup)
			walkHelper(parent, []Statement{v.Right}, callBack)
			walkHelper(parent, []Statement{v.Left}, callBack)
		case TernaryStatement:
			fmt.Printf("%sTernaryStatement\n", otstup)
			walkHelper(parent, []Statement{v.Expression}, callBack)
			walkHelper(parent, []Statement{v.TrueBlock}, callBack)
			walkHelper(parent, []Statement{v.ElseBlock}, callBack)
		case *ReturnStatement:
			fmt.Printf("%sReturnStatement\n", otstup)
			walkHelper(parent, []Statement{v.Param}, callBack)
		case *BuiltinFunctionStatement:
			fmt.Printf("%sBuiltinFunctionStatement %v %d\n", otstup, v.Name, len(v.Param))
			walkHelper(parent, v.Param, callBack)
		default:
			fmt.Printf("%sdefault %T %v\n", otstup, v, v)
		}
		fmt.Printf("%scallBack %T %v\n", otstup, parent, statements[i])
		fmt.Printf("%sstatements[i] %T %v\n", otstup, statements[i], &statements[i])
		fmt.Printf("%sstatements[i] %T\n", otstup, *(&statements[i]))
		callBack(parent, &statements[i])
	}
}
