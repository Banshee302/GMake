# GMake
Lightning fast build automation system written in pure GO and similar to Make

**It is primarily written for Windows and is tested in a Windows enviroment but does not use any kernel features so Linux SHOULD work too.**

# Commands
./gmake.exe runs gmake

you should type in a task you wrote in your GMake file, so say ./gmake.exe build which will run the build task if its written in your GMake file.

# Documentation and Syntax

Variables are stored in $, so $compiler will set a variable called compiler.

to write a task you can do this:

task build:
    PRINT = "Building $src..."
    $compiler build -o $bin $src
    OUT: $bin

it can be any name. 

You can print messages using PRINT which expects a = identifier, so say:
PRINT = "Hello World!"

So an example GMake file can be:

$compiler = go
$src = test/test.go
$bin = bin/app.exe

task build:
    PRINT = "Running: $compiler build -o $bin $src"
    $compiler build -o $bin $src
    PRINT = "Build complete."
    OUT: $bin

task clean:
    PRINT = "Cleaning build artifacts..."
    rm -rf bin/*


# Parrarel Execution

GMake supports running Multiple Build Operations in one task for a major speed increase when it comes to building.

heres a simple GMake script that Builds a GO file and uses the PARRARELL flag to run multiple build jobs in one task.

$compiler = go

task build:
    PRINT = "Compiling in parallel..."
    PARALLEL:
        $compiler build -o bin/a.exe a.go
        $compiler build -o bin/b.exe b.go
        $compiler build -o bin/c.exe c.go
    PRINT = "Build complete."

