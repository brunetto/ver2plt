ver2plt
=======

### Description

`.ver` to `.plt` files converter

### Installation

You can directly download 64-bit binaries:

* [windows](https://github.com/brunetto/ver2plt/blob/master/ver2plt-WIN.exe)
* [linux](https://github.com/brunetto/ver2plt/blob/master/ver2plt)

otherwise install it with 

````bash
go get https://github.com/brunetto/ver2plt
````
### Use

Just call

(on Windows)
````bash
ver2plt-WIN.exe inputfile.ver
````

(on Linux)
````bash
./ver2plt inputfile.ver
````

where `inputfile.ver` is your `.ver` file, in the folder you are working on, and it will 
produce 

````bash
coords-inputfile.plt  
idxs-inputfile.plt
````
