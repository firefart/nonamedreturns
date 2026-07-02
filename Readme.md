# nonamedreturns

A Go linter that reports all [named returns](https://go.dev/tour/basics/7) in function and method signatures.

I hate named returns in golang because they are error prone (see [Why are named returns error prone?](#why-are-named-returns-error-prone) below). That's why I wrote this linter. That's all.

Tutorial on how to write your own linter: <https://disaev.me/p/writing-useful-go-analysis-linter/>

## Installation

### As a standalone binary

```sh
go install github.com/firefart/nonamedreturns@latest
```

Prebuilt binaries for Linux, macOS and Windows are also attached to every [release](https://github.com/firefart/nonamedreturns/releases).

### Via golangci-lint

`nonamedreturns` is bundled with [golangci-lint](https://golangci-lint.run/). You do not need to install it separately; just enable it in your configuration (see [Settings](#settings)).

## Usage

### Standalone

The binary is a standard [`go/analysis`](https://pkg.go.dev/golang.org/x/tools/go/analysis) checker, so it accepts the usual package patterns:

```sh
# check the current module
nonamedreturns ./...

# check a single package
nonamedreturns ./internal/foo
```

Any named return value produces a diagnostic such as:

```text
./foo.go:10:1: named return "err" with type "error" found
```

To see all available flags run:

```sh
nonamedreturns -help
```

### golangci-lint

Enable the linter in your `.golangci.yml`:

```yaml
linters:
  enable:
    - nonamedreturns
```

Then run it as part of your normal lint step:

```sh
golangci-lint run ./...
```

## Settings

The linter has the following settings.

### `report-error-in-defer`

|                       |                          |
| --------------------- | ------------------------ |
| **Type**              | `bool`                   |
| **Default**           | `false`                  |
| **Standalone flag**   | `-report-error-in-defer` |
| **golangci-lint key** | `report-error-in-defer`  |

A common, legitimate use of a named return is a named `error` that is inspected or modified inside a `defer` before the function returns. By default the linter does **not** report a named return when **all** of the following are true:

1. its type is exactly the built-in `error`,
2. it is referenced (read or assigned) inside a `defer` closure in the same function, and
3. it is assigned somewhere in the function: explicitly (inside the `defer` or anywhere else in the body, including via `for ... = range`) or implicitly by a `return` statement with result values (e.g. `return callSomething()`), which assigns every named return before the defers run.

Set `report-error-in-defer` to `true` if you want these named errors reported as well.

#### Example that is allowed by default

```go
func doRequest(ctx context.Context) (err error) {
 span := startSpan(ctx)
 defer func() {
  // err is read here to record the outcome
  if err != nil {
   span.RecordError(err)
  }
  span.End()
 }()

 err = callSomething()
 return
}
```

This is **not** reported with the default settings (the named `error` is used in the `defer` and assigned in the body), but **is** reported when `report-error-in-defer: true`. The same applies when the body ends in `return callSomething()` instead of the explicit assignment — the `return` implicitly assigns `err`.

#### Standalone

```sh
nonamedreturns -report-error-in-defer=true ./...
```

#### golangci-lint

```yaml
linters:
  enable:
    - nonamedreturns
  settings:
    nonamedreturns:
      # report named error if it is assigned inside defer
      report-error-in-defer: true
```

### `allow-unused-named-returns`

|                       |                               |
| --------------------- | ----------------------------- |
| **Type**              | `bool`                        |
| **Default**           | `false`                       |
| **Standalone flag**   | `-allow-unused-named-returns` |
| **golangci-lint key** | `allow-unused-named-returns`  |

Named returns are useful as documentation in a signature (for example
`func add(a, b int) (sum int)` tells the reader what the result means). The risk
comes from _using_ them in the body. Set `allow-unused-named-returns` to `true`
to keep that documentation value while forbidding any reliance on the named
return: the name is allowed in the signature but reported when **either**

1. it is referenced anywhere in the body (read or assigned), including inside a
   `defer`, or
2. the function contains a naked `return` (which implicitly populates every
   named return).

> A naked `return` inside a nested closure populates that closure's own results, not the enclosing function's, so it does not trigger a report for the enclosing function.

When this setting is `true` it fully takes over: the default error-in-defer
exemption and the `report-error-in-defer` setting have no effect.

#### Example that is allowed

```go
// documented but never used in the body, returned explicitly
func add(a, b int) (sum int) {
 return a + b
}
```

#### Examples that are reported

```go
// "sum" is referenced in the body
func add(a, b int) (sum int) {
 sum = a + b
 return
}

// naked return implicitly uses "sum"
func add(a, b int) (sum int) {
 return
}
```

#### Standalone

```sh
nonamedreturns -allow-unused-named-returns=true ./...
```

#### golangci-lint

```yaml
linters:
  enable:
    - nonamedreturns
  settings:
    nonamedreturns:
      # allow named returns for documentation, but report them if used in the body
      allow-unused-named-returns: true
```

## Why are named returns error prone?

1. **Shadowing of Variables**
   If you have a named return variable that is also used as a local variable within the function, it can lead to shadowing issues. This occurs when the named return variable is unintentionally shadowed by a local variable with the same name, leading to unexpected behavior.

2. **Accidental Changes**
   Developers might inadvertently modify the value of a named return variable within the function, thinking it only affects the local variable and not the actual return value.

3. **Readability Issues**
   While named returns can improve readability for method signatures, they can also make the code harder to understand if misused. It may be unclear whether a variable is a local variable or a return variable, especially in larger functions.

4. **Unused Variables**
   Named returns often result in variables being declared in the function signature that are not used within the function body. This can lead to confusion and may make the code less maintainable.

5. **Unintentional scope increase**
   Named return variables have scope for the whole function, as opposed to local variable which have scope only after they are defined. So, even if you are setting the value at the last few lines of the function, its scope still spans the whole function.
   This allows the variable to be referenced, before it's assigned the first time.

   #### Example

   ```golang
   func test(input s) (ret bool) {
      ....
      // ret is accessed before it's first assigned,
      // this code will not error out as the variable
      // s already defined.
      if ret {
         // do some stuff
      }
      ....
      ret = someRandomFunc(input) // ret is first assigned here
      }
   }
   ```
