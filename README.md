3ff
========

This utility shows modified Terrafrom resources by comparing two files
Actually this is an open sourced version of Wix's *tfResDiff* tool


*3ff* reads as _triff_, stands for _TeRraForm difF_ (TRF is substituted by simply 3) 


How to build?
-------------

In the project directory:
`$ go build` 

How to launch?
--------------
From the help function:

     Usage of ./3ff:
        -d    Enable debug output to StdErr
        -t    Output modified resources only (For terraform command)
        -version
              Show version info
     

To compare files command looks just as simple as:

`$ ./3ff <source file/dir> <modified file/dir>`
    
To get more verbose output _-d_ key can be used. Output will be printed to stderr. 
Use still can use redirection to /dev/null to suppress it:

`$ ./3ff -d <source file/dir> <modified file/dir> 2>/dev/null`
    
To get list of modified resources to be used by terraform"

`$ ./3ff -t <source file/dir> <modified file/dir>`


Environment variables
---------------------

*3ff_DEBUG* - Enables debug output (Like _-d_ flag)



Changelog
---------
*v0.0.1* 
> Initial version, which is a truncated copy of Wix's *3ff* utility version _1.2.10_

*v0.1.4*
> Can be used publicly as a standalone Terraform comparison tool. Also this version was tested to use in tfResDif as a library

*v0.1.5*
> This version can be used publicly as a standalone tool

*v0.1.6*
> This version can be used publicly in other tools as a library as well as aa standalone tool
 
*0.1.8*
> Add GetTfResources method which returns all parsed terraform resource names, which can be used as targets in terraform

*0.1.9*
> Add some terraform HCL block resource type support

*0.1.12*
> Can return resource count

*0.1.13*
> Introduced GetTfResourcesCountByPath function

*0.1.14*
> Use of getCount for all known HCL resource types