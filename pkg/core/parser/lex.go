
//line lex.rl:1
package parser

import (
    "strconv"
)


//line lex.go:11
const lexer_start int = 9
const lexer_first_final int = 9
const lexer_error int = 0

const lexer_en_main int = 9


//line lex.rl:13


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
    
//line lex.go:36
	{
	 lex.cs = lexer_start
	 lex.ts = 0
	 lex.te = 0
	 lex.act = 0
	}

//line lex.rl:29
    return lex
}

func ResetLexer(lex *Lexer, data []byte) {
    lex.pe = len(data)
    lex.data = data
    
//line lex.go:52
	{
	 lex.cs = lexer_start
	 lex.ts = 0
	 lex.te = 0
	 lex.act = 0
	}

//line lex.rl:36
}

func (l *Lexer) Error(msg string) {
	println(msg)
}


func (lex *Lexer) Lex(lval *yySymType) int {
    eof := lex.pe
    var tok int

    
//line lex.go:73
	{
	if ( lex.p) == ( lex.pe) {
		goto _test_eof
	}
	switch  lex.cs {
	case 9:
		goto st_case_9
	case 0:
		goto st_case_0
	case 10:
		goto st_case_10
	case 1:
		goto st_case_1
	case 2:
		goto st_case_2
	case 3:
		goto st_case_3
	case 4:
		goto st_case_4
	case 11:
		goto st_case_11
	case 5:
		goto st_case_5
	case 12:
		goto st_case_12
	case 13:
		goto st_case_13
	case 6:
		goto st_case_6
	case 7:
		goto st_case_7
	case 8:
		goto st_case_8
	case 14:
		goto st_case_14
	case 15:
		goto st_case_15
	case 16:
		goto st_case_16
	case 17:
		goto st_case_17
	case 18:
		goto st_case_18
	case 19:
		goto st_case_19
	case 20:
		goto st_case_20
	case 21:
		goto st_case_21
	case 22:
		goto st_case_22
	case 23:
		goto st_case_23
	case 24:
		goto st_case_24
	case 25:
		goto st_case_25
	case 26:
		goto st_case_26
	case 27:
		goto st_case_27
	case 28:
		goto st_case_28
	case 29:
		goto st_case_29
	case 30:
		goto st_case_30
	case 31:
		goto st_case_31
	case 32:
		goto st_case_32
	}
	goto st_out
tr2:
//line lex.rl:106
 lex.te = ( lex.p)+1
{ lval.str = string(lex.data[lex.ts + 1:lex.te - 1]); tok = IDENT; {( lex.p)++;  lex.cs = 9; goto _out }}
	goto st9
tr4:
//line lex.rl:108
 lex.te = ( lex.p)+1
{ lval.str = string(lex.data[lex.ts + 1:lex.te - 1]); tok = SCONST; {( lex.p)++;  lex.cs = 9; goto _out }}
	goto st9
tr10:
//line NONE:1
	switch  lex.act {
	case 0:
	{{goto st0 }}
	case 2:
	{( lex.p) = ( lex.te) - 1
/* nothing */}
	case 6:
	{( lex.p) = ( lex.te) - 1
 lval.str = string(lex.data[lex.ts:lex.te]); tok = SAY; {( lex.p)++;  lex.cs = 9; goto _out }}
	case 7:
	{( lex.p) = ( lex.te) - 1
 lval.str = string(lex.data[lex.ts:lex.te]); tok = HELLO; {( lex.p)++;  lex.cs = 9; goto _out }}
	case 8:
	{( lex.p) = ( lex.te) - 1
 lval.str = string(lex.data[lex.ts:lex.te]); tok = SHOW; {( lex.p)++;  lex.cs = 9; goto _out }}
	case 9:
	{( lex.p) = ( lex.te) - 1
 lval.str = string(lex.data[lex.ts:lex.te]); tok = KURT; {( lex.p)++;  lex.cs = 9; goto _out }}
	case 10:
	{( lex.p) = ( lex.te) - 1
 lval.str = string(lex.data[lex.ts:lex.te]); tok = KOBAIN; {( lex.p)++;  lex.cs = 9; goto _out }}
	case 12:
	{( lex.p) = ( lex.te) - 1
 lval.str = string(lex.data[lex.ts:lex.te]); tok = IDENT; {( lex.p)++;  lex.cs = 9; goto _out }}
	}
	
	goto st9
tr19:
//line lex.rl:110
 lex.te = ( lex.p)+1
{ lval.str = string(lex.data[lex.ts:lex.te]); tok = TEQ; {( lex.p)++;  lex.cs = 9; goto _out }}
	goto st9
tr24:
//line lex.rl:90
 lex.te = ( lex.p)
( lex.p)--
{ /* do nothing */ }
	goto st9
tr25:
//line lex.rl:92
 lex.te = ( lex.p)
( lex.p)--
{/* nothing */}
	goto st9
tr26:
//line lex.rl:97
 lex.te = ( lex.p)
( lex.p)--
{ lval.str = string(lex.data[lex.ts:lex.te]); tok = SCONST; {( lex.p)++;  lex.cs = 9; goto _out }}
	goto st9
tr27:
//line lex.rl:95
 lex.te = ( lex.p)
( lex.p)--
{ lval.int, _ = strconv.Atoi(string(lex.data[lex.ts:lex.te])); tok = ICONST; {( lex.p)++;  lex.cs = 9; goto _out }}
	goto st9
tr28:
//line lex.rl:94
 lex.te = ( lex.p)
( lex.p)--
{ lval.int, _ = strconv.Atoi(string(lex.data[lex.ts:lex.te])); tok = ICONST; {( lex.p)++;  lex.cs = 9; goto _out }}
	goto st9
tr29:
//line lex.rl:107
 lex.te = ( lex.p)
( lex.p)--
{ lval.str = string(lex.data[lex.ts:lex.te]); tok = IDENT; {( lex.p)++;  lex.cs = 9; goto _out }}
	goto st9
	st9:
//line NONE:1
 lex.ts = 0

//line NONE:1
 lex.act = 0

		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof9
		}
	st_case_9:
//line NONE:1
 lex.ts = ( lex.p)

//line lex.go:241
		switch  lex.data[( lex.p)] {
		case 32:
			goto st10
		case 34:
			goto st1
		case 39:
			goto st3
		case 45:
			goto st4
		case 46:
			goto st5
		case 47:
			goto st6
		case 55:
			goto st15
		case 61:
			goto tr19
		case 72:
			goto st18
		case 75:
			goto st22
		case 83:
			goto st29
		case 95:
			goto tr20
		case 104:
			goto st18
		case 107:
			goto st22
		case 115:
			goto st29
		}
		switch {
		case  lex.data[( lex.p)] < 52:
			switch {
			case  lex.data[( lex.p)] > 13:
				if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 51 {
					goto st15
				}
			case  lex.data[( lex.p)] >= 9:
				goto st10
			}
		case  lex.data[( lex.p)] > 57:
			switch {
			case  lex.data[( lex.p)] > 90:
				if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
					goto tr20
				}
			case  lex.data[( lex.p)] >= 65:
				goto tr20
			}
		default:
			goto st17
		}
		goto st0
