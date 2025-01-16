%{
package parser


type YpParser yyParser

func NewYpParser() YpParser {
	return yyNewParser()
}

%}

// fields inside this union end up as the fields in a structure known
// as ${PREFIX}SymType, of which a reference is passed to the lexer.
%union {
	str                    string
    int                    int
    node				   Node
    nodeList		       []Node
    bool			       bool
}

// any non-terminal which returns a value needs a type, which is
// really a field name in the above union struct

// CMDS
%type <node> command

%type<node> say_hello_command show_command copy_command
%type<node> kurt_kobain_command

%type<str> reversed_keyword


/* basic words */
%token<str> SAY HELLO

/* stats */
%token<str> SHOW

/* copy */
%token<str> COPY WITH
%type<node> copy_gengeneric_opt_elem copy_generic_opt_arg
%type<bool> opt_boolean
%type<str> ColLabel
%type<nodeList> copy_options copy_gengeneric_opt_list
%token<str> TOPENBR TCLOSEBR TCOMMA
%token<str> FALSE_P TRUE_P 

/* pseudo-sql */
%token<str> SELECT FROM WHERE ORDER BY SORT ASC DESC GROUP

/* misc */
%token<str> KURT KOBAIN STOP SYSTEM

// same for terminals
%token <str> SCONST IDENT
%token <int> ICONST

/* '=' */
%token<str> TEQ

/* ';' != */
%token<str> TSEMICOLON 

%start any_command

%%

any_command:
    command semicolon_opt

semicolon_opt:
    TSEMICOLON {}
    | /*empty*/ {}
    ;

reversed_keyword:
      SAY {$$=$1}
    | HELLO {$$=$1}
    ;


command:
    say_hello_command
    {
        setParseTree(yylex, $1)
    } |
    show_command {
        setParseTree(yylex, $1)
    } | 
    kurt_kobain_command {
        setParseTree(yylex, $1)
    } | 
    copy_command {
        setParseTree(yylex, $1)
    } | /* nothing */ { $$ = nil }

say_hello_command:
    SAY HELLO { $$ = &SayHelloCommand{} } 
    ;

show_command:
    SHOW IDENT {
        $$ = &ShowCommand{
            Type: $2,
        }
    }
    ;

copy_command:
    COPY SCONST WITH copy_options {
        $$ = &CopyCommand{
            Path: $2,
            Options: $4,
        }
    }
    ;

copy_options: TOPENBR copy_gengeneric_opt_list TCLOSEBR { $$ = $2; };

copy_gengeneric_opt_list:
    copy_gengeneric_opt_elem
    {
        $$ = []Node{$1};
    }
    | copy_gengeneric_opt_list TCOMMA copy_gengeneric_opt_elem
    {
        $$ = append($1, $3);
    }
    ;

copy_gengeneric_opt_elem:
    ColLabel copy_generic_opt_arg
    {
        $$ = &Option{
            Name: $1,
            Arg: $2,
        }
    }
    ;

copy_generic_opt_arg:
    opt_boolean			            { $$ = &AExprBConst{Value: $1} }
    | ICONST					    { $$ = &AExprIConst{Value: $1} }
    | SCONST                        { $$ = &AExprSConst{Value: $1} }
    | /* EMPTY */					{  }
    ;

opt_boolean:
    TRUE_P									{ $$ = true }
    | FALSE_P								{ $$ =  false }
    //| NonReservedWord_or_SCONST				{ $$ = $1; }
    ;

ColLabel:	
    IDENT									{ $$ = $1; }
    ;

kurt_kobain_command:
    STOP SYSTEM {
        $$ = &KKBCommand{
        }
    } | KURT KOBAIN {
        $$ = &KKBCommand{
        }
    }
    ;
