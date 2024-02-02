package ast

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type PrintConf struct {
	// Margin отступы (количество пробелов)
	Margin int

	// автоматически расставить скобки в выражениях
	// LispStyle bool
}

type astPrint struct {
	conf *PrintConf
	ast  *AstNode
}

func (ast *AstNode) Print(conf *PrintConf) string {
	if ast == nil {
		return ""
	}

	p := &astPrint{conf: conf, ast: ast}
	return p.print()
}

func (p *astPrint) print() string {
	if len(p.ast.ModuleStatement.Body) == 0 {
		return ""
	}

	result := ""
	for _, node := range p.ast.ModuleStatement.Body {
		if pf, ok := node.(FunctionOrProcedure); ok {
			result += p.printFunctionOrProcedure(&pf) + "\n\n\n"
		}
	}

	return result
}

func (p *astPrint) printFunctionOrProcedure(pf *FunctionOrProcedure) (result string) {
	builder := strings.Builder{}
	defer func() { result = builder.String() }()

	declaration := ""
	if pf.Type == pfTypeFunction {
		declaration = "Функция"
		defer func() { builder.WriteString("\n\nКонецФункции") }()
	} else if pf.Type == pfTypeProcedure {
		declaration = "Процедура"
		defer func() { builder.WriteString("\n\nКонецПроцедуры") }()
	}

	var params []string
	// buildParam := strings.Builder{}
	for _, param := range pf.Params {
		val, def := "", ""
		if param.IsValue {
			val = "Знач "
		}

		if asText := p.printVarStatement(param.Default); asText != "" {
			def = " = " + asText
		}

		params = append(params, val+param.Name+def)
	}

	export := ""
	if pf.Export {
		export = "Экспорт"
	}

	directive := ""
	if pf.Directive != "" {
		directive = "\n" + pf.Directive + "\n"
	}

	depth := 1

	builder.WriteString(directive)
	builder.WriteString(declaration)
	builder.WriteString(" ")
	builder.WriteString(pf.Name)
	builder.WriteString("(")
	builder.WriteString(strings.Join(params, ", "))
	builder.WriteString(")")
	builder.WriteString(" ")
	builder.WriteString(export)
	builder.WriteString("\n")
	builder.WriteString(p.printBody(pf.Body, depth))

	return
}

func (p *astPrint) printVarStatement(v Statement) string {
	switch val := v.(type) {
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case string:
		return fmt.Sprintf("\"%s\"", val)
	case bool:
		return IF[string](val, "Истина", "Ложь")
	case time.Time:
		return fmt.Sprintf(`'%s'`, val.Format("20060102"))
	case CallChainStatement:
		not := IF[string](val.not, "Не ", "")
		return not + p.printCallChainStatement(val)
	case UndefinedStatement:
		return "Неопределено"
	case MethodStatement:
		not := IF[string](val.not, "Не ", "")
		return not + val.Name + "(" + p.printParams(val.Param) + ")"
	case VarStatement:
		return val.Name
	case ItemStatement:
		return p.printVarStatement(val.Object) + "[" + p.printExpression(val.Item) + "]"
	case TernaryStatement:
		return fmt.Sprintf("?(%s, %s, %s)", p.printExpression(val.Expression), p.printExpression(val.TrueBlock), p.printExpression(val.ElseBlock))
	case NewObjectStatement:
		return fmt.Sprintf("Новый %s(%s)", val.Constructor, p.printParams(val.Param))
	default:
		return ""
	}
}

func (p *astPrint) printParams(Params []Statement) (result string) {
	params := make([]string, len(Params), len(Params))
	for i, parm := range Params {
		params[i] = p.printExpression(parm)
	}

	return strings.Join(params, ", ")
}

func (p *astPrint) printBody(items []Statement, depth int) (result string) {
	builder := strings.Builder{}
	defer func() { result = builder.String() }()

	for _, item := range items {
		builder.WriteString("\n")
		builder.WriteString(p.printBodyItem(item, depth))
	}

	builder.WriteString("\n")
	return
}

func (p *astPrint) printBodyItem(item Statement, depth int) (result string) {
	builder := strings.Builder{}
	defer func() { result = builder.String() }()

	spaces := strings.Repeat(" ", p.conf.Margin*depth)
	builder.WriteString(spaces)

	switch v := item.(type) {
	case IfStatement:
		builder.WriteString(p.printIfStatement(&v, depth))
		builder.WriteString(";")
		builder.WriteString("\n")
	case ExpStatement:
		builder.WriteString(p.printExpression(v))
		builder.WriteString(";")
	case LoopStatement:
		builder.WriteString(p.printLoopStatement(&v, depth))
		builder.WriteString(";")
		builder.WriteString("\n")
	case BreakStatement:
		builder.WriteString("Прервать;")
	case ContinueStatement:
		builder.WriteString("Продолжить;")
	case CallChainStatement:
		builder.WriteString(p.printCallChainStatement(v))
		builder.WriteString(";")
	case TryStatement:
		builder.WriteString(p.printTryStatement(v, depth))
		builder.WriteString(";")
		builder.WriteString("\n")
	case ThrowStatement:
		builder.WriteString("ВызватьИсключение")
		if v.Param != nil {
			builder.WriteString(" ")
			builder.WriteString(p.printVarStatement(v.Param))
		}
		builder.WriteString(";")
	case ReturnStatement:
		builder.WriteString("Возврат")
		if v.Param != nil {
			builder.WriteString(" ")
			builder.WriteString(p.printExpression(v.Param))
		}
		builder.WriteString(";")
	case MethodStatement:
		builder.WriteString(p.printVarStatement(v))
		builder.WriteString(";")
	}

	return
}