st_case_0:
	st0:
		 lex.cs = 0
		goto _out
	st10:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof10
		}
	st_case_10:
		if  lex.data[( lex.p)] == 32 {
			goto st10
		}
		if 9 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 13 {
			goto st10
		}
		goto tr24
	st1:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof1
		}
	st_case_1:
		switch  lex.data[( lex.p)] {
		case 55:
			goto st2
		case 95:
			goto st2
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 51 {
				goto st2
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto st2
			}
		default:
			goto st2
		}
		goto st0
	st2:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof2
		}
	st_case_2:
		switch  lex.data[( lex.p)] {
		case 34:
			goto tr2
		case 36:
			goto st2
		case 95:
			goto st2
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto st2
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto st2
			}
		default:
			goto st2
		}
		goto st0
	st3:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof3
		}
	st_case_3:
		if  lex.data[( lex.p)] == 39 {
			goto tr4
		}
		goto st3
	st4:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof4
		}
	st_case_4:
		switch  lex.data[( lex.p)] {
		case 45:
			goto st11
		case 46:
			goto st5
		}
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st13
		}
		goto st0
	st11:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof11
		}
	st_case_11:
		switch  lex.data[( lex.p)] {
		case 10:
			goto tr25
		case 13:
			goto tr25
		}
		goto st11
	st5:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof5
		}
	st_case_5:
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st12
		}
		goto st0
	st12:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof12
		}
	st_case_12:
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st12
		}
		goto tr26
	st13:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof13
		}
	st_case_13:
		if  lex.data[( lex.p)] == 46 {
			goto st12
		}
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st13
		}
		goto tr27
	st6:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof6
		}
	st_case_6:
		if  lex.data[( lex.p)] == 42 {
			goto st7
		}
		goto st0
	st7:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof7
		}
	st_case_7:
		if  lex.data[( lex.p)] == 42 {
			goto st8
		}
		goto st7
	st8:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof8
		}
	st_case_8:
		switch  lex.data[( lex.p)] {
		case 42:
			goto st8
		case 47:
			goto tr12
		}
		goto st7
