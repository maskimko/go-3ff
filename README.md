3ff
========

This utility shows modified Terrafrom resources by comparing two files
Actually this is an open sourced version of Wix's *tfResDif* tool 


How to build?
-------------

In the project directory:
`$ go build` 

How to launch?
--------------
From the help function:

     Usage of ./tfResDif:
        -d    Enable debug output to StdErr
        -t    Output modified resources only (For terraform command)
        -version
              Show version info
     

To compare files command looks just as simple as:

`$ ./tfResDif <source file/dir> <modified file/dir>`
    
To get more verbose output _-d_ key can be used. Output will be printed to stderr. 
Use still can use redirection to /dev/null to suppress it:

`$ ./tfResDif -d <source file/dir> <modified file/dir> 2>/dev/null`
    
To get list of modified resources to be used by terraform"

`$ ./tfResDif -t <source file/dir> <modified file/dir>`


Environment variables
---------------------

*TFRESDIF_NOPB* - Disables output of progress bar (Like _-nopb_ flag)

*TFRESDIF_DEBUG* - Enables debug output (Like _-d_ flag)



Changelog
---------
*v0.0.1* 
> Initial version, which is a truncated copy of Wix's *tfResDif* utility version _1.2.10_
 

