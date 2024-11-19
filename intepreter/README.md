# Marble Intepreter

## Features

- **C-like syntax**
- **Variable bindings**
- **Data types:** integers, floats, booleans, strings, arrays.
- **Arithmetic expressions:** `+`, `-`, `/`, `*`, `>`, `<`, `>=`, `<=`, `==`, `!=`
- **Comments:** `//`
- **Built-in functions:**
  - **`len`**: Get the length of strings and arrays.
  - **`print`**: Write to stdout.
  - **`push`**: Append to arrays.
- **First-class & higher-order functions**
- **Closures**

## Example

```go
var greeting = "Hello";
var name = "Marble!";
print(greeting + ", " + name); // Prints: Hello, Marble!

var pi = 3.14;
var year = 2024;
var isCool = true;

var array = [1, "", true, [], func(x) { x * -x }];
var newArray = push(array, pi, [1, 2]);
print(len(array));    // Prints: 5
print(len(newArray)); // Prints: 7
print(array[-1](2));  // Prints: -4

var calculator = func(operation, x, y) {
    if (operation == "+") {
        return x + y;
    }
    if (operation == "-") {
        return x - y;
    }
    if (operation == "*") {
        return x * y;
    }
    return x / y;
}
if (!(calculator("+", 10, 20) == (10 + 20))) {
    print("Awesome!");
} else {
    print("Oopsie!"); // Prints: Oopsie!
}

var badFibonacci = func(x) {
    if (x <= 1) {
        return x;
    }
    return badFibonacci(x - 1) + badFibonacci(x - 2);
}
print(badFibonacci(3)); // Prints: 2

var multiplier = func(multiplier) {
    return func(x) {
        return x * multiplier;
    }
}
var twice = func(f, x) { f(f(x)) } (multiplier(3), 2);
print(twice); // Prints 18

if (true) {
    if (true) {
        return "Successful!"; // stops execution here!
    }
}

return "Failure!";
```

## Credits

- Writing An Intepreter In Go - Thorsten Bell
