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
}

// any non-terminal which returns a value needs a type, which is
// really a field name in the above union struct

// CMDS
%type <node> command

%type<node> say_hello_command show_command
%type<node> kurt_kobain_command

%type<str> reversed_keyword


/* basic words */
%token<str> SAY HELLO

/* stats */
%token<str> SHOW

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
    } | kurt_kobain_command {
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

kurt_kobain_command:
    STOP SYSTEM {
        $$ = &KKBCommand{
        }
    } | KURT KOBAIN {
        $$ = &KKBCommand{
        }
    }
    ;
