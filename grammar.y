%{
package main
%}

%type<body> body
%type<opt_body> opt_body
%type<stmt> stmt
%type<stmt_loop> stmt_loop
%type<funcProc> funcProc
%type<stmt_if> stmt_if
%type<opt_elseif_list> opt_elseif_list
%type<opt_else> opt_else
%type<opt_stmt> opt_stmt
%type<opt_param> opt_param
%type<manyfuncProc> manyfuncProc 
%type<exprs> exprs 
%type<expr> expr
%type<opt_export> opt_export
%type<opt_directive> opt_directive 
%type<simple_expr> simple_expr 
%type<declarations_method_params> declarations_method_params
%type<declarations_method_param> declarations_method_param
%type<opt_expr> opt_expr
//%type<opt_exprs> opt_exprs
%type<through_dot> through_dot
%type<new_object> new_object
%type<ternary> ternary
%type<opt_explicit_variables> opt_explicit_variables
%type<explicit_variables> explicit_variables
%type<identifiers> identifiers
%type<stmt_tryCatch> stmt_tryCatch
%type<identifier> identifier


%union {
    token Token
    stmt_if IfStatement
    opt_elseif_list []*IfStatement
    opt_else []Statement
    stmt    Statement
    opt_stmt Statement
    stmt_tryCatch Statement
    stmt_loop LoopStatement
    funcProc FunctionOrProcedure
    body []Statement
    opt_body []Statement
    opt_param Statement
    through_dot Statement
    manyfuncProc []Statement
    declarations_method_params []ParamStatement
    declarations_method_param ParamStatement
    expr Statement
    opt_expr Statement
    //opt_exprs []Statement
    exprs []Statement
    opt_export *Token
    opt_directive *Token
    simple_expr Statement
    new_object Statement
    ternary Statement
    explicit_variables map[string]VarStatement
    opt_explicit_variables map[string]VarStatement
    identifiers []Token
    identifier Statement
}

%token<token> Directive Identifier Procedure Var EndProcedure If Then ElseIf Else EndIf For Each In To Loop EndLoop Break Not ValueParam
%token<token> Continue Try Catch EndTry Number String New Function EndFunction Return Throw NeEq Le Ge Or And True False Undefind Export


//%right '='
%left Or
%left And
%left NeEq
%left Le
%left Ge
%left Not

%right '='
%left Identifier

//%nonassoc NeEq '>' '<'
//%nonassoc NeEq
//%nonassoc Not
%left '>' '<'
%left '+' '-'
%left '*' '/' '%'
%right UNARY

%%

main: manyfuncProc {
    if ast, ok := yylex.(*AstNode); ok {
        ast.ModuleStatement = ModuleStatement{
            Name: "",
            Body: $1,
        }
    }
};

opt_directive:  { $$ = nil}
        | Directive { $$ = &$1}
;

opt_export: { $$ = nil}
        | Export { $$ = &$1}
;

manyfuncProc: funcProc { $$ = []Statement{$1} }
            | manyfuncProc funcProc { $$ = append($1, $2) };

funcProc: opt_directive Function Identifier '(' declarations_method_params ')' opt_export { isFunction(true, yylex) } opt_explicit_variables opt_body EndFunction
        {  
            $$ = createFunctionOrProcedure(NodeTypeFunction, $1, $3.literal, $5, $7, $9, $10)
            isFunction(false, yylex) 
        }
        | opt_directive Procedure Identifier '(' declarations_method_params ')' opt_export opt_explicit_variables opt_body EndProcedure
        { 
            $$ = createFunctionOrProcedure(NodeTypeProcedure, $1, $3.literal, $5, $7, $8, $9)
        }
;

opt_body: { $$ = nil }
	| body { $$ = $1 }
;
    
body: stmt { $$ = []Statement{$1} }
    | body semicolon opt_stmt { 
        if $3 != nil {
            $$ = append($$, $3) 
        }
    }
;

opt_stmt: { $$ = nil }
        | stmt { $$ = $1 }
;

/* переменные */ 
opt_explicit_variables: { $$ = map[string]VarStatement{} }
            | explicit_variables { $$ = $1 }
;

explicit_variables: Var identifiers semicolon { 
                    if vars, err := appendVarStatements(map[string]VarStatement{}, $2); err != nil {
                        yylex.Error(err.Error()) 
                    } else {
                        $$ = vars
                    }
                }
            | explicit_variables Var identifiers semicolon {
                    if vars, err := appendVarStatements($1, $3); err != nil {
                        yylex.Error(err.Error()) 
                    } else {
                        $$ = vars
                    }
                }
;

/* Если Конецесли */
stmt_if : If expr Then opt_body opt_elseif_list opt_else EndIf {  
    $$ = IfStatement {
        Expression: $2,
        TrueBlock:  $4,
        IfElseBlock: $5,
        ElseBlock: $6,
    }
};

/* ИначеЕсли */
opt_elseif_list : { $$ = []*IfStatement{} }
        | ElseIf expr Then opt_body opt_elseif_list { 
             $$ = append($5, &IfStatement{
                Expression: $2,
                TrueBlock:  $4,
            })
        };

/* Иначе */
opt_else : { $$ = nil }
        | Else opt_body { $$ = $2 };

/* тернарный оператор */
ternary: '?' '(' expr comma expr comma expr ')' { 
    $$ = TernaryStatement{
            Expression: $3,
            TrueBlock: $5,
            ElseBlock: $7,
        } 
};

