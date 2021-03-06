package main

var compileArg = map[string]string{
	".cpp": "g++ -std=c++11 -O2 -Wl,-stack_size,0x10000000 {{.Name}}.cpp -o {{.Name}}",
	".d":   "dmd -m64 -O {{.Name}}.d",
	".cs":  "mcs -r:System.Numerics {{.Name}}.cs",
}

var testArg = map[string]string{
	".cpp": "./{{.Name}} --seed={{.Seed}} --hash={{.Hash}}",
	".d":   "./{{.Name}} --seed={{.Seed}} --hash={{.Hash}}",
}

var execArg = map[string]string{
	".cpp": "./{{.Name}}",
	".d":   "./{{.Name}}",
	".cs":  "mono {{.Name}}.exe",
}

var checkerArg = map[string]string{
	".cpp": "./{{.Name}} {{.Input}} {{.Output}} {{.Answer}}",
}
