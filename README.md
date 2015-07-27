GCM Playground
============

A reference implementation of a [GCM Application Server](https://developers.google.com/cloud-messaging/server#role) in the form of a playground that  developers can use to test the Google Cloud Messaging Service.


Introduction
------------

GCM Playground is a basic web UI for sending and receiving messages to give developers a feel for the GCM API. The playground can be used as a sample implementation of the GCM App Server, that developers can use as a reference for their own app server implementation.


Pre-requisites
--------------

- [Google Cloud Messaging](https://developers.google.com/cloud-messaging/gcm)


Getting Started
---------------

#### Installation

- [Install Docker and Docker Compose](https://docs.docker.com/compose/install/).
- Install Node.js >=0.12.0.
- Clone this repo.
- `$ ./start.sh`

#### Accessing services

If using boot2docker, run `$ boot2docker ip` to find out the VM IP address. Usually, it's `192.168.59.103`.

The ports that are being used are:

- **`3000` - Playground Web UI**
- `8080` - Playground server
- `5601` - Kibana web interface
- `9200` - Elasticsearch JSON interface
- `5000` - Logstash server, receives logs from logstash forwarders


Chrome App
-----------

Included in `chrome/` is a Chrome app that can help you get started with the playground. To start using it, you'll need to get the Project ID from your Google Cloud Console, and have the GCM Playground running.

The Chrome app lets you:
- Register the client with the backend
- Receive messages sent through the playground


Support
-------

- Stack Overflow: http://stackoverflow.com/questions/tagged/google-cloud-messaging

If you've found an error in this sample, please file an issue: https://github.com/googlesamples/gcm-playground/issues

Patches are encouraged, and may be submitted by forking this project and submitting a pull request through GitHub.


License
-------

Copyright 2015 Google, Inc.

Licensed to the Apache Software Foundation (ASF) under one or more contributor
license agreements.  See the NOTICE file distributed with this work for
additional information regarding copyright ownership.  The ASF licenses this
file to you under the Apache License, Version 2.0 (the "License"); you may not
use this file except in compliance with the License.  You may obtain a copy of
the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  See the
License for the specific language governing permissions and limitations under
the License.
