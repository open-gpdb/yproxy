package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yezzey-gp/yproxy/pkg/core/parser"
)

func TestParse(t *testing.T) {
	assert := assert.New(t)

	type tcase struct {
		query string
		exp   parser.Node
		err   error
	}

	/*  */
	for _, tt := range []tcase{
		{
			query: "say hello",
			exp:   &parser.SayHelloCommand{},
			err:   nil,
		},
		{
			query: "show clients",
			exp: &parser.ShowCommand{
				Type: "clients",
			},
			err: nil,
		},
		{
			query: "show connections",
			exp: &parser.ShowCommand{
				Type: "connections",
			},
			err: nil,
		},
		{
			query: `STOP SYSTEM`,
			exp:   &parser.KKBCommand{},
			err:   nil,
		},
		{
			query: `COPY '/prefix' WITH (CONFIG '/etc/yproxy/old_config.yaml', PORT 6002)`,
			exp: &parser.CopyCommand{
				Path: "/prefix",
				Options: []parser.Node{
					&parser.Option{Name: "CONFIG", Arg: &parser.AExprSConst{Value: "/etc/yproxy/old_config.yaml"}},
					&parser.Option{Name: "PORT", Arg: &parser.AExprIConst{Value: 6002}},
				},
			},
			err: nil,
		},
	} {
		tmp, err := parser.Parse(tt.query)

		assert.NoError(err, "query %s", tt.query)

		assert.Equal(tt.exp, tmp, "query %s", tt.query)
	}
}
