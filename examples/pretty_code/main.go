package main

import (
	"fmt"

	"github.com/LazarenkoA/1c-language-parser/ast"
)

func main() {
	code :=
		`&НаКлиенте
Процедура КомандаВыполнитьКоманду(Команда)                    
	//ПослеВыполненияРасчета(2, НЕОПРЕДЕЛЕНО);
	ОповещенияПослеРачсета = Новый ОписаниеОповещения("ПослеВыполненияРасчета", ЭтотОбъект);
	ВыполнитьОбработкуОповещения(ОповещенияПослеРачсета, 10);
КонецПроцедуры //КомандаВыполнитьКоманду      

&НаКлиенте
Процедура ПослеВыполненияРасчета(Резульат, Параметры) экспорт
	Сообщить("Результат расчета="+СокрЛП(Резульат));
конецПроцедуры //`

	a := ast.NewAST(code)
	if err := a.Parse(); err != nil {
		fmt.Println(err)
		return
	}
	jdata, _ := a.JSON()
	fmt.Println(string(jdata))

	//pf := make(map[string]funcInfo)
	a.ModuleStatement.Walk(func(currentFP *ast.FunctionOrProcedure, statement *ast.Statement) {
		fmt.Println("Walk FunctionOrProcedure", currentFP.Name)
		fmt.Printf("Walk statement %T %v\n", statement, statement)

		// Вывод информации о текущей функции/процедуре
		if currentFP != nil {
			fmt.Printf("\nФункция/Процедура: %s\n", currentFP.Name)
			fmt.Printf("  Export: %v\n", currentFP.Export)
			fmt.Printf("  Type: %v\n", currentFP.Type)
		}
		// Вывод информации о statement
		fmt.Printf("\nStatement тип: %T\n", *statement)
		switch s := (*statement).(type) {
		case ast.MethodStatement:
			fmt.Printf("  Имя метода: %s\n", s.Name)
			fmt.Printf("  Аргументы: %v\n", s.Params)
		case ast.CallChainStatement:
			fmt.Printf("  Цепочка вызовов: %v\n", s.Call)
		case ast.StatementType:
			fmt.Printf("  StatementType: %v\n", s)
		case *ast.FunctionOrProcedure:
			fmt.Printf("  FunctionOrProcedure: %v\n", s)
		default:
			fmt.Printf("  Значение: %+v\n", s)
		}
		fmt.Println("------------------------")

		/*if currentFP == nil {
			return
		}

		key := a.ModuleStatement.Name + "." + currentFP.Name
		if _, ok := pf[key]; !ok {
			pf[key] = funcInfo{id: len(pf), export: currentFP.Export, notUse: true, moduleName: m.ModuleStatement.Name}
		}

		v := pf[key]

		switch value := (*statement).(type) {
		case ast.MethodStatement:
			v.dependence = lo.Union(v.dependence, []string{m.ModuleStatement.Name + "." + value.Name})
		case ast.CallChainStatement:
			if value.IsMethod() {
				v.dependence = append(v.dependence, printCallChainStatement(value))
			}
		}

		if f, ok := (*statement).(*ast.FunctionOrProcedure); ok {
			v.stCount = len(f.Body) + 1
		}

		pf[key] = v
		*/
	})

	//fmt.Println(a.Print(ast.PrintConf{Margin: 4}))
}
