#### Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Developer workflow](#developer-workflow)
- [File Structure](#file-structure)
- [Commands](#commands)
- [Full Example](#full-example)
- [DevOps](#devops)

# Overview

ClearBlade's CLI tool provides easy-to-use commands for interacting with the ClearBlade Platform. 

ClearBlade's CLI tool:

1. Allows easy promotion of system changes through dev, staging, QA, and production systems.
2. Allows source control and integration with CI.
3. Allows code transpilation (TypeScript, ES6+, etc. for the backend, and TypeScript, React, Vue, etc., for the frontend).
4. Allows developers to use their favorite IDE/text editor.
5. Allows you to write unit tests with your favorite test runner.

# Installation

#### Binary installation

Go to [cb-cli releases](https://github.com/ClearBlade/cb-cli/releases) and download the latest release for your platform (OSX/Linux)

- Unpack the archive to a location of your choice.
- Add the path of your unpacked archive to \$PATH.

#### Source installation

Go should be installed. If it's not, install it from: [goLang](https://golang.org/doc/install)

After Go is installed, run:

```
go get github.com/clearblade/cb-cli
```

You may need to specify the version you want. e.g.:

```
go get github.com/clearblade/cb-cli@9.10.2
```

Add path to cb-cli, executable to \$PATH in bashrc, or execute using full path.

These commands work if you are using bash:

```
$ echo 'export PATH="$PATH:$GOPATH/bin"' >> ~/.bashrc
$ source ~/.bashrc
```

Or execute using full path:
\$GOPATH/bin/cb-cli

# Developer workflow:

ClearBlade recommends developers work in isolated systems when working on a team. The steps below outline how to achieve this.

## Initialize repo

A ClearBlade system and a source control repo should be created when a project starts.

1. Create the master system in the Platform via the console UI.
2. Create the Git repo.

```
git clone
cd my-cloned-repo
cb-cli export -exportrows -exportusers
echo ".cb-cli" >> .gitignore // ignore any system-specific changes
git commit -am "init commit"
git push origin master
```

## How to work in an isolated development system

Once the root system has been created, developers can clone the system with the import command. After importing, the init command must be run to point to the newly created system. Once targeted, developers will make changes against the isolated system.

```
git checkout -b feature-branch
cb-cli import -importrows -importusers
cb-cli init // point to the newly imported system; this will update system.json with new system info. The changes to system.json shouldn't be merged into the master branch.
```

## How to update the local system with changes from the Git repo

The steps below outline how to pull in changes from other developers.

```
git checkout master
git pull origin master
git checkout feature-branch
git merge master
cb-cli init // target feature-branch system
cb-cli push -all // push all changes from master into feature-branch system
```

# File Structure

We must export the system before making changes and pushing it to the platform.

The directory structure after export will look as:

    |_.cb-cli
    | |_cbmeta
    | |_map-name-to-id
    | | |_collections.json
    | | |_roles.json
    | | |_users.json
    |_adapters
    | |_myAdapter
    | | |_files
    | | | |_myAdapterFile1
    | | | | |_myAdapterFile1 
    | | | | |_myAdapterFile1.json
    | | | |_myAdapterFile2
    | | | | |_myAdapterFile2
    | | | | |_myAdapterFile2.json
    | | |_myAdapter.json
    |_bucket-set-files
    |_bucket-sets
    | |_myBucketSet.json
    |_code
    | |_libraries
    | |_services
    | | |_TestPull
    | | | |_TestPull.js
    | | | |_TestPull.json
    |_data
    | |_myCollection.json
    |_deployments
    | |_myDeployment.json
    |_devices
    | |_roles
    | | |_myDevice.json
    | |_myDevice.json
    | |_schema.json
    |_edges
    | |_schema.json
    | |_myEdge.json
    |_external-databases
    | |_myDB.json
    |_file-stores
    | |_myFileStore.json
    |_file-stores-files
    | |_myFileStore
    | | |_myfile.txt
    |_message-history-storage
    | |_storage.json
    |_message-type-triggers
    | |_triggers.json
    |_plugins
    | |_myPlugin.json
    |_portals
    | |_myPortal
    | | |_myPortal.json
    | | |_config
    | | | |_widgets
    | | | |_datasources
    | | | |_internalResources
    |_roles
    | |_Administrator.json
    | |_Anonymous.json
    | |_Authenticated.json
    |_secrets
    | |_mysecret.json
    |_shared-caches
    | |_mySharedCache.json
    |_timers
    | |_myTimer.json
    |_triggers
    | |_myTrigger.json
    |_users
    | |_roles
    | | |_admin@clearblade.com.json
    | |_admin@clearblade.com.json
    | |_schema.json
    |_webhooks
    | |_myWebhook.json
    |_system.json


## Metadata

After exporting the system, .cbmeta is created in the exported folder, as shown above. This file contains devEmail, platformURL, assetRefreshDates, and an auth token for developer access to the system.

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
- [remote](#remote)

## Export

**cb-cli export**: Brings a ClearBlade system to the local development environment

### Synopsis

```
cb-cli export
	[-exportrows]
	[-exportusers]
```

### Description

This command downloads all your system assets to your current directory. It can be used in two ways:

1. Initialize and download the system in your current working directory (as opposed to running cb-cli init and cb-cli export).
2. Download the system in a previously initialized directory.

### Options

- **url**
  The ClearBlade Platform's full URL, e.g., https://platform.clearblade.com

- **system-key**
  The system key for the system being brought locally

- **messaging-url**
  The messaging URL for the system being brought locally

- **email**
  The email of the developer working on the system

- **cleanup**
  Clears all directories before performing the export

- **exportrows**
  This exports the collection's objects, rows, and items. Be careful when using this option, as it may be unfeasible to export vast collections.

- **exportusers**
  This exports the data (minus passwords) from the system’s users table. Only the user's table schema is exported if it is not on the command line.

- **exportitemid**
  This option indicates that the item_id column should be exported with each row when exporting data collections.

- **sort-collections**
  This option, when specified, will sort the rows of an exported data collection by item_id. This is useful when using a version control system, and you wish to view the differences between a data collection's versions.

- **data-page-size**
  When exporting a data collection's rows and there are a large number of rows (> 100k), it is advisable to increase the number of rows constituting a page. This will improve the export performance by decreasing the number of queries against the database.

Once completed, all the services, collections, timers, triggers, etc., will reside in the current repo. Metadata for all objects is in pretty-printed JSON format. In addition, the actual code for services and libraries is in JavaScript (.js) format.

You can shortcut the cb-cli init/cb-cli export steps by calling cb-cli export outside a repo. This will do a combination of init and export. You can provide the init options on the command line, or you will be prompted for them. This is a common way to begin working on a system locally.

#### EXAMPLES

`cb-cli export`

`cb-cli export -exportrows -exportusers`

## init

**cb-cli init**: Initializes a ClearBlade system locally

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

The init command is the first command to run when working on a ClearBlade system locally.

Upon completing this command, a skeleton tree structure is created under your current working directory with the tree's root named after the ClearBlade system you init-ed. If the system name had spaces, they were converted to underscores. Inside the system's root directory, two special files are created:

- **.cbmeta** This holds information specific to the developer. It is used to streamline authentication so you don’t have to enter emails, passwords, and tokens for all future commands.

- **system.json** This contains information specific to the system you’re working on (system name, system key, etc.)

The directory structure for a system looks like this (for a system named Outstanding System):

```
Outstanding_System/
	|- .cbmeta
	|- system.json
  |- adapters/
  |- bucket-set-files/
  |- bucket-sets/
	|- code/
	| |- libraries/
	| |- services/
	|- data/
  |- deployments/
  |- devices/
  | |-roles/
  |- edges/
  |- external-databases/
  |- file-stores/
  |- file-stores-files/
  |- plugins/
  |- portals/
	|- roles/
  |- secrets/
  |- shared-caches/
	|- timers/
	|- triggers/
	|- users/
  | |- roles/
  |- webhooks/
```

Once successfully executing the init command, you should cd into and live in the repo's root when running all future commands.

### Options

- **url**
  The ClearBlade Platform's full URL, e.g., https://platform.clearblade.com

- **system-key**
  The system key for the system being brought local

- **messaging-url**
  The messaging URL for the system being brought local

- **email**
  The email of the developer working on the system

- **password**
  Your ClearBlade system's password

- **skip-update-map-name-to-id**
  Set this to true: skip pulling the IDs for roles, collections, and users. This is useful if the system has many of these assets and the goal is to retrieve the table schemas after initialization.

You can specify all, some, or none of these options on the command line. The system will prompt you for the values for those you didn't select.

## Import

**cb-cli import**: Takes the assets stored locally and creates a new ClearBlade system with the same structure and assets.

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

The import command is run from inside an existing repo for a system. It creates a new system, such as a different ClearBlade Platform instance. It is like cloning the system somewhere else. A common use would be as follows: suppose you’re developing and testing a ClearBlade system inside your private development sandbox. When the system is ready to be deployed to production, you will use the import command to push it into production.

Note: Only local assets are imported into the new system.

### Options

- **url**
  The destination system's URL (i.e., where the new system should be). If you don’t specify this option on the command line, you will be prompted for it.

- **email**
  The developer’s email on the destination system. If you don’t specify this option on the command line, you will be prompted for it.

- **password**
  The developer’s password. If you don’t specify this option (we recommend you don’t), you will be prompted for it.

- **importrows**
  By default, collection rows (items) are not imported. Pass this option to import all items.

- **importusers**
  By default, the users are not imported into the new system. If you set this option, the users will be imported, but their passwords will all be set to “password” since we don’t transfer passwords back and forth between systems.

Once this command is completed, the newly imported system is fully-functional except for the above importusers caveat.

#### Examples

`cb-cli import -url="https://platform.clearblade.com"`

`cb-cli import -email="foo@clearblade.com" -password="foo"`
`cb-cli import -email="foo@clearblade.com" -password="foo" -importrows`

## Push

**cb-cli push**: Send the asset local development versions back to the ClearBlade system

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
  [-all-deployments]
	[-all-collections]
	[-all-roles]
	[-all-users]
	[-all-triggers]
	[-all-timers]
	[-all-shared-caches]
	[-all-webhooks]
	[-all-external-databases]
        [-all-bucket-sets]
        [-all-bucket-set-files]
  [-all-user-secrets]
  [-all-file-stores]
  [-all-file-store-files]
	[-userschema]
	[-edgeschema]
	[-deviceschema]
  [-message-history-storage]
	[-service=<SERVICE_NAME>]
	[-library=<LIBRARY_NAME>]
	[-collection=<COLLECTION_NAME>]
  [-collectionschema=<COLLECTION_NAME>]
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
	[-bucket-set=<BUCKET_SET_NAME>]
    	[-bucket-set-files=<BUCKET_SET_NAME>]
    	[-box=<inbox | outbox | sandbox>]
    	[-file=<FILE_NAME>]
  [-file-store=<FILE_STORE_NAME>]
      [-file-store-files=<FILE_STORE_NAME>]
      [-file-store-file=<FILE_NAME>]
  [-user-secret=<SECRET_NAME>]
```

### Description

The push command allows you to upload changes to local copies of ClearBlade objects to the remote ClearBlade system. It is the opposite of the pull command and has the same options.

You can combine these options on a single command line, like with pull.

### Options

- **all-services**
  Pushes all the services stored in a local repo

- **all-libraries**
  Pushes all the libraries stored in a local repo

- **all-edges**
  Pushes all the edges stored in a local repo

- **all-devices**
  Pushes all the devices stored in a local repo

- **all-portals**
  Pushes all the portals stored in a local repo

- **all-plugins**
  Pushes all the plugins stored in a local repo

- **all-adapters**
  Pushes all the adapters stored in a local repo. Includes adapter metadata and all files associated with each adapter

- **all-bucket-sets**
  Pushes all the bucket sets stored in a local repo. Does not include bucket set files (use -all-bucket-set-files for that)

- **all-bucket-set-files**
  Pushes all the bucket set files stored in a local repo

- **all-file-stores**
  Pushes all the file stores stored in a local repo. Does not include file store files (use -all-file-store-files for that)

- **all-file-store-files**
  Pushes all the file store files stored in a local repo

- **userschema**
  Pushes the local version of the users' table schema to a remote ClearBlade system

- **message-history-storage**
  Pushes the local version of the message history storage to a remote ClearBlade system

- **edgeschema**
  Pushes the local version of the edge table schema to a remote ClearBlade system

- **deviceschema**
  Pushes the local version of the device table schema to a remote ClearBlade system

- **service=< service_name >**
  Pushes the local version of a specific service to a remote ClearBlade system

- **library=< library_name >**
  Pushes the local version of a specific library to a remote ClearBlade system

- **collection=< collection_name >**
  Pushes the local version of a specific collection's metadata to a remote ClearBlade system

- **user=< email >**
  Pushes the local version of the user record to a remote ClearBlade system. Also, pushes the roles assigned to a user

- **role=< role_name >**
  Pushes all the capability details of the specific role to a remote ClearBlade system

- **trigger=< trigger_name >**
  Pushes the local version of a specific trigger to a remote ClearBlade system

- **timer=< timer_name >**
  Pushes the local version of a specific timer to a remote ClearBlade system

- **edge=< edge_name >**
  Pushes the local version of a specific edge to a remote ClearBlade system

- **device=< device_name >**
  Pushes the local version of a specific device to a remote ClearBlade system

- **portal=< portal_name >**
  Pushes the local version of a specific portal to a remote ClearBlade system

- **plugin=< plugin-name >**
  Pushes the local version of a specific plugin to a remote ClearBlade system

- **adapter=< adapter-name >**
  Pushes the local version of a specific adapter to a remote ClearBlade system. Includes the adapter metadata and the files associated with the adapter

- **bucket-set=< bucket-set-name >**
  Pushes the local version of a specific bucket set to a remote ClearBlade system. Does not include the bucket set's files (use -bucket-set-files for that)

- **bucket-set-files=< bucket-set-name >**
  Pushes all the specific bucket set files

- **box=< inbox | outbox | sandbox >**
  Pushes all the files inside a specific box for a specific bucket set. Must be used with -bucket-set-files

- **file=< file-path-relative-to-box >**
  Pushes a specific file within a specific box for a specific bucket set. Must be used with -bucket-set-files and -box

- **file-store=< file-store-name >**
  Pushes the local version of a specific file store to a remote ClearBlade system. Does not include the file store's files (use -file-store-files for that)

- **file-store-files=< file-store-name >**
  Pushes all the specific file store files

- **file-store-file=< file-path >**
  Pushes a specific file within a specific file store. Must be used with -file-store-files

#### Examples

`cb-cli push`

`cb-cli push -all`

`cb-cli push -all-services`

`cb-cli push -collection=MyCollection`

`cb-cli push -bucket-set-files=MyBucketSet`

`cb-cli push -bucket-set-files=MyBucketSet -box=inbox`

`cb-cli push -bucket-set-files=MyBucketSet -box=inbox -file=relative/path/from/inbox`

`cb-cli push -file-store-files=MyFileStore`

`cb-cli push -file-store-files=MyFileStore -file-store-file=path/within/my/file/store.txt`


## Update

TODO

## Create

TODO

## Delete

TODO

## Pull

**cb-cli pull**: Brings the latest asset versions from the ClearBlade system to the local development environment

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
  [-all-bucket-sets]
  [-all-bucket-set-files]
  [-all-file-stores]
  [-all-file-store-files]
  [-all-user-secrets]
	[-userschema]
  [-deviceschema]
  [-edgeschema]
  [-message-history-storage]
	[-service=<SERVICE_NAME>]
	[-library=<LIBRARY_NAME>]
  [-collection=<COLLECTION_NAME>]
  [-collectionschema=<COLLECTION_NAME>]
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
	[-bucket-set=<BUCKET_SET_NAME>]
  [-bucket-set-files=<BUCKET_SET_NAME>]
    [-box=<inbox | outbox | sandbox>]
    [-file=<FILE_NAME>]
  [-file-store-files=<FILE_STORE_NAME>]
    [-file-store-file=<FILE_NAME>]
  [-user-secret=<SECRET_NAME>]
```

### Description

The pull command allows you to selectively grab a specific object (e.g., a particular code service or library) from the associated ClearBlade system and pull it down to your local repo. This is useful when multiple developers are working on the same code service. When one developer modifies the code service, you can pull it down and modify the latest version.

### Options
- **all**
  Pulls all assets stored in the system to a local repo

- **all-services**
  Pulls all the services stored in the system to a local repo

- **all-libraries**
  Pulls all the libraries stored in the system to a local repo

- **all-edges**
  Pulls all the edges stored in the system to a local repo

- **all-devices**
  Pulls all the devices stored in the system to a local repo

- **all-portals**
  Pulls all the portals stored in the system to a local repo

- **all-plugins**
  Pulls all the plugins stored in the system to a local repo

- **all-adapters**
  Pulls all the adapters stored in the system to a local repo. Includes adapter metadata and all files associated with each adapter

- **all-deployments**
  Pulls all the deployments stored in the system to a local repo

- **all-collections**
  Pulls all the collections stored in the system to a local repo

- **all-roles**
  Pulls all the roles stored in the system to a local repo

- **all-users**
  Pulls all the users stored in the system to a local repo

- **all-triggers**
  Pulls all the triggers stored in the system to a local repo

- **all-timers**
  Pulls all the timers stored in the system to a local repo

- **all-shared-caches**
  Pulls all the shared-caches stored in the system to a local repo

- **all-webhooks**
  Pulls all the webhooks stored in the system to a local repo

- **all-external-databases**
  Pulls all the external-databases stored in the system to a local repo

- **all-bucket-sets**
  Pulls all the remote bucket sets stored in a system to a local repo. Does not include bucket set files (use -all-bucket-set-files for that)

- **all-bucket-set-files**
  Pulls all the files of all remote bucket sets stored in a system to a local repo

- **all-file-stores**
  Pulls all the remote file stores in a system to a local repo. Does not include file store files (use -all-file-store-files for that)

- **all-file-store-files**
  Pulls all the files of all remote file stores in a system to a local repo

- **all-user-secrets**
  Pulls all the secrets stored in the system to a local repo

- **userschema**
  Pulls the remote version of the users' table schema to a local repository

- **devicechema**
  Pulls the remote version of the device table schema to a local repository

- **edgeschema**
  Pulls the remote version of the edge table schema to a local repository

- **message-history-storage**
  Pulls the remote version of the message history storage to a local repository

- **service=< service_name >**
  Pulls the remote version of a specific service to a local repository

- **library=< library_name >**
  Pulls the remote version of a specific library to a local repository

- **collection=< collection_name >**
  Pulls the remote version of a specific collection's metadata to a local repository

- **collectionschema=< collection_name >**
  Pulls the remote version of a specific collections' schema to a local repository

- **user=< email >**
  Pulls the remote version of a specific user record to a local repository. Also, pulls the roles assigned to a user

- **role=< role_name >**
  Pulls all the capability details of the specific role to a local repository

- **trigger=< trigger_name >**
  Pulls the remote version of a specific trigger to a local repository

- **timer=< timer_name >**
  Pulls the remote version of a specific timer to a local repository

- **edge=< edge_name >**
  Pulls the remote version of a specific edge to a local repository

- **device=< device_name >**
  Pulls the remote version of a specific device to a local repository

- **portal=< portal_name >**
  Pulls the remote version of a specific portal to a local repository

- **plugin=< plugin-name >**
  Pulls the remote version of a specific plugin to a local repository

- **adapter=< adapter-name >**
  Pulls the remote version of a specific adapter to a local repository. Includes the adapter metadata and the files associated with the adapter

- **deployment=< deployment-name >**
  Pulls the remote version of a specific deployment to a local repository

- **shared-cache=< shared-cache-name >**
  Pulls the remote version of a specific deployment to a local repository

- **webhook=< webhook-name >**
  Pulls the remote version of a specific webhook to a local repository

- **external-database=< external-database-name >**
  Pulls the remote version of a specific external-database to a local repository

- **bucket-set=< bucket-set-name >**
  Pulls the remote version of a specific bucket set to a local repository. Does not include the bucket set's files (use -bucket-set-files for that)

- **bucket-set-files=< bucket-set-name >**
  Pulls all the files for a specific remote bucket set

- **box=< inbox | outbox | sandbox >**
  Pulls all the files inside a specific box for a specific bucket set. Must be used with -bucket-set-files

- **file=< file-path-relative-to-box >**
  Pulls a specific file within a specific box for a specific bucket set. Must be used with -bucket-set-files and -box

- **file-store=< file-store-name >**
  Pulls the remote version of a specific file store to a local repository. Does not include the file store's files (use -file-store-files for that)

- **file-store-files=< file-store-name >**
  Pulls all the files for a specific remote file store

- **file-store-file=< file-path >**
  Pulls a specific file within a specific file store. Must be used with -file-store-files

- **user-secret=< secret-name >**
  Pulls the remote version of a specific secret to a local repository.

#### Example

`cb-cli pull`

`cb-cli pull -all`

`cb-cli pull -all-services`

`cb-cli pull -collection=MyCollection`

`cb-cli pull -bucket-set-files=MyBucketSet`

`cb-cli pull -bucket-set-files=MyBucketSet -box=inbox`

`cb-cli pull -bucket-set-files=MyBucketSet -box=inbox -file=relative/path/from/inbox`

`cb-cli pull -file-store-files=MyFileStore`

`cb-cli pull -file-store-files=MyBucketSet -file-store-file=path/within/my/file/store.txt`

## Remote

**cb-cli remote**: Manage multiple remote endpoints

### Synopsis

```
Usage:
NAME:
   cb-cli remote - manage remotes

USAGE:
   cb-cli remote command [command options] [arguments...]

COMMANDS:
   list, ls     List remotes
   put          Create or update remotes
   remove, rm   Remove remotes
   set-current  Set current remote
   help, h      Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

### Description

This command is used for managing multiple remote endpoints. For example, you can have a real remote (production) and a local remote (testing).

### Options

- **name**
  The remote's name, as given by cb-cli remote list

- **platform-url**
  The Platform URL to use for the remote

- **messaging-url**
  The messaging URL to use for the remote

- **dev-email**
  The developer's email with access to the Platform

- **dev-password**
  The developer's password

- **dev-token**
  The developer's token

- **system-key**
  The system to use for the remote

#### EXAMPLES

Init a new project with a custom remote name:

```
cb-cli init --name my-remote-name ...
```

List all remotes:

```
cb-cli remote list
```

Add a new remote:

```
cb-cli remote put --name other-remote --platform-url URL --messaging-url MESSAGING_URL --dev-email DEV_EMAIL --dev-password DEV_PASSWORD --system-key SYS_KEY
```

Change the remote:

```
cb-cli remote set-current --name other-remote
```

## Test

**cb-cli test**: Execute a code service, update the code service from your local file system, and send an MQTT message

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

The test command allows you to execute your code services from your local machine, update the code, and send MQTT messages.

You can combine these options on a single command line, like with pull.

### Options

- **service = < service_name >**
  Executes the selected service. If your local version has newer changes than the cloud platform, we recommend using -push flag to push new changes.

- **params = < params >**
  The payload is the code service request's parameter. This must be valid JSON.

- **topic = < topic >**
  Sends an MQTT message with the provided topic. The payload contains the MQTT message payload.

- **payload = < payload >**
  Payload for the respective service or MQTT message. If -topic is used, then the payload is the MQTT payload.

- **push**
  -push flag pushes the local version of the respective code service before executing.

#### Examples

`cb-cli test -service=MyService`

# Full Example

## Step 1: The developer uses init to create an initial directory structure

**Command:**

    $ cb-cli init

- Platform URL : - System Key : - Developer Email : - Password :

**Result:**
A new directory will be created with your system name. Inside, there will be a single file called .cbmeta.

    -/<SYSTEM_NAME>

|- .cbmeta

## Step 2: The developer exports their system locally

This process will bring all the schema and asset definitions to your local environment.

**Command:**

    $ cd <SYSTEM_NAME>
    $ cb-cli export

**Result:**
Your folder should now be filled with a structure that looks like this:

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

## Step 3: The developer modifies a local service

This step represents a typical developer activity of modifying a service. For this task, modify a service of your choice.

    function helloworld(req, resp){
    	// COMMENT CHANGED!!!!
    	resp.success("hello world: ");
    }

## Step 4: The developer views the differences

**Action:**

    $ git diff path/to/servicename.js

**Result:**

    The differences in your file will be listed.

## Step 5: The developer tests the new service

This is an optional step, but developers often want to ensure their changes work.

**Action:**

    $ cb-cli test -service=helloworld params="{'name':'Bill'}"

**Result:**

The latest service source code is uploaded to the system and then executed with the parameters passed in.

**NOTE: Testing a service will push your service to your system. This activity should be done against a test system.**

## Step 6: The developer pushes their changes back to the ClearBlade system

If the developer has diffs that have not been tested and have already been sent to the server, they can now push those changes.

**NOTE: Now is a good time to commit your code to a source control repository!**

**Action:**

    cb-cli push -service=helloworld

**Results:**

Upon completion, you will receive a list of all the files successfully pushed to the ClearBlade system.

## Step 7: Pulling the latest code

In a team environment, you can pull the latest changes that others have running on the system. You can accomplish this using:

- A source control branch
- The asset running in the system

Use the pull command to get the asset running in the system locally.

**Action:**

    cb-cli pull -service=anotherworld

**Result:**
Your file system is now updated with the latest JavaScript in the anotherworld.js file.

### Summary

The process of developing a local system continues with these above steps. Done correctly, you should include a source control repository tool and best practices for backup and change history purposes.

See the DevOps example to learn more about the DevOps lifecycle the CLI supports.

# DevOps

In addition to supporting local development on services and systems, the CLI also provides the critical capability to integrate your current DevOps processes with the ClearBlade Platform.

This describes a typical agile development team working within the ClearBlade Platform:

- Multiple developers working with their systems
- A ClearBlade system for a shared development instance
- A ClearBlade system for testing and QA
- A ClearBlade system for production

ClearBlade expects a mature use of a source control environment where developers can work in shared isolation. - git, svn, cvs, and others are supported with the same pattern.

## Task: The developer works in the local branch.

The developer will use their ClearBlade system and the standard local development process during this period.

The developer will commit their work to a source control environment's development branch when moving forward with a feature.

## Task: Promotion of a feature branch to development

With the development source control branch updated, a current DevOps build system can take control. This involves the following steps:

1.  A build system checks out the latest version of the development branch:

    \$ git pull origin development

2.  The build system ensures they are init'd into the correct development system:

    \$ cb-cli init

3.  The build system promotes any code that is now in the stream:

    \$ cb-cli push

At this point, the development system is running the latest code captured in source control.

### Summary

DevOps tools can leverage this process to continue the promotion and rollback of any environment.
