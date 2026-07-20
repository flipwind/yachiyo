package ymath

import (
	"fmt"
	"maps"
	"math"
	"yachiyo/yachiyo-util/yerror"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type Curve interface {
	Verify(x float64, expect float64) (bool, error)
	Calculate(x float64) (float64, error)
}

// ExprCurve

var baseEnv = map[string]any{
	"π": math.Pi,
	"e": math.E,

	"sin": math.Sin,
	"cos": math.Cos,
	"tan": math.Tan,
}

type ExprCurve struct {
	Expression string
	Program    *vm.Program
}

func New(expression string, x float64) (*ExprCurve, error) {
	env := maps.Clone(baseEnv)

    env["x"] = x
	
	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return nil, err
	}

	return &ExprCurve{
		Expression: expression,
		Program:    program,
	}, nil
}

func (c *ExprCurve) Verify(x float64, expect float64) (bool, error) {
	env := maps.Clone(baseEnv)

    env["x"] = x

	const accuracy = 1e-5
	result, err := expr.Run(c.Program, env)

	if err != nil {
		return false, err
	}

	res, ok := result.(float64)
	if ok != true {
		return false, yerror.FieldInvalid(fmt.Sprintf("curve `%s`", c.Expression),
			fmt.Sprintf("need `%v`, got `%v`(%T), which can't be converted into float64", expect, result, result))
	}

	if math.Abs(res-expect) > accuracy {
		return false, nil
	}

	return true, nil
}

func (c *ExprCurve) Calculate(x float64) (float64, error) {
	env := maps.Clone(baseEnv)

    env["x"] = x

	result, err := expr.Run(c.Program, env)
	if err != nil {
		return 0, err
	}

	return result.(float64), nil
}
