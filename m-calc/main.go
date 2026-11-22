package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

var vars = map[string]float64{}

type MathFunc func(...float64) (float64, error)

var functions = map[string]MathFunc{
	"sqrt":  unary(math.Sqrt),
	"abs":   unary(math.Abs),
	"round": unary(math.Round),
	"floor": unary(math.Floor),
	"ceil":  unary(math.Ceil),
	"sin":   unary(math.Sin),
	"cos":   unary(math.Cos),
	"tan":   unary(math.Tan),
	"log":   unary(math.Log),
	"log10": unary(math.Log10),
	"exp":   unary(math.Exp),
	"pow":   binary(math.Pow),
	"min":   binary(math.Min),
	"max":   binary(math.Max),
}

func unary(fn func(float64) float64) MathFunc {
	return func(args ...float64) (float64, error) {
		if len(args) != 1 {
			return 0, fmt.Errorf("expected 1 argument, got %d", len(args))
		}
		return fn(args[0]), nil
	}
}

func binary(fn func(float64, float64) float64) MathFunc {
	return func(args ...float64) (float64, error) {
		if len(args) != 2 {
			return 0, fmt.Errorf("expected 2 arguments, got %d", len(args))
		}
		return fn(args[0], args[1]), nil
	}
}

// Evaluate expression
func eval(expr string) (float64, error) {
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, err
	}

	value, err := calculate(node)
	if err != nil {
		return 0, err
	}

	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("invalid mathematical result")
	}

	return value, nil
}

// Evaluate AST nodes
func calculate(node ast.Expr) (float64, error) {
	switch n := node.(type) {

	// Numbers
	case *ast.BasicLit:
		if n.Kind != token.INT && n.Kind != token.FLOAT {
			return 0, fmt.Errorf("unsupported literal: %s", n.Value)
		}

		return strconv.ParseFloat(n.Value, 64)

	// Variables and constants
	case *ast.Ident:
		switch n.Name {
		case "pi":
			return math.Pi, nil
		case "e":
			return math.E, nil
		}

		value, ok := vars[n.Name]
		if !ok {
			return 0, fmt.Errorf("undefined variable: %s", n.Name)
		}

		return value, nil

	// Parentheses
	case *ast.ParenExpr:
		return calculate(n.X)

	// Unary operators
	case *ast.UnaryExpr:
		value, err := calculate(n.X)
		if err != nil {
			return 0, err
		}

		switch n.Op {
		case token.SUB:
			return -value, nil
		case token.ADD:
			return value, nil
		}

		return 0, fmt.Errorf("unsupported unary operator: %s", n.Op)

	// Binary operators
	case *ast.BinaryExpr:
		left, err := calculate(n.X)
		if err != nil {
			return 0, err
		}

		right, err := calculate(n.Y)
		if err != nil {
			return 0, err
		}

		switch n.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		case token.REM:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return math.Mod(left, right), nil
		}

		return 0, fmt.Errorf("unsupported operator: %s", n.Op)

	// Function calls
	case *ast.CallExpr:
		name, ok := n.Fun.(*ast.Ident)
		if !ok {
			return 0, fmt.Errorf("invalid function")
		}

		fn, ok := functions[name.Name]
		if !ok {
			return 0, fmt.Errorf("unknown function: %s", name.Name)
		}

		args := make([]float64, len(n.Args))

		for i, arg := range n.Args {
			value, err := calculate(arg)
			if err != nil {
				return 0, err
			}
			args[i] = value
		}

		return fn(args...)
	}

	return 0, fmt.Errorf("unsupported expression: %T", node)
}

// Find assignment operator
func findAssignment(s string) int {
	depth := 0

	for i, c := range s {
		switch c {
		case '(':
			depth++
		case ')':
			depth--
		case '=':
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

// Validate variable name
func validIdentifier(s string) bool {
	if s == "" {
		return false
	}

	for i, c := range s {
		valid := c == '_' ||
			(c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(i > 0 && c >= '0' && c <= '9')

		if !valid {
			return false
		}
	}

	return true
}

// Format result
func format(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// Process input expression
func execute(line string) {
	line = strings.TrimSpace(line)

	if line == "" {
		return
	}

	if idx := findAssignment(line); idx >= 0 {
		name := strings.TrimSpace(line[:idx])
		expr := strings.TrimSpace(line[idx+1:])

		if !validIdentifier(name) {
			fmt.Println("Error: invalid variable name")
			return
		}

		if name == "pi" || name == "e" || functions[name] != nil {
			fmt.Println("Error: reserved name:", name)
			return
		}

		if expr == "" {
			fmt.Println("Error: missing expression")
			return
		}

		value, err := eval(expr)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		vars[name] = value
		fmt.Println(format(value))
		return
	}

	value, err := eval(line)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(format(value))
}

// Display help
func showHelp() {
	fmt.Println(`
m-calc - Simple Expression Calculator

OPERATORS:
  +             Addition
  -             Subtraction
  *             Multiplication
  /             Division
  %             Modulo
  ( )           Parentheses

FUNCTIONS:
  sqrt(x)       Square root
  abs(x)        Absolute value
  round(x)      Round
  floor(x)      Round down
  ceil(x)        Round up

  min(a,b)      Minimum
  max(a,b)      Maximum
  pow(a,b)      Power

  sin(x)        Sine
  cos(x)        Cosine
  tan(x)        Tangent

  log(x)        Natural logarithm
  log10(x)      Base 10 logarithm
  exp(x)        Exponential

CONSTANTS:
  pi            3.141592...
  e             2.718281...

VARIABLES:
  a = 10
  b = 20
  c = a + b
  c

COMMANDS:
  help          Show help
  vars          List variables
  clear         Clear variables
  exit          Exit calculator

EXAMPLES:
  2 + 3
  sqrt(23)
  pow(2, 8)
  a = 12 + 12
  b = sqrt(a)
  a + b

NOTE:
  Trigonometric functions use radians.
`)
}

// Display variables
func showVars() {
	if len(vars) == 0 {
		fmt.Println("No variables defined")
		return
	}

	names := make([]string, 0, len(vars))

	for name := range vars {
		names = append(names, name)
	}

	sort.Strings(names)

	fmt.Println("Variables:")

	for _, name := range names {
		fmt.Printf("  %s = %s\n", name, format(vars[name]))
	}
}

// Main
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 1024*1024)

	fmt.Println("CALC - Simple Calculator")
	fmt.Println("Type help for commands")

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())

		switch line {
		case "":
			continue

		case "exit":
			fmt.Println("Goodbye!")
			return

		case "help":
			showHelp()
			continue

		case "vars":
			showVars()
			continue

		case "clear":
			vars = make(map[string]float64)
			fmt.Println("Variables cleared")
			continue
		}

		execute(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Input error:", err)
	}
}
