<!--
This was generated by org-mode export as Markdown.
If you wish to edit, please edit the file 'readme.org'
and regenerate this document
-->

<div id="table-of-contents">
<h2>Table of Contents</h2>
<div id="text-table-of-contents">
<ul>
<li><a href="#sec-1">1. cb-cli</a>
<ul>
<li><a href="#sec-1-1">1.1. Installation</a>
<ul>
<li><a href="#sec-1-1-1">1.1.1. Source Installation</a></li>
<li><a href="#sec-1-1-2">1.1.2. Binary Installation</a></li>
</ul>
</li>
<li><a href="#sec-1-2">1.2. Basic Usage</a>
<ul>
<li><a href="#sec-1-2-1">1.2.1. Pulling</a></li>
<li><a href="#sec-1-2-2">1.2.2. Pushing</a></li>
<li><a href="#sec-1-2-3">1.2.3. Exporting</a></li>
<li><a href="#sec-1-2-4">1.2.4. Importing</a></li>
</ul>
</li>
<li><a href="#sec-1-3">1.3. Usage Example</a>
</li>
<li><a href="#sec-1-4">1.4. Todo</a></li>
</ul>
</li>
</ul>
</div>
</div>

# cb-cli<a id="sec-1" name="sec-1"></a>

The command line interface for interacting with the [ClearBlade Platform](https://platform.clearblade.com).

This tool is currently in an alpha state and comes with no warantee. Please create issues for any problems you have, so we can continue to improve this tool.

## Installation<a id="sec-1-1" name="sec-1-1"></a>

Installation is easy and can be done two ways depending on whether or not you want to work with the source.

### Source Installation<a id="sec-1-1-1" name="sec-1-1-1"></a>

Installation from source requires that you have a Go toolchain all set up and ready to go. If you do not have Go installed but would like to do so, the instructions for doing this can be found [here](https://golang.org/doc/install). After you have Go all set up, you can run the following command to install the cb-cli tool:

    $ go get github.com/clearblade/cb-cli

This will pull the tool and all the dependencies and install everything. If you have your $GOPATH/bin on your $PATH then you should have cb-cli available in your shell where ever you are. If you do not then you will have to reference it by full path ($GOPATH/bin/cb-cli) or add that directory to your $PATH. To add this to your path run this:

    $ echo 'export PATH="$PATH:$GOPATH/bin"' >> ~/.bashrc
    $ source ~/.bashrc

This of course, will only work if you are using bash. If you are using an alternative shell, replace `.bashrc` in the line above with whatever the init file for your shell is.

### Binary Installation<a id="sec-1-1-2" name="sec-1-1-2"></a>

Installation from the binary distribution is quite simple. Go to the [releases page](https://github.com/ClearBlade/cb-cli/releases) and grab the lates release for your system. We currently only support OS X and Linux. Download the correct archive for your system, and unpack to the location in which you want the binary to be. Add this location to your $PATH with the following command:

    $ echo 'export PATH="$PATH:/path/to/cb-cli/dir"' >> ~/.bashrc
    $ source ~/.bashrc

This of course, will only work if you are using bash. If you are using an alternative shell, replace `.bashrc` in the line above with whatever the init file for your shell is. After this, you will have the cb-cli program anywhere in your shell.

## Basic Usage<a id="sec-1-2" name="sec-1-2"></a>

The tool currently has four sub-commands: `pull`, `push`, `export` and `import`. The first time you run the tool it will prompt you for your username and password. It will get an authentication token, and save it in ~/.cbauth. If you want to change the location of this auth file use the flag `-authinfo="/path/to/auth/file"`.

### Pulling<a id="sec-1-2-1" name="sec-1-2-1"></a>

A basic example of the pull sub-command is:

    $ cb-cli pull aaaaaaaabbbbbbbbccccc111222333

Where aaaaaaaabbbbbbbbccccc111222333 is your systemKey. This will pull all of the services for that system and some meta data and it will put it in a directory named the same as the system name. The result of the example above:

    $ cb-cli pull aaaaaaaabbbbbbbbccccc111222333
    Code for aaaaaaaabbbbbbbbccccc111222333 has been successfully pulled and put in a directory testSystem
    $ tree -a testSystem
    testSystem/
    ├── .meta.json
    ├── testService2.js
    └── testService.js

    0 directories, 3 files

As you can see a directory named 'testSystem' has been created. There are three files in this directory. `.meta.json` is the first file. This file will exist in every system directory. It contains some things that cb-cli needs to make your life easier. Do not edit this file. The other two files are services that I had in my system. You can edit these services locally and then push them back to the server using the other sub-command `push`.

### Pushing<a id="sec-1-2-2" name="sec-1-2-2"></a>

If your current working directory is a system directory (a directory that contains a `.meta.json` file and service files), you can push your local changes to the server using the following command:

    $ cb-cli push
    Push successful

If you have no changes locally, you will get the following error:

    $ cb-cli push
    Error pushing: No services have changed, nothing to push

If you are in the root directory of multiple systems, you will have to specify the systemKey of the system, you want to push:

    $ cb-cli push aaaaaaaabbbbbbbbccccc111222333
    Push successful

### Exporting<a id="sec-1-2-3" name="sec-1-2-3"></a>

The `export` command will pull down all of the metadata (collection info, services, and roles) for a given system. This command must be used before using the `import` command.

Example:

    $ cb-cli export aaaaaaaabbbbbbbbccccc111222333

Where aaaaaaaabbbbbbbbccccc111222333 is the systemKey of the system you wish to export.

### Importing<a id="sec-1-2-4" name="sec-1-2-4"></a>

The `import` command will take a previously exported system and upload it to a platform instance. To use the command navigate to the directory of a previously exported system and run the following command:

    $ cb-cli import

Flags:

    -url = the url of the platform instance that you wish to upload the system (default platform.clearblade.com)
    -importrows = if supplied, the `import` command will migrate all data from collections (default: false)

Advanced example:

    $ cb-cli -url=platforminstance.com -importrows import  
    
## Usage Example<a id="sec-1-3" name="sec-1-3"></a>  

Here is an example of exporting a system from one platform instance to another. First run the following command:  

    $ rm ~/.cbauth  
    
This will remove any previously saved authentication details. Next, make a note of the ___platform URL___ and ___system key___ of the system you wish to export. Then run the following command:  

    $ cb-cli -url=http://your_platform_URL export -exportrows -exportusers <SYSTEM-KEY>
    
Replace the url flag with either http or https depending on your platform and replace the URL with the URL of your platform instance. The ___-exportrows___ flag is used to export all the data in the collections and the ___-exportusers___ flag is used to export all the users in the ___Auth___ tab of the platform console. Also replace the ___SYSTEM-KEY___ with your system key. After hitting "Enter", you will be asked to enter your email and password for your developer account. And if everything goes well you will get a message similar to this: ___System 'YOUR_SYSTEM_NAME' has been exported into directory YOUR_SYSTEM_NAME___. The cb-cli export command will create a new directory in your current working directory with the name of the system you just exported.  

Now, to import this system to another platform instance `cd` to the directory where you have exported your system and execute the following command:  

     $ rm ~/.cbauth  
     
This will remove any previously saved authentication details. Now execute the following command to import your system:

    $ cb-cli -url=http://your_new_platform_url import -importrows -importusers  
    
Replace the url flag with either http or https depending on your platform and replace the URL with the URL of your new platform instance. If everything goes well, you will have your system imported to your new platform instance and ready to go!

## Todo<a id="sec-1-4" name="sec-1-4"></a>

-   Create more tooling around local/server conflicts.
-   Expose more settings to the cb-cli (params, service creation)
-   Allow for local execution of service via local nodejs