/* циклы */
stmt_loop: For Each Identifier In through_dot Loop { setLoopFlag(true, yylex) } opt_body EndLoop { 
    $$ = LoopStatement{
        For: $3.literal,
        In: $5,
        Body: $8,
    }
    setLoopFlag(false, yylex) 
} 
| For expr To expr Loop { setLoopFlag(true, yylex) } opt_body EndLoop {
     $$ = LoopStatement{
        For: $2,
        To: $4,
        Body: $7,
    }
    setLoopFlag(false, yylex)
};

stmt : expr { $$ = $1 }
    | stmt_if { $$ = $1 }
    | stmt_loop {$$ = $1 }
    | stmt_tryCatch { $$ = $1 }
    | Continue { $$ = ContinueStatement{}; checkLoopOperator($1, yylex) }
    | Break { $$ = BreakStatement{}; checkLoopOperator($1, yylex) }
    | Throw opt_param { $$ = ThrowStatement{ Param: $2 }; checkThrowParam($1, $2, yylex) }
    | Return opt_param { $$ = ReturnStatement{ Param: $2 }; checkReturnParam($2, yylex) }
;

opt_param: { $$ = nil } 
            | simple_expr { $$ = $1 }
;


/* вызовы через точку */
through_dot: identifier { $$ = $1 }
        | through_dot dot identifier { $$ = CallChainStatement{ Unit: $3, Call:  $1 } }
;

identifier: Identifier { $$ = VarStatement{ Name: $1.literal } }
        | Identifier '(' exprs ')' { $$ = MethodStatement{ Name: $1.literal, Param: $3 } }
        | identifier '[' expr ']' { $$ = ItemStatement{ Object: $1, Item: $3 } }
;


/* попытка */
stmt_tryCatch: Try opt_body Catch { setTryFlag(true, yylex) } opt_body EndTry { 
    $$ = TryStatement{ Body: $2, Catch: $5 }
    setTryFlag(false, yylex)
};

/* выражения */
expr : simple_expr { $$ = $1 }
    |'(' expr ')' { $$ = $2 }
    | expr '+' expr { $$ = ExpStatement{Operation: OpPlus, Left: $1, Right: $3} }
    | expr '-' expr { $$ = ExpStatement{Operation: OpMinus, Left: $1, Right: $3} }
    | expr '*' expr { $$ = ExpStatement{Operation: OpMul, Left: $1, Right: $3} }
    | expr '/' expr { $$ = ExpStatement{Operation: OpDiv, Left: $1, Right: $3} }
    | expr '%' expr { $$ = ExpStatement{Operation: OpMod, Left: $1, Right: $3} }
    | expr '>' expr { $$ = ExpStatement{Operation: OpGt, Left: $1, Right: $3} }
    | expr '<' expr { $$ = ExpStatement{Operation: OpLt, Left: $1, Right: $3} }
	| expr '=' expr { $$ = ExpStatement{Operation: OpEq, Left: $1, Right: $3 } }
    | '-' expr %prec UNARY { $$ = unary($2) }
    | expr Or expr {  $$ = ExpStatement{Operation: OpOr, Left: $1, Right: $3 } } 
    | expr And expr { $$ = ExpStatement{Operation: OpAnd, Left: $1, Right: $3 } } 
    | expr NeEq expr { $$ = ExpStatement{Operation: OpNe, Left: $1, Right: $3 } }
    | expr Le expr { $$ = ExpStatement{Operation: OpLe, Left: $1, Right: $3 } }
    | expr Ge expr { $$ = ExpStatement{Operation: OpGe, Left: $1, Right: $3 } }
    | Not expr { $$ = not($2) }
    | new_object { $$ = $1 } 
;

opt_expr: { $$ = nil } | expr { $$ = $1 };
//opt_exprs: { $$ = nil } | exprs { $$ = $1 };


// опиасываются правила по которым можно объявлять параметры в функции или процедуре
declarations_method_param: Identifier {  $$ = *(&ParamStatement{}).Fill(nil, $1) } // обычный параметр
            | ValueParam Identifier { $$ = *(&ParamStatement{}).Fill(&$1, $2) } // знач
            | declarations_method_param '=' simple_expr { $$ = *($$.DefaultValue($3)) } // необязательный параметр 
;

declarations_method_params : { $$ = []ParamStatement{} }
                | declarations_method_param  { $$ = []ParamStatement{$1} }
                | declarations_method_params comma declarations_method_param { $$ = append($1, $3) }
;

// для ключевого слова Новый
new_object:  New Identifier { $$ = NewObjectStatement{ Constructor: $2.literal } }
            | New Identifier '(' exprs ')' { $$ = NewObjectStatement{ Constructor: $2.literal, Param: $4 } }
;

simple_expr:  String { $$ = $1.value  }
            | Number { $$ =  $1.value }
            | True { $$ =  $1.value  }
            | False { $$ =  $1.value  }
            | Undefind { $$ = $1.value  }
            | through_dot { 
                if tok, ok := $1.(Token); ok {
                    $$ = tok.literal
                } else {
                    $$ =  $1
                }
            }
            | ternary { $$ =  $1  } // тернарный оператор
;

exprs : opt_expr {
    if $1 != nil {
        $$ = []Statement{$1} 
    } else {
        $$ = nil;
    }
}
	| exprs comma opt_expr { $$ = append($1, $3) }
;    

identifiers: Identifier { $$ = []Token{$1} }
        | identifiers comma Identifier {$$ = append($$, $3) }
;

semicolon: ';' 
comma: ','
dot: '.'

%%