func (p *astPrint) printIfStatement(expr *IfStatement, depth int) (result string) {
	builder := strings.Builder{}
	defer func() { result = builder.String() }()

	spaces := strings.Repeat(" ", p.conf.Margin*depth)
	defer func() {
		builder.WriteString(spaces)
		builder.WriteString("КонецЕсли")
	}()

	builder.WriteString("Если ")
	builder.WriteString(p.printExpression(expr.Expression))
	builder.WriteString(" Тогда")

	builder.WriteString(p.printBody(expr.TrueBlock, depth+1))
	for _, item := range expr.IfElseBlock {
		builder.WriteString(spaces)
		builder.WriteString("ИначеЕсли ")
		builder.WriteString(p.printExpression(item.Expression))
		builder.WriteString(" Тогда")
		builder.WriteString(p.printBody(item.TrueBlock, depth+1))
	}

	if expr.ElseBlock != nil {
		builder.WriteString(spaces)
		builder.WriteString("Иначе")
		builder.WriteString(p.printBody(expr.ElseBlock, depth+1))
	}

	return
}

func (p *astPrint) printLoopStatement(loop *LoopStatement, depth int) (result string) {
	builder := strings.Builder{}
	defer func() { result = builder.String() }()

	spaces := strings.Repeat(" ", p.conf.Margin*depth)
	if loop.WhileExpr != nil {
		builder.WriteString("Пока ")
		builder.WriteString(p.printExpression(loop.WhileExpr))
		builder.WriteString(" Цикл")
	} else {
		builder.WriteString("Для ")
	}
	defer func() {
		builder.WriteString(spaces)
		builder.WriteString("КонецЦикла")
	}()

	if loop.In != nil {
		builder.WriteString("Каждого ")
		builder.WriteString(loop.For.(string))
		builder.WriteString(" Из ")
		builder.WriteString(p.printExpression(loop.In))
		builder.WriteString(" Цикл")
	}
	if loop.To != nil {
		builder.WriteString(p.printExpression(loop.For))
		builder.WriteString(" По ")
		builder.WriteString(p.printExpression(loop.To))
		builder.WriteString(" Цикл")
	}

	builder.WriteString(p.printBody(loop.Body, depth+1))

	return
}

func (p *astPrint) printExpression(expr Statement) (result string) {
	builder := strings.Builder{}
	defer func() { result = builder.String() }()

	switch v := expr.(type) {
	case ExpStatement:
		if v.not {
			builder.WriteString("Не ")
		}
		if v.unary {
			builder.WriteString("-")
		}

		if v.unary || v.not {
			builder.WriteString("(")
		}

		builder.WriteString(p.printExpression(v.Left))
		builder.WriteString(" ")
		builder.WriteString(v.Operation.String())
		builder.WriteString(" ")
		builder.WriteString(p.printExpression(v.Right))

		if v.unary || v.not {
			builder.WriteString(")")
		}
	case VarStatement:
		if v.not {
			builder.WriteString("Не ")
		}
		if v.unary {
			builder.WriteString("-")
		}
		builder.WriteString(p.printVarStatement(v))
	default:
		builder.WriteString(p.printVarStatement(v))
	}

	return
}

func (p *astPrint) printCallChainStatement(call Statement) (result string) {
	switch v := call.(type) {
	case CallChainStatement:
		if v.Call != nil {
			return p.printCallChainStatement(v.Call) + "." + p.printVarStatement(v.Unit)
		}
	case VarStatement, ItemStatement:
		return p.printVarStatement(call)
	}

	return
}

func (p *astPrint) printTryStatement(try TryStatement, depth int) (result string) {
	builder := strings.Builder{}
	defer func() { result = builder.String() }()

	spaces := strings.Repeat(" ", p.conf.Margin*depth)

	builder.WriteString("Попытка")
	defer func() {
		builder.WriteString(spaces)
		builder.WriteString("КонецПопытки")
	}()

	if try.Body != nil {
		builder.WriteString(p.printBody(try.Body, depth+1))
	} else {
		builder.WriteString("\n")
	}

	builder.WriteString(spaces)
	builder.WriteString("Исключение")
	if try.Catch != nil {
		builder.WriteString(p.printBody(try.Catch, depth+1))
	} else {
		builder.WriteString("\n")
	}

	return
}