tr12:
//line NONE:1
 lex.te = ( lex.p)+1

//line lex.rl:92
 lex.act = 2;
	goto st14
	st14:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof14
		}
	st_case_14:
//line lex.go:471
		if  lex.data[( lex.p)] == 42 {
			goto st8
		}
		goto st7
	st15:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof15
		}
	st_case_15:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 46:
			goto st12
		case 95:
			goto tr20
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto st15
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr28
tr20:
//line NONE:1
 lex.te = ( lex.p)+1

//line lex.rl:107
 lex.act = 12;
	goto st16
tr33:
//line NONE:1
 lex.te = ( lex.p)+1

//line lex.rl:100
 lex.act = 7;
	goto st16
tr39:
//line NONE:1
 lex.te = ( lex.p)+1

//line lex.rl:104
 lex.act = 10;
	goto st16
tr41:
//line NONE:1
 lex.te = ( lex.p)+1

//line lex.rl:103
 lex.act = 9;
	goto st16
tr44:
//line NONE:1
 lex.te = ( lex.p)+1

//line lex.rl:99
 lex.act = 6;
	goto st16
tr46:
//line NONE:1
 lex.te = ( lex.p)+1

//line lex.rl:101
 lex.act = 8;
	goto st16
	st16:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof16
		}
	st_case_16:
