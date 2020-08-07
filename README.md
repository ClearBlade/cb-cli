#### Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Developer workflow](#developer-workflow)
- [File Structure](#file-structure)
- [Commands](#commands)
- [Full Example](#full-example)
- [DevOps](#devops)

# Overview

The ClearBlade CLI tool provides easy to use commands for interacting with ClearBlade platform

1. allows for easy promotion of system changes through dev, staging, QA, production systems
2. allows for source control
3. allows for integration with CI
4. allows for code transpilation (TypeScript, ES6+, etc. for the backend and TypeScript, React, Vue, etc. for the frontend)
5. developers can use their favorite IDE/text editor
6. write unit tests with your favorite test runner

# Installation

#### Binary Installation

Go to [cb-cli releases](https://github.com/ClearBlade/cb-cli/releases) and download latest release for your platform (OSX/Linux)

- Unpack the archive to a location of your choise
- Add path of your unpacked archive to \$PATH

#### Source Installation

For source installation, GO should be installed. If its not, install go from: [goLang](https://golang.org/doc/install)

After go is installed, run:

```
go get github.com/clearblade/cb-cli
```

Either add path to cb-cli executable to \$PATH in bashrc or execute using full path
Add path.These commands work if your are using bash:

```
$ echo 'export PATH="$PATH:$GOPATH/bin"' >> ~/.bashrc
$ source ~/.bashrc
```

or - execute using full path:
\$GOPATH/bin/cb-cli

# Developer workflow:

ClearBlade recommends developers to work in isolated systems when working on a team. The steps below outline how to achieve this.

## Initialize repo

When a project is first started, a ClearBlade system and a source control repo should be created.

1. create master system in platform via console UI
2. create git repo

```
git clone
cd my-cloned-repo
cb-cli export -exportrows -exportusers
echo ".cb-cli" >> .gitignore // ignore any system-specific changes
git commit -am "init commit"
git push origin master
```

## how to work in isolated development system

Once the root system has been created, developers can clone the system with the `import` command. After importing, the `init` command must be run to point to the newly created system. Once targeted, developers will make changes against the isolated system.

```
git checkout -b feature-branch
cb-cli import -importrows -importusers
cb-cli init // point to the newly imported system; this will update system.json with new system info. NOTE: the changes to system.json shouldn't be merged into the master branch
```

## how to update local system with changes from git repo

The steps below outline how to pull in changes from other developers

```
git checkout master
git pull origin master
git checkout feature-branch
git merge master
cb-cli init // target feature-branch system
cb-cli push -all // push all changes from master into feature-branch system
```

# File Structure

Before we start making changes and pushing to platform, we need to export the system

The directory structure after export will look as:

    |_.cb-cli
    | |_cbmeta
    | |_map-name-to-id
    | | |_collections.json
    | | |_roles.json
    | | |_users.json
    |_code
    | |_libraries
    | |_services
    | | |_TestPull
    | | | |_TestPull.js
    | | | |_TestPull.json
    |_data
    | |_myCollection.json
    |_roles
    | |_Administrator.json
    | |_Anonymous.json
    | |_Authenticated.json
    |_shared-caches
    | |_mySharedCache.json
    |_system.json
    |_timers
    | |_myTimer.json
    |_triggers
    | |_myTrigger.json
    |_users
    | |_roles
    | | |_admin@clearblade.com.json
    | |_admin@clearblade.com.json
    | |_schema.json
    |_devices
    | |_roles
    | | |_myDevice.json
    | |_myDevice.json
    | |_schema.json
    |_portals
    | |_myPortal
    | | |_myPortal.json
    | | |_config
    | | | |_widgets
    | | | |_datasources
    | | | |_internalResources

## MetaData

After exporting the system, .cbmeta is created in the exported folder as shown above. This file contains devEmail, platformURL, assetRefreshDates and an auth token for developer access to the system

# Commands

- [export](#export)
- [init](#init)
- [import](#import)
- [pull](#pull)
- [push](#push)
- [test](#test)
- [update](#update)
- [create](#create)
- [delete](#delete)

## Export

**cb-cli export** - Brings a ClearBlade platform System to the local development environment

### Synopsis

```
cb-cli export
	[-exportrows]
	[-exportusers]
```

### Description

This command is used to download all of your system assets to your current directory. It can be used in two ways:

1. Initialize and download the system in your current working directory (as opposed to running `cb-cli init` then `cb-cli export`)
2. Download the system in a previously initialized directory

### Options

- **url**  
   The full url of the ClearBlade platform eg: https://platform.clearblade.com

- **system-key**  
   The system key for the system being brought local

- **messaging-url**  
   The messaging url for the system being brought local

- **email**  
   The email address of the developer working on the system (you)

- **cleanup**  
   Clears all directories prior to performing the export.

- **exportrows**  
   This not only exports the collection objects, but also export all “rows” (or items) in each collection. Be very careful when using this option as it may be unfeasable to export very large collections.

- **exportusers**  
   This exports the data (minus passwords) from the system’s users table. If not present on the command line, only the users table schema is exported.

- **exportitemid**  
   When exporting data collections, this option indicates that the item_id column should also be exported with each row.

- **sort-collections**  
   This option, when specified, will sort the rows of an exported data collection by item_id. This is useful when using a version control system and you wish to view the differences between two versions of a data collection.

- **data-page-size**  
   When exporting the rows of a data collection and there are a large number of rows (> 100k), it is adviseable to increase the number of rows constituting a page. This will improve the performance of the export by decreasing the number of queries against the database.

Once completed, all of the services, collections, timers, triggers, etc will reside in current repo. Meta data for all objects is in pretty-printed json format. In addition, the actual code for services and libraries is in javascript (.js) format.

You can shortcut the cb-cli init/cb-cli export steps by just calling cb-cli export outside of a repo. This will do a combination of init and export. You can either provide the init options on the command line or you will be prompted for them. This is a common way to begin working on a system locally.

#### EXAMPLES

`cb-cli export`

`cb-cli export -exportrows -exportusers`

## init

**cb-cli init** - Initializes a ClearBlade system locally

### Synopsis

```
cb-cli init
	[-url = <URL>]
	[-system-key = <SYSTEM_KEY>]
	[-email = <DEVELOPER_EMAIL>]
	[-password = <PASSWORD>]
        [-skip-update-map-name-to-id]
```

### Description

The init command is the first command to be run when wanting to work on a ClearBlade system locally.

On successful completion of this command a skeleton tree structure is created under your current working directory, with the root of the tree named after the ClearBlade system you just init-ed. Note that if the system name had spaces, they are converted to underscores. Inside the root directory of the system, two special files are created:

- **.cbmeta** This holds information specific to the developer. It is used to streamline authentication so that you don’t have to enter emails, passwords, and tokens for all future commands.

- **system.json** This contains information specific to the system you’re working on (system name, system key, etc.)

The directory structure for a system looks like this (for a system named “Outstanding System”):

```
Outstanding_System/
	|- .cbmeta
	|- system.json
	|- code/
	|- libraries/
	|- services/
	|- data/
	|- roles/
	|- timers/
	|- triggers/
	|- users/
```

Once you’ve successfully executed the init command, you should cd into and live in the root of the repo when running all future commands.

### Options

- **url**  
   The full url of the ClearBlade platform eg: https://platform.clearblade.com

- **system-key**  
   The system key for the system being brought local

- **messaging-url**  
   The messaging url for the system being brought local

- **email**  
   The email address of the developer working on the system (you)

- **password**  
   Your password for the ClearBlade platform system

- **skip-update-map-name-to-id**
  Set this to true to skip pulling the IDs for roles, collections, and users. This is useful if the system has lots of these types of assets and the goal is to retrieve the schema for the tables after initialization.

You can specify all, some, or none of these options on the command line itself. For those that you didn’t specify, The system will prompt you for the values.

## Import

**cb-cli import** - Takes the assets stored locally and creates a new ClearBlade platform System with the same structure and assets

### Synopsis

```
cb-cli import
	[-url = <URL>]
	[-email = <EMAIL>]
	[-password = <PASSWORD>]
	[-importrows]
	[-importusers]
```

### Description

The import command is run from inside an existing repo for a system. It creates an entirely new system, perhaps on a different clearblade platform instance. Think of it as cloning the system somewhere else. A common use would be as follows. Suppose you’re developing and testing a ClearBlade system inside your private development sandbox. When the system is ready to be deployed to production, you would use the import command to effectively push it into production.

_Note: Only assets that a currently local are imported into the new system._

### Options

- **url**  
   The URL of the destination system (ie, where the new system should be). If you don’t specify this option on the command line, you will be prompted for it.

- **email**  
   The developer’s email on the destination system. If you don’t specify this option on the command line, you will be prompted for it.

- **password**  
   The developer’s password. If you don’t specify this option (we recommend you don’t), you will be prompted for it.

- **importrows**  
   By default, collection rows (items) are not imported. Pass this option to import all items.

- **importusers**  
   By default, the users are not imported into the new system. If you set this option, the users will imported, but their passwords will all be set to “password”, since we don’t transfer passwords back and forth between systems.

Once this command is completed, the newly imported system is fully-functional except for the importusers caveat mentioned above.

#### Examples

`cb-cli import -url="https://platform.clearblade.com"`

`cb-cli import -email="foo@clearblade.com" -password="foo"`
`cb-cli import -email="foo@clearblade.com" -password="foo" -importrows`

## Push

**cb-cli push** - Send the local development versions of assets back to the ClearBlade platform system

### Synopsis

```
cb-cli push
	[-all]
	[-all-services]
	[-all-libraries]
	[-all-edges]
	[-all-devices]
	[-all-portals]
	[-all-plugins]
	[-all-adapters]
	[-all-collections]
	[-all-roles]
	[-all-users]
	[-all-triggers]
	[-all-timers]
	[-all-deployments]
	[-all-shared-caches]
	[-all-webhooks]
	[-all-external-databases]
	[-userschema]
	[-edgeschema]
	[-deviceschema]
	[-service=<SERVICE_NAME>]
	[-library=<LIBRARY_NAME>]
	[-collection=<COLLECTION_NAME>]
	[-user=<EMAIL>]
	[-role=<ROLE_NAME>]
	[-trigger=<TRIGGER_NAME>]
	[-timer=<TIMER_NAME>]
	[-edge=<EDGE_NAME>]
	[-device=<DEVICE_NAME>]
	[-portal=<PORTAL_NAME>]
	[-plugin=<PLUGIN_NAME>]
	[-adapter=<ADAPTER_NAME>]
	[-deployment=<DEPLOYMENT_NAME>]
	[-shared-cache=<SHARED_CACHE_NAME>]
	[-webhook=<WEBHOOK_NAME>]
	[-external-database=<EXTERNAL_DATABASE_NAME>]
```

### Description

The push command allows you upload changes to local copies of ClearBlade objects back out the the remote ClearBlade system. Obviously, it is the opposite of the pull command. Again, it has the same options as the pull command.

You can combine these options on a single command line just like with pull

### Options

- **all-services**  
   Pushes all the services stored in a local repo

- **all-libraries**  
   Pushes all of the libraries stored in a local repo

- **all-edges**  
   Pushes all of the edges stored in a local repo

- **all-devices**  
   Pushes all of the devices stored in a local repo

- **all-portals**  
   Pushes all of the portals stored in a local repo

- **all-plugins**  
   Pushes all of the plugins stored in a local repo

- **all-adapters**  
   Pushes all of the adapters stored in a local repo. Includes adapter metadata as well as all files associated with each adapter.

- **userschema**  
   Pushes the local version of the users table schema to a remote ClearBlade system.

- **edgeschema**  
   Pushes the local version of the edge table schema to a remote ClearBlade system.

- **deviceschema**  
   Pushes the local version of the device table schema to a remote ClearBlade system.

- **service=< service_name >**  
   Pushes the local version of a specific service to a remote ClearBlade system.

- **library=< library_name >**  
   Pushes the local version of a specific library to a remote ClearBlade system.

- **collection=< collection_name >**  
   Pushes the local version of a specific collections' meta-data to a remote ClearBlade system.

- **user=< email >**  
   Pushes the local version of the user record to a remote ClearBlade system. Also Pushes the roles assigned to a user.

- **role=< role_name >**  
   Pushes all the capability details of the specific role to a remote ClearBlade system.

- **trigger=< trigger_name >**  
   Pushes the local version of a specific trigger to a remote ClearBlade system.

- **timer=< timer_name >**  
   Pushes the local version of a specific timer to a remote ClearBlade system.

- **edge=< edge_name >**  
   Pushes the local version of a specific edge to a remote ClearBlade system.

- **device=< device_name >**  
   Pushes the local version of a specific device to a remote ClearBlade system.

- **portal=< portal_name >**  
   Pushes the local version of a specific portal to a remote ClearBlade system.

- **plugin=< plugin-name >**  
   Pushes the local version of a specific plugin to a remote ClearBlade system.

- **adapter=< adapter-name >**  
   Pushes the local version of a specific adapter to a remote ClearBlade system. Includes the adapter metadata as well as the files associated with the adapter.

#### Examples

`cb-cli push`

`cb-cli push -all`

`cb-cli push -all-services`

`cb-cli push -collection=MyCollection`

## Update

TODO

## Create

TODO

## Delete

TODO

## Pull

**cb-cli pull** - Brings the latest versions of assets from the ClearBlade platform System to the local development environment

### Synopsis

```
cb-cli pull
	[-all]
	[-all-services]
	[-all-libraries]
	[-all-edges]
	[-all-devices]
	[-all-portals]
	[-all-plugins]
	[-all-adapters]
	[-all-deployments]
	[-all-collections]
	[-all-roles]
	[-all-users]
	[-all-triggers]
	[-all-timers]
	[-all-shared-caches]
	[-all-webhooks]
	[-all-external-databases]
	[-userschema]
	[-service=<SERVICE_NAME>]
	[-library=<LIBRARY_NAME>]
	[-collection=<COLLECTION_NAME>]
	[-user=<EMAIL>]
	[-role=<ROLE_NAME>]
	[-trigger=<TRIGGER_NAME>]
	[-timer=<TIMER_NAME>]
	[-edge=<EDGE_NAME>]
	[-device=<DEVICE_NAME>]
	[-portal=<PORTAL_NAME>]
	[-plugin=<PLUGIN_NAME>]
	[-adapter=<ADAPTER_NAME>]
	[-deployment=<DEPLOYMENT_NAME>]
	[-shared-cache=<SHARED_CACHE_NAME>]
	[-webhook=<WEBHOOK_NAME>]
	[-external-database=<EXTERNAL_DATABASE_NAME>]
```

### Description

The pull command allows you to selectively grab a specific object (eg a specific code service or library) from the associated ClearBlade system and pull it down to your local repo. This is useful when (for example) multiple developers are working on the same code service. When one developer modifies the code service, you can pull it down and make modifications to the latest version.

### Options

- **all-services**  
   Pulls all of the services stored in the repo

- **all-libraries**  
   Pulls all of the libraries stored in the repo

- **all-edges**  
   Pulls all of the edges stored in the repo

- **all-devices**  
   Pulls all of the devices stored in the repo

- **all-portals**  
   Pulls all of the portals stored in the repo

- **all-plugins**  
   Pulls all of the plugins stored in the repo

- **all-adapters**  
   Pulls all of the adapters stored in the repo. Includes adapter metadata as well as all files associated with each adapter.

- **userschema**  
   Pulls the remote version of the users table schema to a local repository.

- **service=< service_name >**  
   Pulls the remote version of a specific service to a local repository.

- **library=< library_name >**  
   Pulls the remote version of a specific library to a local repository.

- **collection=< collection_name >**  
   Pulls the remote version of a specific collections' meta-data to a local repository.

- **user=< email >**  
   Pulls the remote version of a specific user record to a local repository. Also Pulls the roles assigned to a user.

- **role=< role_name >**  
   Pulls all the capability details of the specific role to a local repository.

- **trigger=< trigger_name >**  
   Pulls the remote version of a specific trigger to a local repository.

- **timer=< timer_name >**  
   Pulls the remote version of a specific timer to a local repository.

- **edge=< edge_name >**  
   Pulls the remote version of a specific edge to a local repository.

- **device=< device_name >**  
   Pulls the remote version of a specific device to a local repository.

- **portal=< portal_name >**  
   Pulls the remote version of a specific portal to a local repository.

- **plugin=< plugin-name >**  
   Pulls the remote version of a specific plugin to a local repository.

- **adapter=< adapter-name >**  
   Pulls the remote version of a specific adapter to a local repository. Includes the adapter metadata as well as the files associated with the adapter.

#### Example

`cb-cli pull`

`cb-cli pull -all`

`cb-cli pull -all-services`

`cb-cli pull -collection=MyCollection`

## Test

**cb-cli test** - Execute a code service, update code service from your local file system, and send MQTT message

### Synopsis

```
cb-cli test
	[-service=<SERVICE_NAME>]
	[-params=<PARAMS>]
	[-topic = <TOPIC>]
	[-payload = <PAYLOAD>]
	[-push]
```

### Description

The test command allows you execute your code services from your local machine, along with update the code itself, and send MQTT Messages

You can combine these options on a single command line just like with pull

### Options

- **service = < service_name >**  
   Executes the selected service. If your local version has newer changes than the cloud platform, we recommend using `-push` flag to push new changes.

- **params = < params >**
  The payload is the parameters for the code service request. This must be valid JSON.

- **topic = < topic >**
  Sends an MQTT Message with the provided topic. The payload contains the MQTT message payload

- **payload = < payload >**
  Payload for the respective service or MQTT message. If `-service` flag is used, then . If -topic is used, then the payload is the MQTT payload.

- **push**
  -push flag pushes the local version of the respective code service prior to executing

#### Examples

`cb-cli test -service=MyService`

# Full Example

## Step 1: Developer uses init to create initial directory structure

**Command:**

    $ cb-cli init

- Platform URL : - System Key : - Developer Email : - Password :

**Result:**
A new directory will be created named your System Name. Inside there will be a single file called .cbmeta

    -/<SYSTEM_NAME>

|- .cbmeta

## Step 2: Developer exports their system locally

This process will bring all the schema and asset definitions to your local environment

**Command:**

    $ cd <SYSTEM_NAME>
    $ cb-cli export

**Result:**
Your folder should now be filled with a structure that looks like

    -/<SYSTEM_NAME>
      |- .cb-cli
      	  |- .cbmeta
      |- system.json
      |- services
          |- helloworld
    	  	|- helloworld.js
    		|- helloworld.json
      |- libraries
      |- data
      |- roles
      |- users
      |- triggers
      |- timers

## Step 3: Developer modifies a local service

This step represents a typical developer activity of making a modification to a service. For this task modify a service of your choice

    function helloworld(req, resp){
    	// COMMENT CHANGED!!!!
    	resp.success("hello world: ");
    }

## Step 4: Developer views the differences

**Action:**

    $ git diff path/to/servicename.js

**Result:**

    the differences in your file will be listed

## Step 5: Developer tests the new service

This is an optional step but often developers want to ensure that their changes work

**Action:**

    $ cb-cli test -service=helloworld params="{'name':'Bill'}"

**Result:**

The latest service source code is uploaded to the system and then execute with the parameters passed in.
The response will be similar to

**NOTE: Testing a service will push your service to your system. This actvity should be done against a Test System**

## Step 6: Developer pushes their changes back to the ClearBlade system

Next if the developer has diffs that have not been tested and already sent the server they can now push those changes

**NOTE: Now is a good time to commit your code to a source control repository!**

**Action:**

    cb-cli push -service=helloworld

**Results:**

Upon completion you will receive a list of all the files successully pushed up to the ClearBlade System

## Step 7: Pulling the latest code

In a team environment you may want to pull latest changes that others have running on the system. You can accomplish this using

- a source control branch
- the asset running in the system

To get the asset running in the system locally use the pull command

**Action:**

    cb-cli pull -service=anotherworld

**Result:**
Your file system is now updated with the latest javascript in the anotherworld.js file.

### Summary

The process for developing a local system continues with these above steps. Done correctly you should include a source control repository tool and best practice for back-up and change history purposes.

To learn more about the devops lifecycle the CLI also supports see the DevOps Example

# DevOps

In addition to supporting local development on services and systems the CLI also provides critical capability to integrate your current DevOps processes with the ClearBlade platform.

The below describes a typical agile development team working within the ClearBlade Platform

- Multiple Developers working with their own Systems
- A ClearBlade platform System for a shared Development instance
- A ClearBlade platform System for Test and QA
- A ClearBlade platform System for Production

To support this environment ClearBlade expects a mature use of a source control environment where developers are able to work in shared isolation. - git, svn, cvs and others are supported with the same pattern

## Task: Developer works in local branch

During this period the developer will use their own ClearBlade platform System and use the standard local development process.

When ready to move forward with a feature the developer will commit their work to a source control development branch with-in the source control environment.

## Task: Promotion of feature branch to Development

With the development source control branch updated current devops build system can take control. This involves the following steps

1.  A build system checks out the latest version of the development branch

    \$ git pull origin development

2.  The build system ensures they are init'd into the correct development system

    \$ cb-cli init

3.  The build system promotes any code that is now in the stream

    \$ cb-cli push

At this point the development system is now running the latest code running as captured in source control

### Summary

This process can be leveraged by Dev-Ops tools to continue the promotion and roll back of any environment
