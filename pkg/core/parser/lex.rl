package parser

import (
    "strconv"
)

%%{ 
    machine lexer;
    write data;
    access lex.;
    variable p lex.p;
    variable pe lex.pe;
}%%

type Lexer struct {
	data         []byte
	p, pe, cs    int
	ts, te, act  int

	result []string
}

func NewLexer(data []byte) *Lexer {
    lex := &Lexer{ 
        data: data,
        pe: len(data),
    }
    %% write init;
    return lex
}

func ResetLexer(lex *Lexer, data []byte) {
    lex.pe = len(data)
    lex.data = data
    %% write init;
}

func (l *Lexer) Error(msg string) {
	println(msg)
}


func (lex *Lexer) Lex(lval *yySymType) int {
    eof := lex.pe
    var tok int

    %%{
        # /* digit = [0-9] ; already defined */


#        xcstart		=	\/\*{op_chars}*;
#        xcstop		=	\*+\/;
#        xcinside	=	[^*/]+;

        integer = digit+;
        ninteger = '-' integer;
        param = '$' integer;
        
        decimal	= ((digit*'.'digit+)|(digit+'.'digit*));
        real = (decimal)|('-'decimal);

        ident_start	=	[A-Za-z\200-\377_];
        ident_cont	=	[A-Za-z\200-\377_0-9$];

        identifier	=	ident_start ident_cont*;

        qidentifier	=	'"' ident_start ident_cont* '"' ;


#        space		=	[ \t\n\r\f];
        horiz_space	= [ \t\f];
        newline		=	[\n\r];
        non_newline	=	[^\n\r];

        sql_comment = '-''-' non_newline*;
        c_style_comment = '/''*' (any - '*''/')* '*''/';
        comment		= sql_comment | c_style_comment;


#       whitespace	=	({space}+|{comment});
        whitespace	=	space+;


        op_chars	=	( '~' | '!' | '@' | '#' | '^' | '&' | '|' | '`' | '?' | '+' | '-' | '*' | '\\' | '%' | '<' | '>' | '=' ) ;
        operator	=	op_chars+;

        sconst = '\'' (any-'\'')* '\'';
        
        main := |*
            whitespace => { /* do nothing */ };
            # integer const is string const 
            comment => {/* nothing */};

            integer =>  { lval.int, _ = strconv.Atoi(string(lex.data[lex.ts:lex.te])); tok = ICONST; fbreak;};
            ninteger => { lval.int, _ = strconv.Atoi(string(lex.data[lex.ts:lex.te])); tok = ICONST; fbreak;};

            real =>  { lval.str = string(lex.data[lex.ts:lex.te]); tok = SCONST; fbreak;};

            /SAY/i => { lval.str = string(lex.data[lex.ts:lex.te]); tok = SAY; fbreak;};
            /HELLO/i => { lval.str = string(lex.data[lex.ts:lex.te]); tok = HELLO; fbreak;};
            /SHOW/i => { lval.str = string(lex.data[lex.ts:lex.te]); tok = SHOW; fbreak;};

            /COPY/i => { lval.str = string(lex.data[lex.ts:lex.te]); tok = COPY; fbreak;};
            /WITH/i => { lval.str = string(lex.data[lex.ts:lex.te]); tok = WITH; fbreak;};

            /KURT/i => { lval.str = string(lex.data[lex.ts:lex.te]); tok = KURT; fbreak;};
            /KOBAIN/i => { lval.str = string(lex.data[lex.ts:lex.te]); tok = KOBAIN; fbreak;};

            /STOP/i => { lval.str = string(lex.data[lex.ts:lex.te]); tok = STOP; fbreak;};
            /SYSTEM/i => { lval.str = string(lex.data[lex.ts:lex.te]); tok = SYSTEM; fbreak;};

            qidentifier      => { lval.str = string(lex.data[lex.ts + 1:lex.te - 1]); tok = IDENT; fbreak;};
            identifier      => { lval.str = string(lex.data[lex.ts:lex.te]); tok = IDENT; fbreak;};
            sconst      => { lval.str = string(lex.data[lex.ts + 1:lex.te - 1]); tok = SCONST; fbreak;};

            '=' => { lval.str = string(lex.data[lex.ts:lex.te]); tok = TEQ; fbreak;};
            '(' => { lval.str = string(lex.data[lex.ts:lex.te]); tok = TOPENBR; fbreak;};
            ')' => { lval.str = string(lex.data[lex.ts:lex.te]); tok = TCLOSEBR; fbreak;};
            ',' => { lval.str = string(lex.data[lex.ts:lex.te]); tok = TCOMMA; fbreak;};

            /true/i => { lval.str = string(lex.data[lex.ts:lex.te]); tok = TRUE_P; fbreak;};
            /false/i => { lval.str = string(lex.data[lex.ts:lex.te]); tok = FALSE_P; fbreak;};

        *|;

        write exec;
    }%%

    return int(tok);
}