//line lex.go:549
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 95:
			goto tr20
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr10
	st17:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof17
		}
	st_case_17:
		if  lex.data[( lex.p)] == 46 {
			goto st12
		}
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st17
		}
		goto tr28
	st18:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof18
		}
	st_case_18:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 69:
			goto st19
		case 95:
			goto tr20
		case 101:
			goto st19
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st19:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof19
		}
	st_case_19:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 76:
			goto st20
		case 95:
			goto tr20
		case 108:
			goto st20
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st20:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof20
		}
	st_case_20:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 76:
			goto st21
		case 95:
			goto tr20
		case 108:
			goto st21
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st21:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof21
		}
	st_case_21:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 79:
			goto tr33
		case 95:
			goto tr20
		case 111:
			goto tr33
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st22:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof22
		}
	st_case_22:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 79:
			goto st23
		case 85:
			goto st27
		case 95:
			goto tr20
		case 111:
			goto st23
		case 117:
			goto st27
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st23:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof23
		}
	st_case_23:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 66:
			goto st24
		case 95:
			goto tr20
		case 98:
			goto st24
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st24:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof24
		}
	st_case_24:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 65:
			goto st25
		case 95:
			goto tr20
		case 97:
			goto st25
		}
		switch {
		case  lex.data[( lex.p)] < 66:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 98 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st25:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof25
		}
	st_case_25:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 73:
			goto st26
		case 95:
			goto tr20
		case 105:
			goto st26
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st26:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof26
		}
	st_case_26:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 78:
			goto tr39
		case 95:
			goto tr20
		case 110:
			goto tr39
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st27:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof27
		}
	st_case_27:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 82:
			goto st28
		case 95:
			goto tr20
		case 114:
			goto st28
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st28:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof28
		}
	st_case_28:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 84:
			goto tr41
		case 95:
			goto tr20
		case 116:
			goto tr41
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st29:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof29
		}
	st_case_29:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 65:
			goto st30
		case 72:
			goto st31
		case 95:
			goto tr20
		case 97:
			goto st30
		case 104:
			goto st31
		}
		switch {
		case  lex.data[( lex.p)] < 66:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 98 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st30:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof30
		}
	st_case_30:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 89:
			goto tr44
		case 95:
			goto tr20
		case 121:
			goto tr44
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st31:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof31
		}
	st_case_31:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 79:
			goto st32
		case 95:
			goto tr20
		case 111:
			goto st32
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st32:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof32
		}
	st_case_32:
		switch  lex.data[( lex.p)] {
		case 36:
			goto tr20
		case 87:
			goto tr46
		case 95:
			goto tr20
		case 119:
			goto tr46
		}
		switch {
		case  lex.data[( lex.p)] < 65:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto tr20
			}
		case  lex.data[( lex.p)] > 90:
			if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
				goto tr20
			}
		default:
			goto tr20
		}
		goto tr29
	st_out:
	_test_eof9:  lex.cs = 9; goto _test_eof
	_test_eof10:  lex.cs = 10; goto _test_eof
	_test_eof1:  lex.cs = 1; goto _test_eof
	_test_eof2:  lex.cs = 2; goto _test_eof
	_test_eof3:  lex.cs = 3; goto _test_eof
	_test_eof4:  lex.cs = 4; goto _test_eof
	_test_eof11:  lex.cs = 11; goto _test_eof
	_test_eof5:  lex.cs = 5; goto _test_eof
	_test_eof12:  lex.cs = 12; goto _test_eof
	_test_eof13:  lex.cs = 13; goto _test_eof
	_test_eof6:  lex.cs = 6; goto _test_eof
	_test_eof7:  lex.cs = 7; goto _test_eof
	_test_eof8:  lex.cs = 8; goto _test_eof
	_test_eof14:  lex.cs = 14; goto _test_eof
	_test_eof15:  lex.cs = 15; goto _test_eof
	_test_eof16:  lex.cs = 16; goto _test_eof
	_test_eof17:  lex.cs = 17; goto _test_eof
	_test_eof18:  lex.cs = 18; goto _test_eof
	_test_eof19:  lex.cs = 19; goto _test_eof
	_test_eof20:  lex.cs = 20; goto _test_eof
	_test_eof21:  lex.cs = 21; goto _test_eof
	_test_eof22:  lex.cs = 22; goto _test_eof
	_test_eof23:  lex.cs = 23; goto _test_eof
	_test_eof24:  lex.cs = 24; goto _test_eof
	_test_eof25:  lex.cs = 25; goto _test_eof
	_test_eof26:  lex.cs = 26; goto _test_eof
	_test_eof27:  lex.cs = 27; goto _test_eof
	_test_eof28:  lex.cs = 28; goto _test_eof
	_test_eof29:  lex.cs = 29; goto _test_eof
	_test_eof30:  lex.cs = 30; goto _test_eof
	_test_eof31:  lex.cs = 31; goto _test_eof
	_test_eof32:  lex.cs = 32; goto _test_eof

	_test_eof: {}
	if ( lex.p) == eof {
		switch  lex.cs {
		case 10:
			goto tr24
		case 11:
			goto tr25
		case 12:
			goto tr26
		case 13:
			goto tr27
		case 7:
			goto tr10
		case 8:
			goto tr10
		case 14:
			goto tr25
		case 15:
			goto tr28
		case 16:
			goto tr10
		case 17:
			goto tr28
		case 18:
			goto tr29
		case 19:
			goto tr29
		case 20:
			goto tr29
		case 21:
			goto tr29
		case 22:
			goto tr29
		case 23:
			goto tr29
		case 24:
			goto tr29
		case 25:
			goto tr29
		case 26:
			goto tr29
		case 27:
			goto tr29
		case 28:
			goto tr29
		case 29:
			goto tr29
		case 30:
			goto tr29
		case 31:
			goto tr29
		case 32:
			goto tr29
		}
	}

	_out: {}
	}

//line lex.rl:115


    return int(tok);
}